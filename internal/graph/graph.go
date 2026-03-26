package graph

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/kennetholsenatm-gif/omnigraph/internal/coerce"
	"github.com/kennetholsenatm-gif/omnigraph/internal/plan"
	"github.com/kennetholsenatm-gif/omnigraph/internal/project"
	"github.com/kennetholsenatm-gif/omnigraph/internal/state"
	"github.com/kennetholsenatm-gif/omnigraph/internal/telemetry"
)

const apiVersion = "omnigraph/graph/v1"
const kind = "Graph"

// Document is the versioned graph payload for UI and PR comments.
type Document struct {
	APIVersion string    `json:"apiVersion"`
	Kind       string    `json:"kind"`
	Metadata   Metadata  `json:"metadata"`
	Spec       GraphSpec `json:"spec"`
}

// Metadata identifies the graph emission.
type Metadata struct {
	GeneratedAt string `json:"generatedAt"`
	Project     string `json:"project,omitempty"`
	Environment string `json:"environment,omitempty"`
}

// GraphSpec holds nodes and edges for the visualizer.
type GraphSpec struct {
	Phase   string      `json:"phase"`
	Nodes   []Node      `json:"nodes"`
	Edges   []Edge      `json:"edges"`
	Phases  []PhaseInfo `json:"phases,omitempty"`
	Summary *RunSummary `json:"summary,omitempty"`
}

// Node is a vertex in the dependency / topology graph.
type Node struct {
	ID         string         `json:"id"`
	Kind       string         `json:"kind"`
	Label      string         `json:"label"`
	State      string         `json:"state,omitempty"`
	Attributes map[string]any `json:"attributes,omitempty"`
}

// Edge links two nodes.
type Edge struct {
	From string `json:"from"`
	To   string `json:"to"`
	Kind string `json:"kind,omitempty"`
}

// PhaseInfo records lifecycle progress.
type PhaseInfo struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Detail string `json:"detail,omitempty"`
}

// RunSummary captures coarse tool outcomes (for PR annotations).
type RunSummary struct {
	ValidateOK bool   `json:"validateOk"`
	CoerceOK   bool   `json:"coerceOk"`
	Inventory  string `json:"inventoryPreview,omitempty"`
}

// EmitOptions configures optional plan/state enrichment.
type EmitOptions struct {
	PlanJSONPath   string
	TerraformState *state.TerraformState
	// TelemetryPath loads omnigraph/telemetry/v1 JSON and merges nodes/edges (see internal/telemetry).
	TelemetryPath string
}

// Emit builds a Graph v1 document from a validated project document and coercion artifacts.
func Emit(doc *project.Document, art *coerce.Artifacts, opts EmitOptions) (*Document, error) {
	if doc == nil {
		return nil, fmt.Errorf("nil project document")
	}
	now := time.Now().UTC().Format(time.RFC3339)
	g := &Document{
		APIVersion: apiVersion,
		Kind:       kind,
		Metadata: Metadata{
			GeneratedAt: now,
			Project:     doc.Metadata.Name,
			Environment: doc.Metadata.Environment,
		},
		Spec: GraphSpec{
			Phase: "plan",
			Nodes: []Node{
				{ID: "project", Kind: "project", Label: doc.Metadata.Name, State: "active"},
				{ID: "tf", Kind: "tool", Label: "OpenTofu/Terraform", State: "pending"},
				{ID: "ansible", Kind: "tool", Label: "Ansible", State: "pending"},
			},
			Edges: []Edge{
				{From: "project", To: "tf", Kind: "provisions"},
				{From: "tf", To: "ansible", Kind: "configures"},
			},
			Phases: []PhaseInfo{
				{Name: "validate", Status: "ok"},
				{Name: "coerce", Status: "ok"},
				{Name: "plan", Status: "pending"},
				{Name: "apply", Status: "pending"},
			},
			Summary: &RunSummary{ValidateOK: true, CoerceOK: art != nil},
		},
	}
	if opts.PlanJSONPath != "" {
		pj, err := plan.Load(opts.PlanJSONPath)
		if err != nil {
			return nil, err
		}
		hosts := plan.ProjectedHosts(pj)
		for _, addr := range sortedStringKeys(hosts) {
			ip := hosts[addr]
			id := "planned-" + slug(addr)
			g.Spec.Nodes = append(g.Spec.Nodes, Node{
				ID:    id,
				Kind:  "host",
				Label: addr,
				State: "planned",
				Attributes: map[string]any{
					"ansible_host": ip,
				},
			})
			g.Spec.Edges = append(g.Spec.Edges, Edge{From: "tf", To: id, Kind: "creates"})
		}
	}
	if opts.TerraformState != nil {
		hosts := state.ExtractHosts(opts.TerraformState)
		for _, addr := range sortedStringKeys(hosts) {
			ip := hosts[addr]
			id := "live-" + slug(addr)
			g.Spec.Nodes = append(g.Spec.Nodes, Node{
				ID:    id,
				Kind:  "host",
				Label: addr,
				State: "live",
				Attributes: map[string]any{
					"ansible_host": ip,
				},
			})
			g.Spec.Edges = append(g.Spec.Edges, Edge{From: "tf", To: id, Kind: "managed"})
		}
	}
	if opts.TelemetryPath != "" {
		bun, err := telemetry.LoadBundle(opts.TelemetryPath)
		if err != nil {
			return nil, err
		}
		mergeTelemetry(g, bun)
	}
	return g, nil
}

func mergeTelemetry(d *Document, b *telemetry.Bundle) {
	if d == nil || b == nil {
		return
	}
	seen := make(map[string]struct{}, len(d.Spec.Nodes))
	for _, n := range d.Spec.Nodes {
		seen[n.ID] = struct{}{}
	}
	for _, n := range b.Nodes {
		if n.ID == "" {
			continue
		}
		if _, ok := seen[n.ID]; ok {
			continue
		}
		seen[n.ID] = struct{}{}
		d.Spec.Nodes = append(d.Spec.Nodes, Node{
			ID:         n.ID,
			Kind:       n.Kind,
			Label:      n.Label,
			State:      n.State,
			Attributes: n.Attributes,
		})
	}
	for _, e := range b.Edges {
		if e.From == "" || e.To == "" {
			continue
		}
		d.Spec.Edges = append(d.Spec.Edges, Edge{From: e.From, To: e.To, Kind: e.Kind})
	}
}

func sortedStringKeys(m map[string]string) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	return ks
}

func truncateString(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

func slug(s string) string {
	b := make([]rune, 0, len(s))
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			b = append(b, r)
		default:
			b = append(b, '_')
		}
	}
	if len(b) == 0 {
		return "host"
	}
	return string(b)
}

// EncodeIndent returns indented JSON for human-readable artifacts.
func EncodeIndent(g *Document) ([]byte, error) {
	return json.MarshalIndent(g, "", "  ")
}
