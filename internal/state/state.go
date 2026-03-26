package state

import (
	"encoding/json"
	"fmt"
	"os"
)

// TerraformState is a minimal view of OpenTofu/Terraform JSON state (version 4).
type TerraformState struct {
	Values *StateValues `json:"values,omitempty"`
}

// StateValues holds outputs and root module resources.
type StateValues struct {
	Outputs    map[string]OutputValue `json:"outputs,omitempty"`
	RootModule *RootModule            `json:"root_module,omitempty"`
}

// OutputValue is a state output wrapper.
type OutputValue struct {
	Value any `json:"value"`
	Type  any `json:"type,omitempty"`
}

// RootModule lists resources at the module root.
type RootModule struct {
	Resources []Resource `json:"resources,omitempty"`
}

// Resource is a managed or data resource entry in state.
type Resource struct {
	Address string         `json:"address"`
	Mode    string         `json:"mode"`
	Type    string         `json:"type"`
	Name    string         `json:"name"`
	Values  map[string]any `json:"values,omitempty"`
}

// Load reads and decodes a .tfstate JSON file.
func Load(path string) (*TerraformState, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Parse(b)
}

// Parse decodes state JSON bytes.
func Parse(b []byte) (*TerraformState, error) {
	var st TerraformState
	if err := json.Unmarshal(b, &st); err != nil {
		return nil, fmt.Errorf("decode state: %w", err)
	}
	return &st, nil
}

// ExtractHosts collects ansible_host candidates from outputs and aws_instance public/private IPs.
func ExtractHosts(st *TerraformState) map[string]string {
	out := make(map[string]string)
	if st == nil || st.Values == nil {
		return out
	}
	for name, ov := range st.Values.Outputs {
		if s, ok := stringify(ov.Value); ok && s != "" {
			out["output."+name] = s
		}
	}
	if st.Values.RootModule != nil {
		for _, res := range st.Values.RootModule.Resources {
			if res.Mode != "managed" {
				continue
			}
			if res.Type != "aws_instance" {
				continue
			}
			host := res.Address
			if v, ok := res.Values["public_ip"]; ok {
				if s, ok := stringify(v); ok && s != "" && s != "null" {
					out[host] = s
					continue
				}
			}
			if v, ok := res.Values["private_ip"]; ok {
				if s, ok := stringify(v); ok && s != "" && s != "null" {
					out[host] = s
				}
			}
		}
	}
	return out
}

func stringify(v any) (string, bool) {
	switch t := v.(type) {
	case string:
		return t, true
	case float64:
		return fmt.Sprintf("%.0f", t), true
	case bool:
		return fmt.Sprintf("%v", t), true
	default:
		return "", false
	}
}
