package plan

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

// JSON is a subset of `terraform show -json` / OpenTofu plan JSON.
type JSON struct {
	PlannedValues   *PlannedValues   `json:"planned_values,omitempty"`
	ResourceChanges []ResourceChange `json:"resource_changes,omitempty"`
}

// ResourceChange mirrors the plan file resource_changes[] entry (fields used for blast seeds).
type ResourceChange struct {
	Address string `json:"address"`
	Mode    string `json:"mode"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	Change  struct {
		Actions []string `json:"actions,omitempty"`
	} `json:"change"`
}

// PlannedValues mirrors state.StateValues for resource/output shapes.
type PlannedValues struct {
	Outputs    map[string]OutputValue `json:"outputs,omitempty"`
	RootModule *RootModule            `json:"root_module,omitempty"`
}

// OutputValue wraps planned outputs.
type OutputValue struct {
	Value any `json:"value"`
}

// RootModule lists planned resources.
type RootModule struct {
	Resources []Resource `json:"resources,omitempty"`
}

// Resource is a planned resource instance.
type Resource struct {
	Address string         `json:"address"`
	Mode    string         `json:"mode"`
	Type    string         `json:"type"`
	Name    string         `json:"name"`
	Values  map[string]any `json:"values,omitempty"`
}

// Load reads plan JSON from a file (from `terraform show -json tfplan`).
func Load(path string) (*JSON, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Parse(b)
}

// Parse decodes plan JSON bytes.
func Parse(b []byte) (*JSON, error) {
	var p JSON
	if err := json.Unmarshal(b, &p); err != nil {
		return nil, fmt.Errorf("decode plan json: %w", err)
	}
	return &p, nil
}

// IsMutatingAction reports whether OpenTofu/Terraform change.actions imply a real mutation.
func IsMutatingAction(actions []string) bool {
	for _, a := range actions {
		switch strings.ToLower(strings.TrimSpace(a)) {
		case "create", "delete", "update", "replace":
			return true
		}
	}
	return false
}

// MutationAddresses returns Terraform addresses with non–no-op changes (managed resources).
func MutationAddresses(p *JSON) []string {
	if p == nil || len(p.ResourceChanges) == 0 {
		return nil
	}
	seen := make(map[string]struct{})
	var out []string
	for _, rc := range p.ResourceChanges {
		if strings.TrimSpace(rc.Address) == "" {
			continue
		}
		mode := strings.TrimSpace(rc.Mode)
		if mode != "" && mode != "managed" {
			continue
		}
		if !IsMutatingAction(rc.Change.Actions) {
			continue
		}
		if _, ok := seen[rc.Address]; ok {
			continue
		}
		seen[rc.Address] = struct{}{}
		out = append(out, rc.Address)
	}
	sort.Strings(out)
	return out
}

// MutationSeedAddresses returns addresses to use as blast seeds: resource_changes when present,
// otherwise addresses from planned_values that ProjectedHosts would surface (legacy plans).
func MutationSeedAddresses(p *JSON) []string {
	if p == nil {
		return nil
	}
	if a := MutationAddresses(p); len(a) > 0 {
		return a
	}
	hosts := ProjectedHosts(p)
	out := make([]string, 0, len(hosts))
	for k := range hosts {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// ProjectedHosts extracts the same host keys as state.ExtractHosts but from planned values.
func ProjectedHosts(p *JSON) map[string]string {
	out := make(map[string]string)
	if p == nil || p.PlannedValues == nil {
		return out
	}
	for name, ov := range p.PlannedValues.Outputs {
		if s, ok := stringify(ov.Value); ok && s != "" {
			out["output."+name] = s
		}
	}
	if p.PlannedValues.RootModule != nil {
		for _, res := range p.PlannedValues.RootModule.Resources {
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
