package emitter

import (
	"strings"

	"github.com/kennetholsenatm-gif/omnigraph/internal/graph"
)

// RelationsGraphSpec builds an omnigraph/graph/v1-style GraphSpec from IR components and relations.
// Each non-empty component ID becomes a node; each relation becomes a directed edge (From, To, Kind = relationType).
// Use with graph.TopologicalOrder or graph.TopologicalOrdersPerWeakComponent; disjoint relation chains yield
// multiple weak components without error.
func RelationsGraphSpec(doc *Document) graph.GraphSpec {
	if doc == nil {
		return graph.GraphSpec{Phase: "plan"}
	}
	seen := make(map[string]struct{})
	nodes := make([]graph.Node, 0, len(doc.Spec.Components))
	for _, c := range doc.Spec.Components {
		id := strings.TrimSpace(c.ID)
		if id == "" {
			continue
		}
		if _, dup := seen[id]; dup {
			continue
		}
		seen[id] = struct{}{}
		nodes = append(nodes, graph.Node{ID: id, Kind: "component", Label: id})
	}
	edges := make([]graph.Edge, 0, len(doc.Spec.Relations))
	for _, r := range doc.Spec.Relations {
		from := strings.TrimSpace(r.From)
		to := strings.TrimSpace(r.To)
		if from == "" || to == "" {
			continue
		}
		edges = append(edges, graph.Edge{From: from, To: to, Kind: r.RelationType})
	}
	return graph.GraphSpec{Phase: "plan", Nodes: nodes, Edges: edges}
}
