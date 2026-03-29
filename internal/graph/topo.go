package graph

import (
	"fmt"
	"sort"
	"strings"
)

type edgePair struct {
	from string
	to   string
}

const (
	colorUnseen  byte = 0
	colorOnStack byte = 1
	colorDone    byte = 2
)

// CycleError reports a directed cycle found while ordering the graph.
// Path is a closed walk of node IDs: [a, b, c, a] means a→b→c→a.
// errors.Is(err, ErrCycle) is true for *CycleError values.
type CycleError struct {
	Path []string
}

func (e *CycleError) Error() string {
	if e == nil || len(e.Path) == 0 {
		return "topological order: directed cycle"
	}
	return "topological order: directed cycle: " + strings.Join(e.Path, " -> ")
}

// Unwrap supports errors.Is(err, ErrCycle) for *CycleError.
func (e *CycleError) Unwrap() error {
	if e == nil {
		return nil
	}
	return ErrCycle
}

// topoPrepared holds validated node indexing and deduplicated edges for topology algorithms.
type topoPrepared struct {
	V       int
	idToIdx map[string]int
	idxToID []string
	deduped []edgePair
}

func prepareTopoSpec(spec GraphSpec) (*topoPrepared, error) {
	nodes := spec.Nodes
	edges := spec.Edges
	V := len(nodes)
	if V == 0 {
		return &topoPrepared{}, nil
	}

	idToIdx := make(map[string]int, V)
	idxToID := make([]string, V)
	for i, n := range nodes {
		if n.ID == "" {
			return nil, &TopoNodeError{Index: i, Err: ErrEmptyNodeID}
		}
		if _, dup := idToIdx[n.ID]; dup {
			return nil, &TopoDuplicateNodeIDError{ID: n.ID}
		}
		idToIdx[n.ID] = i
		idxToID[i] = n.ID
	}

	deduped := make([]edgePair, 0, len(edges))
	seen := make(map[edgePair]struct{}, len(edges))
	for i, e := range edges {
		if e.From == "" || e.To == "" {
			if e.From == "" {
				return nil, &TopoEdgeError{Index: i, Err: ErrEmptyEdgeFrom}
			}
			return nil, &TopoEdgeError{Index: i, Err: ErrEmptyEdgeTo}
		}
		p := edgePair{from: e.From, to: e.To}
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		deduped = append(deduped, p)
	}

	for _, p := range deduped {
		if _, ok := idToIdx[p.from]; !ok {
			return nil, &UnknownNodeError{ID: p.from}
		}
		if _, ok := idToIdx[p.to]; !ok {
			return nil, &UnknownNodeError{ID: p.to}
		}
	}

	return &topoPrepared{V: V, idToIdx: idToIdx, idxToID: idxToID, deduped: deduped}, nil
}

// TopologicalOrder returns node IDs in an order where every edge From→To has From before To.
// It uses Kahn's algorithm in O(V+E) time with dense indices and pre-sized adjacency lists.
// Parallel edges with the same (From, To) are deduplicated. Unknown node references return *UnknownNodeError.
// A cycle (including a self-loop) yields a *CycleError with Path set when it can be recovered.
//
// Multiple weakly disconnected components are valid: this returns one linear extension of the
// partial order spanning all nodes. Relative order between components is not constrained by edges
// (only by the queue processing of zero-in-degree starts). For explicit per-component orders
// suitable for parallel execution, use TopologicalOrdersPerWeakComponent.
func TopologicalOrder(spec GraphSpec) ([]string, error) {
	st, err := prepareTopoSpec(spec)
	if err != nil {
		return nil, err
	}
	if st.V == 0 {
		return nil, nil
	}
	return kahnTopo(st)
}

