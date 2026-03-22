package coerce

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/kennetholsenatm-gif/omnigraph/internal/project"
	"gopkg.in/yaml.v3"
)

// Artifacts are in-memory representations of tool inputs (never secret material from this package).
type Artifacts struct {
	TerraformTfvarsJSON map[string]any
	GroupVarsAllYAML    []byte
	Env                 map[string]string
}

// FromDocument maps a validated document into Terraform tfvars (JSON shape), Ansible group_vars/all, and env pairs.
func FromDocument(doc *project.Document) (*Artifacts, error) {
	if doc == nil {
		return nil, fmt.Errorf("nil document")
	}
	tf := make(map[string]any)
	tf["project_name"] = doc.Metadata.Name
	if doc.Metadata.Environment != "" {
		tf["environment"] = doc.Metadata.Environment
	}
	if doc.Spec.Network != nil {
		if doc.Spec.Network.VpcCidr != "" {
			tf["vpc_cidr"] = doc.Spec.Network.VpcCidr
		}
		if len(doc.Spec.Network.PublicPorts) > 0 {
			ports := make([]any, len(doc.Spec.Network.PublicPorts))
			for i, p := range doc.Spec.Network.PublicPorts {
				ports[i] = p
			}
			tf["public_ports"] = ports
		}
	}
	if len(doc.Spec.Tags) > 0 {
		tags := make(map[string]any, len(doc.Spec.Tags))
		for k, v := range doc.Spec.Tags {
			tags[k] = v
		}
		tf["tags"] = tags
	}

	ansible := make(map[string]any)
	ansible["omnigraph_project_name"] = doc.Metadata.Name
	if doc.Metadata.Environment != "" {
		ansible["omnigraph_environment"] = doc.Metadata.Environment
	}
	if doc.Spec.Network != nil {
		if doc.Spec.Network.VpcCidr != "" {
			ansible["omnigraph_vpc_cidr"] = doc.Spec.Network.VpcCidr
		}
		if len(doc.Spec.Network.PublicPorts) > 0 {
			ansible["omnigraph_public_ports"] = doc.Spec.Network.PublicPorts
		}
	}
	if len(doc.Spec.Tags) > 0 {
		ansible["omnigraph_tags"] = doc.Spec.Tags
	}
	yb, err := yaml.Marshal(ansible)
	if err != nil {
		return nil, fmt.Errorf("marshal group_vars yaml: %w", err)
	}

	env := map[string]string{
		"OMNIGRAPH_PROJECT_NAME": doc.Metadata.Name,
	}
	if doc.Metadata.Environment != "" {
		env["OMNIGRAPH_ENVIRONMENT"] = doc.Metadata.Environment
	}
	if doc.Spec.Network != nil && doc.Spec.Network.VpcCidr != "" {
		env["OMNIGRAPH_VPC_CIDR"] = doc.Spec.Network.VpcCidr
	}
	if doc.Spec.Network != nil && len(doc.Spec.Network.PublicPorts) > 0 {
		parts := make([]string, len(doc.Spec.Network.PublicPorts))
		for i, p := range doc.Spec.Network.PublicPorts {
			parts[i] = fmt.Sprintf("%d", p)
		}
		env["OMNIGRAPH_PUBLIC_PORTS"] = strings.Join(parts, ",")
	}

	return &Artifacts{
		TerraformTfvarsJSON: tf,
		GroupVarsAllYAML:    yb,
		Env:                 env,
	}, nil
}

// FormatTerraformTfvarsJSON returns indented JSON for tfvars-compatible consumption.
func FormatTerraformTfvarsJSON(a *Artifacts) ([]byte, error) {
	if a == nil {
		return nil, fmt.Errorf("nil artifacts")
	}
	return json.MarshalIndent(a.TerraformTfvarsJSON, "", "  ")
}

// FormatEnvLines returns KEY=value lines sorted by key (suitable for process env or display).
func FormatEnvLines(a *Artifacts) string {
	if a == nil || len(a.Env) == 0 {
		return ""
	}
	keys := make([]string, 0, len(a.Env))
	for k := range a.Env {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for _, k := range keys {
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(a.Env[k])
		b.WriteByte('\n')
	}
	return b.String()
}
