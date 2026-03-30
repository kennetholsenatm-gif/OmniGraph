package telemetry

// APIVersion is the document version for file-based telemetry merges into omnigraph/graph/v1.
const APIVersion = "omnigraph/telemetry/v1"

// Bundle is a small JSON fixture or webhook-derived snapshot merged into graph emit.
type Bundle struct {
	APIVersion string `json:"apiVersion"`
	Nodes      []Node `json:"nodes"`
	Edges      []Edge `json:"edges"`
}

// Node mirrors graph.Node without importing graph (avoids package cycles).
type Node struct {
	ID         string         `json:"id"`
	Kind       string         `json:"kind"`
	Label      string         `json:"label"`
	State      string         `json:"state,omitempty"`
	Attributes map[string]any `json:"attributes,omitempty"`
}

// Edge mirrors graph.Edge.
type Edge struct {
	From           string `json:"from"`
	To             string `json:"to"`
	Kind           string `json:"kind,omitempty"`
	DependencyRole string `json:"dependencyRole,omitempty"`
}