// TopologicalOrdersPerWeakComponent partitions nodes by weak connectivity (treating edges as
// undirected), sorts components by minimum node ID, then runs TopologicalOrder on each induced
// subgraph. Isolated vertices are single-node components. Each inner slice is a valid topo order
// for that component only; there is no edge between components in the original graph.
func TopologicalOrdersPerWeakComponent(spec GraphSpec) ([][]string, error) {
	st, err := prepareTopoSpec(spec)
	if err != nil {
		return nil, err
	}
	if st.V == 0 {
		return nil, nil
	}

	parent := make([]int, st.V)
	for i := range parent {
		parent[i] = i
	}
	var find func(int) int
	find = func(x int) int {
		if parent[x] != x {
			parent[x] = find(parent[x])
		}
		return parent[x]
	}
	union := func(a, b int) {
		ra, rb := find(a), find(b)
		if ra != rb {
			if ra < rb {
				parent[rb] = ra
			} else {
				parent[ra] = rb
			}
		}
	}
	for _, p := range st.deduped {
		u := st.idToIdx[p.from]
		v := st.idToIdx[p.to]
		union(u, v)
	}

	groups := make(map[int][]int, st.V)
	for i := 0; i < st.V; i++ {
		r := find(i)
		groups[r] = append(groups[r], i)
	}

	type compWrap struct {
		minID string
		idxs  []int
	}
	comps := make([]compWrap, 0, len(groups))
	for _, idxs := range groups {
		ids := make([]string, len(idxs))
		for i, ix := range idxs {
			ids[i] = st.idxToID[ix]
		}
		sort.Strings(ids)
		comps = append(comps, compWrap{minID: ids[0], idxs: idxs})
	}
	sort.Slice(comps, func(i, j int) bool {
		return comps[i].minID < comps[j].minID
	})

	idSet := make(map[string]struct{}, st.V)
	dedupInComp := make(map[edgePair]struct{})
	out := make([][]string, 0, len(comps))
	for _, cw := range comps {
		clear(idSet)
		clear(dedupInComp)
		for _, ix := range cw.idxs {
			idSet[st.idxToID[ix]] = struct{}{}
		}
		subNodes := make([]Node, 0, len(cw.idxs))
		for _, id := range sortedIDsFromSet(idSet) {
			subNodes = append(subNodes, Node{ID: id, Kind: "_", Label: id})
		}
		subEdges := make([]Edge, 0, len(st.deduped))
		for _, p := range st.deduped {
			if _, ok := idSet[p.from]; !ok {
				continue
			}
			if _, ok := idSet[p.to]; !ok {
				continue
			}
			ep := edgePair{from: p.from, to: p.to}
			if _, dup := dedupInComp[ep]; dup {
				continue
			}
			dedupInComp[ep] = struct{}{}
			subEdges = append(subEdges, Edge{From: p.from, To: p.to})
		}
		order, err := TopologicalOrder(GraphSpec{Phase: spec.Phase, Nodes: subNodes, Edges: subEdges})
		if err != nil {
			return nil, err
		}
		out = append(out, order)
	}
	return out, nil
}

func sortedIDsFromSet(m map[string]struct{}) []string {
	ids := make([]string, 0, len(m))
	for id := range m {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func kahnTopo(st *topoPrepared) ([]string, error) {
	V := st.V
	idToIdx := st.idToIdx
	idxToID := st.idxToID
	deduped := st.deduped

	indeg := make([]int, V)
	outCount := make([]int, V)
	for _, p := range deduped {
		u := idToIdx[p.from]
		v := idToIdx[p.to]
		indeg[v]++
		outCount[u]++
	}

	outs := make([][]int, V)
	for i := range outs {
		outs[i] = make([]int, 0, outCount[i])
	}
	for _, p := range deduped {
		u := idToIdx[p.from]
		v := idToIdx[p.to]
		outs[u] = append(outs[u], v)
	}

	q := make([]int, 0, V)
	for i := 0; i < V; i++ {
		if indeg[i] == 0 {
			q = append(q, i)
		}
	}

	dequeued := make([]bool, V)
	order := make([]string, 0, V)
	for qi := 0; qi < len(q); qi++ {
		u := q[qi]
		dequeued[u] = true
		order = append(order, idxToID[u])
		for _, v := range outs[u] {
			indeg[v]--
			if indeg[v] == 0 {
				q = append(q, v)
			}
		}
	}

	if len(order) != V {
		start := -1
		for i := 0; i < V; i++ {
			if !dequeued[i] {
				start = i
				break
			}
		}
		if path := findDirectedCyclePath(V, outs, idxToID, start); len(path) > 0 {
			return nil, &CycleError{Path: path}
		}
		return nil, fmt.Errorf("%w: could not recover cycle path", ErrCycle)
	}
	return order, nil
}

// findDirectedCyclePath returns one directed cycle as node IDs [v0, v1, ..., vk, v0].
// preferStart is tried first; remaining unseen vertices are tried if needed. O(V+E).
func findDirectedCyclePath(V int, outs [][]int, idxToID []string, preferStart int) []string {
	color := make([]byte, V)
	var stack []int

	var dfs func(u int) []string
	dfs = func(u int) []string {
		color[u] = colorOnStack
		stack = append(stack, u)
		for _, v := range outs[u] {
			switch color[v] {
			case colorOnStack:
				for j := 0; j < len(stack); j++ {
					if stack[j] == v {
						cycleIdx := stack[j:]
						out := make([]string, len(cycleIdx)+1)
						for i, ix := range cycleIdx {
							out[i] = idxToID[ix]
						}
						out[len(cycleIdx)] = idxToID[v]
						return out
					}
				}
			case colorUnseen:
				if p := dfs(v); p != nil {
					return p
				}
			}
		}
		stack = stack[:len(stack)-1]
		color[u] = colorDone
		return nil
	}

	if preferStart >= 0 && preferStart < V {
		if p := dfs(preferStart); p != nil {
			return p
		}
	}
	for i := 0; i < V; i++ {
		if color[i] == colorUnseen {
			if p := dfs(i); p != nil {
				return p
			}
		}
	}
	return nil
}
