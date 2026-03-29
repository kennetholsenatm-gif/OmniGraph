package omnistate

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kennetholsenatm-gif/omnigraph/internal/state"
)

// TerraformNormalizer parses OpenTofu/Terraform JSON state into fragments.
type TerraformNormalizer struct{}

func (TerraformNormalizer) Kind() SourceKind { return SourceTerraformState }

// Normalize decodes state JSON and emits one node per resource address plus depends_on edges.
func (TerraformNormalizer) Normalize(ctx context.Context, in NormalizerInput) (OmniGraphStateFragment, error) {
	_ = ctx
	var fr OmniGraphStateFragment
	if len(in.Data) == 0 {
		return fr, fmt.Errorf("empty payload")
	}
	st, err := state.Parse(in.Data)
	if err != nil {
		fr.PartialErrors = append(fr.PartialErrors, NormalizeError{
			Path:    in.Name,
			Code:    "E_TF_DECODE",
			Message: err.Error(),
		})
		return fr, nil
	}
	if st.Values == nil {
		return fr, nil
	}
	ref := in.Ref
	if ref.Type == "" {
		ref.Type = SourceTerraformState
	}
	if ref.Name == "" {
		ref.Name = in.Name
	}
	walkRootModule(st.Values.RootModule, ref, &fr)
	return fr, nil
}

func walkRootModule(rm *state.RootModule, ref SourceRef, fr *OmniGraphStateFragment) {
	if rm == nil {
		return
	}
	for i := range rm.Resources {
		addTFResource(&rm.Resources[i], ref, fr)
	}
	for i := range rm.ChildModules {
		walkChildModule(&rm.ChildModules[i], ref, fr)
	}
}

func walkChildModule(cm *state.ChildModule, ref SourceRef, fr *OmniGraphStateFragment) {
	if cm == nil {
		return
	}
	for i := range cm.Resources {
		addTFResource(&cm.Resources[i], ref, fr)
	}
	for i := range cm.ChildModules {
		walkChildModule(&cm.ChildModules[i], ref, fr)
	}
}

func addTFResource(res *state.Resource, ref SourceRef, fr *OmniGraphStateFragment) {
	if res == nil || res.Address == "" {
		return
	}
	id := "tf:" + res.Address
	attrs := map[string]any{
		"tfAddress": res.Address,
		"tfMode":    res.Mode,
		"tfType":    res.Type,
		"tfName":    res.Name,
	}
	for k, v := range res.Values {
		attrs[k] = v
	}
	fr.Nodes = append(fr.Nodes, StateNode{
		ID:         id,
		Kind:       "tf_resource",
		Label:      res.Address,
		State:      res.Mode,
		Attributes: attrs,
		Provenance: ref,
	})
	for _, dep := range dependsOnList(res.Values) {
		dep = strings.TrimSpace(dep)
		if dep == "" {
			continue
		}
		toID := "tf:" + dep
		fr.Edges = append(fr.Edges, StateEdge{
			From:       id,
			To:         toID,
			Kind:       "depends_on",
			Provenance: ref,
		})
	}
}

func dependsOnList(values map[string]any) []string {
	if values == nil {
		return nil
	}
	raw, ok := values["depends_on"]
	if !ok || raw == nil {
		return nil
	}
	switch v := raw.(type) {
	case []string:
		return v
	case []any:
		var out []string
		for _, x := range v {
			if s, ok := x.(string); ok {
				out = append(out, s)
			}
		}
		return out
	default:
		return nil
	}
}

// MarshalTerraformState re-encodes for mutation round-trips (best-effort).
func MarshalTerraformState(st *state.TerraformState) ([]byte, error) {
	if st == nil {
		return nil, fmt.Errorf("nil state")
	}
	return json.MarshalIndent(st, "", "  ")
}
