package graph

import (
	"fmt"
	"slices"
	"sort"
	"sync"
)

// ConcurrentGraph is a goroutine-safe mutable graph keyed by node ID.
// For JSON emission, call Snapshot or SnapshotDocument; Document/GraphSpec values remain immutable snapshots.
// inDeg/outDeg track incident edge counts for O(1) lookup; they update only when a non-duplicate edge is stored.
type ConcurrentGraph struct {
	mu       sync.RWMutex
	nodes    map[string]Node
	edges    []Edge
	edgeSeen map[edgePair]struct{}
	phase    string
	inDeg    map[string]int
	outDeg   map[string]int
}

// NewConcurrentGraph returns an empty graph with phase "plan".
func NewConcurrentGraph() *ConcurrentGraph {
	return &ConcurrentGraph{
		nodes:    make(map[string]Node),
		edgeSeen: make(map[edgePair]struct{}),
		inDeg:    make(map[string]int),
		outDeg:   make(map[string]int),
		phase:    "plan",
	}
}

func cloneNodeAttributes(src map[string]any) map[string]any {
	if src == nil {
		return nil
	}
	dst := make(map[string]any, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func (g *ConcurrentGraph) initDegreeMapsLocked() {
	if g.inDeg == nil {
		g.inDeg = make(map[string]int)
	}
	if g.outDeg == nil {
		g.outDeg = make(map[string]int)
	}
}

func (g *ConcurrentGraph) ensureNodeDegreesLocked(id string) {
	g.initDegreeMapsLocked()
	if _, ok := g.inDeg[id]; !ok {
		g.inDeg[id] = 0
		g.outDeg[id] = 0
	}
}

// SetPhase updates the phase string included in Snapshot (default "plan").
func (g *ConcurrentGraph) SetPhase(phase string) {
	if g == nil {
		return
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	g.phase = phase
}

// UpsertNode stores a copy of n keyed by n.ID. Empty ID is rejected.
// The stored Attributes map is copied so callers cannot mutate internal state without holding the lock.
func (g *ConcurrentGraph) UpsertNode(n Node) error {
	if g == nil {
		return ErrNilConcurrentGraph
	}
	if n.ID == "" {
		return fmt.Errorf("%w", ErrConcurrentEmptyNodeID)
	}
	store := Node{
		ID:         n.ID,
		Kind:       n.Kind,
		Label:      n.Label,
		State:      n.State,
		Attributes: cloneNodeAttributes(n.Attributes),
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.nodes == nil {
		g.nodes = make(map[string]Node)
	}
	_, existed := g.nodes[n.ID]
	g.nodes[n.ID] = store
	if !existed {
		g.ensureNodeDegreesLocked(n.ID)
	}
	return nil
}

// AddEdge appends an edge if From and To exist and (From, To) is not already present (same dedupe as TopologicalOrder).
func (g *ConcurrentGraph) AddEdge(e Edge) error {
	if g == nil {
		return ErrNilConcurrentGraph
	}
	if e.From == "" || e.To == "" {
		return fmt.Errorf("%w", ErrConcurrentEmptyEdgeEnds)
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.nodes == nil {
		return fmt.Errorf("concurrent graph: %w", &UnknownNodeError{ID: e.From})
	}
	if _, ok := g.nodes[e.From]; !ok {
		return fmt.Errorf("concurrent graph: %w", &UnknownNodeError{ID: e.From})
	}
	if _, ok := g.nodes[e.To]; !ok {
		return fmt.Errorf("concurrent graph: %w", &UnknownNodeError{ID: e.To})
	}
	if g.edgeSeen == nil {
		g.edgeSeen = make(map[edgePair]struct{})
	}
	p := edgePair{from: e.From, to: e.To}
	if _, dup := g.edgeSeen[p]; dup {
		return nil
	}
	g.edgeSeen[p] = struct{}{}
	g.edges = append(g.edges, Edge{From: e.From, To: e.To, Kind: e.Kind, DependencyRole: e.DependencyRole})
	g.ensureNodeDegreesLocked(e.From)
	g.ensureNodeDegreesLocked(e.To)
	g.inDeg[e.To]++
	g.outDeg[e.From]++
	return nil
}

// BatchUpsertNodes validates all node IDs (non-empty) without locking, then applies all under one lock.
// On validation failure nothing is mutated. Attributes maps are copied like UpsertNode.
func (g *ConcurrentGraph) BatchUpsertNodes(nodes []Node) error {
	if g == nil {
		return ErrNilConcurrentGraph
	}
	for i := range nodes {
		if nodes[i].ID == "" {
			return fmt.Errorf("%w", ErrConcurrentEmptyNodeID)
		}
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.nodes == nil {
		g.nodes = make(map[string]Node, len(nodes))
	}
	for _, n := range nodes {
		_, existed := g.nodes[n.ID]
		g.nodes[n.ID] = Node{
			ID:         n.ID,
			Kind:       n.Kind,
			Label:      n.Label,
			State:      n.State,
			Attributes: cloneNodeAttributes(n.Attributes),
		}
		if !existed {
			g.ensureNodeDegreesLocked(n.ID)
		}
	}
	return nil
}

// BatchAddEdges adds edges under one lock with the same rules as AddEdge (endpoints must exist, dedupe).
func (g *ConcurrentGraph) BatchAddEdges(edges []Edge) error {
	if g == nil {
		return ErrNilConcurrentGraph
	}
	for i := range edges {
		if edges[i].From == "" || edges[i].To == "" {
			return fmt.Errorf("%w", ErrConcurrentEmptyEdgeEnds)
		}
	}
	if len(edges) == 0 {
		return nil
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.nodes == nil {
		return fmt.Errorf("concurrent graph: %w", &UnknownNodeError{ID: edges[0].From})
	}
	for _, e := range edges {
		if _, ok := g.nodes[e.From]; !ok {
			return fmt.Errorf("concurrent graph: %w", &UnknownNodeError{ID: e.From})
		}
		if _, ok := g.nodes[e.To]; !ok {
			return fmt.Errorf("concurrent graph: %w", &UnknownNodeError{ID: e.To})
		}
	}
	if g.edgeSeen == nil {
		g.edgeSeen = make(map[edgePair]struct{})
	}
	g.edges = slices.Grow(g.edges, len(edges))
	for _, e := range edges {
		p := edgePair{from: e.From, to: e.To}
		if _, dup := g.edgeSeen[p]; dup {
			continue
		}
		g.edgeSeen[p] = struct{}{}
		g.edges = append(g.edges, Edge{From: e.From, To: e.To, Kind: e.Kind, DependencyRole: e.DependencyRole})
		g.ensureNodeDegreesLocked(e.From)
		g.ensureNodeDegreesLocked(e.To)
		g.inDeg[e.To]++
		g.outDeg[e.From]++
	}
	return nil
}

// InDegree returns the number of incoming edges (deduped by endpoint pair) for an existing node.
func (g *ConcurrentGraph) InDegree(id string) (int, bool) {
	if g == nil {
		return 0, false
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.nodes == nil {
		return 0, false
	}
	if _, ok := g.nodes[id]; !ok {
		return 0, false
	}
	if g.inDeg == nil {
		return 0, true
	}
	return g.inDeg[id], true
}

// OutDegree returns the number of outgoing edges (deduped by endpoint pair) for an existing node.
func (g *ConcurrentGraph) OutDegree(id string) (int, bool) {
	if g == nil {
		return 0, false
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.nodes == nil {
		return 0, false
	}
	if _, ok := g.nodes[id]; !ok {
		return 0, false
	}
	if g.outDeg == nil {
		return 0, true
	}
	return g.outDeg[id], true
}

// GetNode returns a copy of the node and cloned Attributes, or false if missing.
func (g *ConcurrentGraph) GetNode(id string) (Node, bool) {
	if g == nil {
		return Node{}, false
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	n, ok := g.nodes[id]
	if !ok {
		return Node{}, false
	}
	return Node{
		ID:         n.ID,
		Kind:       n.Kind,
		Label:      n.Label,
		State:      n.State,
		Attributes: cloneNodeAttributes(n.Attributes),
	}, true
}

// NodeCount returns the number of stored nodes.
func (g *ConcurrentGraph) NodeCount() int {
	if g == nil {
		return 0
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	return len(g.nodes)
}

// Snapshot returns a consistent GraphSpec: nodes sorted by ID, edges in insertion order.
func (g *ConcurrentGraph) Snapshot() GraphSpec {
	if g == nil {
		return GraphSpec{Phase: "plan"}
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.snapshotLocked()
}

func (g *ConcurrentGraph) snapshotLocked() GraphSpec {
	ids := make([]string, 0, len(g.nodes))
	for id := range g.nodes {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	nodes := make([]Node, len(ids))
	for i, id := range ids {
		n := g.nodes[id]
		nodes[i] = Node{
			ID:         n.ID,
			Kind:       n.Kind,
			Label:      n.Label,
			State:      n.State,
			Attributes: cloneNodeAttributes(n.Attributes),
		}
	}
	edges := make([]Edge, len(g.edges))
	copy(edges, g.edges)
	phase := g.phase
	if phase == "" {
		phase = "plan"
	}
	return GraphSpec{Phase: phase, Nodes: nodes, Edges: edges}
}

// SnapshotDocument builds a graph/v1 Document using Snapshot as Spec.
func (g *ConcurrentGraph) SnapshotDocument(meta Metadata) Document {
	if g == nil {
		return Document{APIVersion: apiVersion, Kind: kind, Metadata: meta, Spec: GraphSpec{Phase: "plan"}}
	}
	return Document{
		APIVersion: apiVersion,
		Kind:       kind,
		Metadata:   meta,
		Spec:       g.Snapshot(),
	}
}

// TopologicalOrder runs TopologicalOrder on a consistent snapshot (read lock held only for copy).
func (g *ConcurrentGraph) TopologicalOrder() ([]string, error) {
	if g == nil {
		return nil, ErrNilConcurrentGraph
	}
	g.mu.RLock()
	spec := g.snapshotLocked()
	g.mu.RUnlock()
	return TopologicalOrder(spec)
}
