package repo

import (
	"encoding/json"
	"fmt"
)

// TerraformResourceAddresses returns managed resource addresses from Terraform/OpenTofu JSON state.
// Supports legacy top-level "resources" (Python agent style) and v4 "values.root_module" trees.
func TerraformResourceAddresses(raw []byte) ([]string, error) {
	var probe map[string]json.RawMessage
	if err := json.Unmarshal(raw, &probe); err != nil {
		return nil, fmt.Errorf("json: %w", err)
	}
	if r, ok := probe["resources"]; ok {
		return legacyResourceAddresses(r)
	}
	var doc struct {
		Values *struct {
			RootModule *v4Module `json:"root_module"`
		} `json:"values"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, err
	}
	if doc.Values == nil || doc.Values.RootModule == nil {
		return nil, nil
	}
	var out []string
	walkV4(doc.Values.RootModule, &out)
	return out, nil
}

type v4Module struct {
	Resources    []v4Resource `json:"resources"`
	ChildModules []*v4Module  `json:"child_modules"`
}

type v4Resource struct {
	Address string `json:"address"`
	Mode    string `json:"mode"`
	Type    string `json:"type"`
	Name    string `json:"name"`
}

func walkV4(m *v4Module, out *[]string) {
	if m == nil {
		return
	}
	for _, r := range m.Resources {
		if r.Mode == "data" {
			continue
		}
		addr := r.Address
		if addr == "" && r.Type != "" && r.Name != "" {
			addr = r.Type + "." + r.Name
		}
		if addr != "" {
			*out = append(*out, addr)
		}
	}
	for _, ch := range m.ChildModules {
		walkV4(ch, out)
	}
}

func legacyResourceAddresses(raw json.RawMessage) ([]string, error) {
	var resources []map[string]any
	if err := json.Unmarshal(raw, &resources); err != nil {
		return nil, err
	}
	var out []string
	for _, res := range resources {
		if res == nil {
			continue
		}
		if m, _ := res["mode"].(string); m == "data" {
			continue
		}
		addr := tfResourceAddressFromMap(res)
		if addr != "" {
			out = append(out, addr)
		}
	}
	return out, nil
}

func tfResourceAddressFromMap(res map[string]any) string {
	typ, _ := res["type"].(string)
	name, _ := res["name"].(string)
	if typ == "" || name == "" {
		return ""
	}
	if mod, ok := res["module"].(string); ok && mod != "" {
		return mod + "." + typ + "." + name
	}
	return typ + "." + name
}
