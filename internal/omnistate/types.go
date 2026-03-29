// Package omnistate defines a unified intermediate representation for Terraform/OpenTofu state
// and Ansible inventories, independent of graph/v1 UI documents.
package omnistate

import (
	"context"
	"time"
)

const APIVersion = "omnigraph/omnistate/v1"

// SourceKind identifies the origin artifact type.
type SourceKind string

const (
	SourceTerraformState SourceKind = "terraform_state"
	SourceAnsibleYAML    SourceKind = "ansible_yaml"
	SourceAnsibleINI     SourceKind = "ansible_ini"
	// SourceAgentLocal marks synthetic nodes from an in-cluster sync agent.
	SourceAgentLocal SourceKind = "agent_local"
)

// SourceRef is client-supplied provenance (opaque name; not a server filesystem path).
type SourceRef struct {
	Type     SourceKind `json:"type"`
	Name     string     `json:"name"`
	PathHint string     `json:"pathHint,omitempty"`
}

// SourceProvenance records which source contributed nodes/edges.
type SourceProvenance struct {
	Ref       SourceRef `json:"ref"`
	NodeCount int       `json:"nodeCount"`
	EdgeCount int       `json:"edgeCount"`
}

// StateNode is a normalized vertex (stable ID: use NodeID* helpers from parsers).
type StateNode struct {
	ID          string         `json:"id"`
	Kind        string         `json:"kind"`
	Label       string         `json:"label"`
	State       string         `json:"state,omitempty"`
	Attributes  map[string]any `json:"attributes,omitempty"`
	Provenance  SourceRef      `json:"provenance"`
}

// StateEdge links two normalized node IDs.
type StateEdge struct {
	From       string         `json:"from"`
	To         string         `json:"to"`
	Kind       string         `json:"kind,omitempty"`
	Attributes map[string]any `json:"attributes,omitempty"`
	Provenance SourceRef      `json:"provenance"`
}

