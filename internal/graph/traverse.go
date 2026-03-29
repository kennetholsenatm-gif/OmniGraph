package graph

import (
	"fmt"
	"sort"
)

// Adjacencies builds forward (out-edges) and reverse (in-edges) adjacency lists from spec.
// Validation matches TopologicalOrder: duplicate node IDs, empty edge endpoints, and unknown
// endpoints are errors. Parallel edges with the same (From, To) are deduplicated.
// Neighbor lists are sorted by neighbor ID so BFS order is stable for a given spec.
//
// Call once per batch of traversals and reuse the maps to avoid repeated prepareTopoSpec work.
func Adjacencies(spec GraphSpec) (forward, reverse map[string][]string, err error) {
	st, err := prepareTopoSpec(spec)
	if err != nil {
		return nil, nil, err
	}
	if st.V == 0 {
		return make(map[string][]string), make(map[string][]string), nil
	}
	forward = make(map[string][]string, st.V)
	reverse = make(map[string][]string, st.V)
	for id := range st.idToIdx {
		forward[id] = nil
		reverse[id] = nil
	}
	for _, p := range st.deduped {
		forward[p.from] = append(forward[p.from], p.to)
		reverse[p.to] = append(reverse[p.to], p.from)
	}
	for id := range st.idToIdx {
		sort.Strings(forward[id])
		sort.Strings(reverse[id])
	}
	return forward, reverse, nil
}

// ForwardAdjacency returns only the forward adjacency map (see Adjacencies).
func ForwardAdjacency(spec GraphSpec) (map[string][]string, error) {
	f, _, err := Adjacencies(spec)
	return f, err
}

// ReverseAdjacency returns only the reverse adjacency map (see Adjacencies).
func ReverseAdjacency(spec GraphSpec) (map[string][]string, error) {
	_, r, err := Adjacencies(spec)
	return r, err
}

// DescendantsBFSInto appends reachable node IDs from root along directed out-edges (BFS order:
// layer by layer; within a layer, neighbors follow sorted adjacency order). The result includes root.
//
// If visited is nil, a new map is allocated. If visited is non-nil, the caller must clear it
// between traversals from different roots (or reuse only for disjoint queries); otherwise nodes
// stay marked and will be skipped.
//
// queue and out are reset with queue[:0] and out[:0] before use; re-slice them between calls to
// reuse backing arrays and reduce allocations.
func DescendantsBFSInto(adj map[string][]string, root string, visited map[string]struct{}, queue, out []string) ([]string, error) {
	if root == "" {
		return nil, fmt.Errorf("%w", &UnknownNodeError{ID: root})
	}
	if _, ok := adj[root]; !ok {
		return nil, fmt.Errorf("%w", &UnknownNodeError{ID: root})
	}
	if visited == nil {
		visited = make(map[string]struct{}, len(adj))
	}
	out = out[:0]
	queue = queue[:0]
	visited[root] = struct{}{}
	out = append(out, root)
	queue = append(queue, root)
	for qi := 0; qi < len(queue); qi++ {
		u := queue[qi]
		for _, v := range adj[u] {
			if _, ok := visited[v]; !ok {
				visited[v] = struct{}{}
				out = append(out, v)
				queue = append(queue, v)
			}
		}
	}
	return out, nil
}

// AncestorsBFSInto is like DescendantsBFSInto but follows reverse edges (all nodes that can reach
// root via a directed path). The result includes root. visited/queue/out reuse rules are the same.
func AncestorsBFSInto(rev map[string][]string, root string, visited map[string]struct{}, queue, out []string) ([]string, error) {
	if root == "" {
		return nil, fmt.Errorf("%w", &UnknownNodeError{ID: root})
	}
	if _, ok := rev[root]; !ok {
		return nil, fmt.Errorf("%w", &UnknownNodeError{ID: root})
	}
	if visited == nil {
		visited = make(map[string]struct{}, len(rev))
	}
	out = out[:0]
	queue = queue[:0]
	visited[root] = struct{}{}
	out = append(out, root)
	queue = append(queue, root)
	for qi := 0; qi < len(queue); qi++ {
		u := queue[qi]
		for _, v := range rev[u] {
			if _, ok := visited[v]; !ok {
				visited[v] = struct{}{}
				out = append(out, v)
				queue = append(queue, v)
			}
		}
	}
	return out, nil
}

// Descendants returns all nodes reachable from root along out-edges, including root (BFS order).
// It validates spec and root in one shot; for repeated queries, use Adjacencies and DescendantsBFSInto.
func Descendants(spec GraphSpec, root string) ([]string, error) {
	fwd, _, err := Adjacencies(spec)
	if err != nil {
		return nil, err
	}
	visited := make(map[string]struct{}, len(fwd))
	queue := make([]string, 0, len(fwd))
	out := make([]string, 0, len(fwd))
	return DescendantsBFSInto(fwd, root, visited, queue, out)
}

// Ancestors returns all nodes that can reach root along directed edges, including root (BFS on reverse adjacency).
func Ancestors(spec GraphSpec, root string) ([]string, error) {
	_, rev, err := Adjacencies(spec)
	if err != nil {
		return nil, err
	}
	visited := make(map[string]struct{}, len(rev))
	queue := make([]string, 0, len(rev))
	out := make([]string, 0, len(rev))
	return AncestorsBFSInto(rev, root, visited, queue, out)
}
