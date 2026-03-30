package emitter

// Document is omnigraph/ir/v1 InfrastructureIntent (JSON-serializable).
type Document struct {
	APIVersion string   `json:"apiVersion"`
	Kind       string   `json:"kind"`
	Metadata   Metadata `json:"metadata"`
	Spec       Spec     `json:"spec"`
}

// Metadata carries naming and labels for the intent bundle.
type Metadata struct {
	Name   string            `json:"name"`
	Labels map[string]string `json:"labels,omitempty"`
}

// Spec is the engine-neutral intent graph.
type Spec struct {
	Targets    []Target    `json:"targets"`
	Components []Component `json:"components"`
	Relations  []Relation  `json:"relations"`
	EmitHints  *EmitHints  `json:"emitHints,omitempty"`
}

// Target is an inventory-oriented endpoint.
type Target struct {
	ID          string            `json:"id"`
	AnsibleHost string            `json:"ansibleHost,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
}

// Component is an abstract building block.
type Component struct {
	ID            string         `json:"id"`
	ComponentType string         `json:"componentType"`
	Config        map[string]any `json:"config,omitempty"`
}

// Relation connects two components (or future target ids).
type Relation struct {
	From         string `json:"from"`
	To           string `json:"to"`
	RelationType string `json:"relationType"`
}

// EmitHints optionally orders backends for emit.
type EmitHints struct {
	Backends []string `json:"backends,omitempty"`
}
