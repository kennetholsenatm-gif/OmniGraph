package graph

import (
	"fmt"
	"sort"
	"strings"
)

// EffectiveDependencyRole returns EdgeDependencyNecessary when DependencyRole is empty or whitespace.
func EffectiveDependencyRole(e Edge) string {
	s := strings.TrimSpace(e.DependencyRole)
	if s == "" {
		return EdgeDependencyNecessary
	}
	return s
}

// IsNecessaryEdge reports whether blast-radius traversals follow this directed edge forward or backward.
func IsNecessaryEdge(e Edge) bool {
	return EffectiveDependencyRole(e) == EdgeDependencyNecessary
}

// BlastReport is the structured result of graph-relative dependency analysis for incident triage.
type BlastReport struct {
	IncidentIDs         []string
	DownstreamNodeIDs   []string
	UpstreamNodeIDs     []string
	TraversedDownstream []Edge
	TraversedUpstream   []Edge
}

// DownstreamBlast returns every node ID reachable from incidents by following outgoing **necessary** edges
// (transitive closure). Incidents are included in the returned slice. Order is deterministic (sorted BFS layers).
func DownstreamBlast(doc *Document, incidents []string) (*BlastReport, error) {
	if doc == nil {
		return nil, fmt.Errorf("%w", ErrNilDocument)
	}
	if len(incidents) == 0 {
		return &BlastReport{}, nil
	}
	nodeIDs := buildNodeIDSet(doc.Spec.Nodes)
	uniqInc := make([]string, 0, len(incidents))
	seenInc := make(map[string]struct{})
	for _, id := range incidents {
		if id == "" {
			return nil, fmt.Errorf("incident node id cannot be empty")
		}
		if _, ok := nodeIDs[id]; !ok {
			return nil, fmt.Errorf("%w", &UnknownNodeError{ID: id})
		}
		if _, dup := seenInc[id]; dup {
			continue
		}
		seenInc[id] = struct{}{}
		uniqInc = append(uniqInc, id)
	}
	incidents = uniqInc
	adj := make(map[string][]Edge)
	for _, e := range doc.Spec.Edges {
		if !IsNecessaryEdge(e) {
			continue
		}
		adj[e.From] = append(adj[e.From], e)
	}
	seen := make(map[string]struct{})
	var order []string
	var trav []Edge
	queue := append([]string(nil), incidents...)
	for _, id := range incidents {
		seen[id] = struct{}{}
	}
	for len(queue) > 0 {
		u := queue[0]
		queue = queue[1:]
		order = append(order, u)
		for _, e := range adj[u] {
			trav = append(trav, e)
			if _, ok := seen[e.To]; ok {
				continue
			}
			seen[e.To] = struct{}{}
			queue = append(queue, e.To)
		}
	}
	sort.Strings(order)
	return &BlastReport{
		IncidentIDs:         append([]string(nil), incidents...),
		DownstreamNodeIDs:   order,
		TraversedDownstream: trav,
	}, nil
}

// UpstreamBlast returns every node ID that can reach **any** incident by following **necessary** edges backward
// (i.e. who you depend on). Incidents are included. Order is deterministic.
func UpstreamBlast(doc *Document, incidents []string) (*BlastReport, error) {
	if doc == nil {
		return nil, fmt.Errorf("%w", ErrNilDocument)
	}
	if len(incidents) == 0 {
		return &BlastReport{}, nil
	}
	nodeIDs := buildNodeIDSet(doc.Spec.Nodes)
	uniqInc := make([]string, 0, len(incidents))
	seenInc := make(map[string]struct{})
	for _, id := range incidents {
		if id == "" {
			return nil, fmt.Errorf("incident node id cannot be empty")
		}
		if _, ok := nodeIDs[id]; !ok {
			return nil, fmt.Errorf("%w", &UnknownNodeError{ID: id})
		}
		if _, dup := seenInc[id]; dup {
			continue
		}
		seenInc[id] = struct{}{}
		uniqInc = append(uniqInc, id)
	}
	incidents = uniqInc
	rev := make(map[string][]Edge)
	for _, e := range doc.Spec.Edges {
		if !IsNecessaryEdge(e) {
			continue
		}
		rev[e.To] = append(rev[e.To], e)
	}
	seen := make(map[string]struct{})
	var order []string
	var trav []Edge
	queue := append([]string(nil), incidents...)
	for _, id := range incidents {
		seen[id] = struct{}{}
	}
	for len(queue) > 0 {
		u := queue[0]
		queue = queue[1:]
		order = append(order, u)
		for _, e := range rev[u] {
			trav = append(trav, e)
			if _, ok := seen[e.From]; ok {
				continue
			}
			seen[e.From] = struct{}{}
			queue = append(queue, e.From)
		}
	}
	sort.Strings(order)
	return &BlastReport{
		IncidentIDs:       append([]string(nil), incidents...),
		UpstreamNodeIDs:   order,
		TraversedUpstream: trav,
	}, nil
}