// NormalizeError is a non-fatal per-artifact failure during multi-file ingest.
type NormalizeError struct {
	Path    string `json:"path,omitempty"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// OmniGraphState is the merged unified view after normalization.
type OmniGraphState struct {
	APIVersion     string             `json:"apiVersion"`
	GeneratedAt    string             `json:"generatedAt"`
	CorrelationID  string             `json:"correlationId,omitempty"`
	Nodes          []StateNode        `json:"nodes"`
	Edges          []StateEdge        `json:"edges"`
	Sources        []SourceProvenance `json:"sources,omitempty"`
	PartialErrors  []NormalizeError   `json:"partialErrors,omitempty"`
	Revision       int64              `json:"revision,omitempty"`
}

// OmniGraphStateFragment is the output of a single normalizer invocation.
type OmniGraphStateFragment struct {
	Nodes         []StateNode
	Edges         []StateEdge
	PartialErrors []NormalizeError
}

// NormalizerInput carries raw artifact bytes and metadata.
type NormalizerInput struct {
	Data        []byte
	ContentType string
	Name        string
	Ref         SourceRef
}

// Normalizer maps one artifact kind into a fragment.
type Normalizer interface {
	Kind() SourceKind
	Normalize(ctx context.Context, in NormalizerInput) (OmniGraphStateFragment, error)
}

// StatePatch is a structured delta applied by the sync WebSocket hub.
type StatePatch struct {
	UpsertNodes   []StateNode `json:"upsertNodes,omitempty"`
	RemoveNodes   []string    `json:"removeNodes,omitempty"`
	UpsertEdges   []StateEdge `json:"upsertEdges,omitempty"`
	RemoveEdges   []EdgeKey   `json:"removeEdges,omitempty"`
}

// EdgeKey identifies an edge for removal.
type EdgeKey struct {
	From string `json:"from"`
	To   string `json:"to"`
	Kind string `json:"kind,omitempty"`
}

// DegradedNode describes runtime divergence from intended omnistate.
type DegradedNode struct {
	NodeID     string            `json:"nodeId"`
	Reasons    []string          `json:"reasons"`
	AttrDiffs  map[string]DiffPair `json:"attrDiffs,omitempty"`
}

// DiffPair compares intended vs runtime scalar summaries.
type DiffPair struct {
	Intended string `json:"intended"`
	Runtime  string `json:"runtime"`
}

// FracturedEdge is an intended edge whose endpoint is missing or unresolved in runtime.
type FracturedEdge struct {
	From              string `json:"from"`
	To                string `json:"to"`
	Kind              string `json:"kind,omitempty"`
	Reason            string `json:"reason"`
	TranslucentTarget string `json:"translucentTarget,omitempty"`
}

// DriftReport is the result of CompareIntendedVsRuntime.
type DriftReport struct {
	DegradedNodes   []DegradedNode   `json:"degradedNodes"`
	FracturedEdges  []FracturedEdge  `json:"fracturedEdges"`
	AnalyzedNodes   int              `json:"analyzedNodes"`
	AnalyzedEdges   int              `json:"analyzedEdges"`
}

// MergeFragments combines fragments into one state with fresh metadata.
func MergeFragments(correlationID string, frags ...OmniGraphStateFragment) OmniGraphState {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	out := OmniGraphState{
		APIVersion:    APIVersion,
		GeneratedAt:   now,
		CorrelationID: correlationID,
		Nodes:         nil,
		Edges:         nil,
		Sources:       nil,
		PartialErrors: nil,
	}
	seenN := make(map[string]struct{})
	seenE := make(map[string]struct{})
	srcCounts := make(map[string]struct {
		nodes int
		edges int
		ref   SourceRef
	})

	for _, f := range frags {
		out.PartialErrors = append(out.PartialErrors, f.PartialErrors...)
		for _, n := range f.Nodes {
			if _, ok := seenN[n.ID]; ok {
				continue
			}
			seenN[n.ID] = struct{}{}
			out.Nodes = append(out.Nodes, n)
			k := sourceKey(n.Provenance)
			sc := srcCounts[k]
			sc.nodes++
			sc.ref = n.Provenance
			srcCounts[k] = sc
		}
		for _, e := range f.Edges {
			ek := e.From + "\x00" + e.To + "\x00" + e.Kind
			if _, ok := seenE[ek]; ok {
				continue
			}
			seenE[ek] = struct{}{}
			out.Edges = append(out.Edges, e)
			k := sourceKey(e.Provenance)
			sc := srcCounts[k]
			sc.edges++
			sc.ref = e.Provenance
			srcCounts[k] = sc
		}
	}
	for _, sc := range srcCounts {
		out.Sources = append(out.Sources, SourceProvenance{
			Ref:       sc.ref,
			NodeCount: sc.nodes,
			EdgeCount: sc.edges,
		})
	}
	return out
}

func sourceKey(r SourceRef) string {
	return string(r.Type) + "\x00" + r.Name + "\x00" + r.PathHint
}

// ApplyPatch merges a patch into a copy of st (for in-memory hub).
func ApplyPatch(st OmniGraphState, p StatePatch) OmniGraphState {
	nodeIdx := make(map[string]int, len(st.Nodes))
	for i, n := range st.Nodes {
		nodeIdx[n.ID] = i
	}
	out := st
	out.Nodes = append([]StateNode(nil), st.Nodes...)
	out.Edges = append([]StateEdge(nil), stEdgesWithout(st.Edges, p.RemoveEdges)...)
	rm := make(map[string]struct{}, len(p.RemoveNodes))
	for _, id := range p.RemoveNodes {
		rm[id] = struct{}{}
	}
	if len(rm) > 0 {
		filtered := out.Nodes[:0]
		for _, n := range out.Nodes {
			if _, drop := rm[n.ID]; !drop {
				filtered = append(filtered, n)
			}
		}
		out.Nodes = filtered
		nodeIdx = make(map[string]int, len(out.Nodes))
		for i, n := range out.Nodes {
			nodeIdx[n.ID] = i
		}
	}
	for _, n := range p.UpsertNodes {
		if i, ok := nodeIdx[n.ID]; ok {
			out.Nodes[i] = n
		} else {
			nodeIdx[n.ID] = len(out.Nodes)
			out.Nodes = append(out.Nodes, n)
		}
	}
	for _, e := range p.UpsertEdges {
		out.Edges = append(out.Edges, e)
	}
	out.Revision++
	return out
}

func stEdgesWithout(edges []StateEdge, remove []EdgeKey) []StateEdge {
	if len(remove) == 0 {
		return edges
	}
	rm := make(map[string]struct{}, len(remove))
	for _, k := range remove {
		rm[k.From+"\x00"+k.To+"\x00"+k.Kind] = struct{}{}
	}
	out := edges[:0]
	for _, e := range edges {
		key := e.From + "\x00" + e.To + "\x00" + e.Kind
		if _, drop := rm[key]; !drop {
			out = append(out, e)
		}
	}
	return out
}
