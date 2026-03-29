package project

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

// Document is the typed shape of an .omnigraph.schema file after successful JSON Schema validation.
type Document struct {
	APIVersion string   `json:"apiVersion" yaml:"apiVersion"`
	Kind       string   `json:"kind" yaml:"kind"`
	Metadata   Metadata `json:"metadata" yaml:"metadata"`
	Spec       Spec     `json:"spec" yaml:"spec"`
}

// Metadata holds project identity fields.
type Metadata struct {
	Name        string `json:"name" yaml:"name"`
	Environment string `json:"environment,omitempty" yaml:"environment,omitempty"`
}

// Spec holds user-facing infrastructure intent.
type Spec struct {
	Network *NetworkSpec      `json:"network,omitempty" yaml:"network,omitempty"`
	Tags    map[string]string `json:"tags,omitempty" yaml:"tags,omitempty"`
}

// NetworkSpec describes coarse network intent coerced into toolchains.
type NetworkSpec struct {
	VpcCidr     string `json:"vpcCidr,omitempty" yaml:"vpcCidr,omitempty"`
	PublicPorts []int  `json:"publicPorts,omitempty" yaml:"publicPorts,omitempty"`
}

// ParseDocument decodes JSON, YAML, or TOML bytes into Document.
// For non-JSON payloads, YAML is tried first; if it fails, TOML is attempted.
func ParseDocument(raw []byte) (*Document, error) {
	trim := bytes.TrimSpace(raw)
	if len(trim) == 0 {
		return nil, fmt.Errorf("empty document")
	}
	var doc Document
	switch trim[0] {
	case '{', '[':
		if err := json.Unmarshal(trim, &doc); err != nil {
			return nil, fmt.Errorf("parse json: %w", err)
		}
	default:
		if err := yaml.Unmarshal(trim, &doc); err != nil {
			if err2 := toml.Unmarshal(trim, &doc); err2 != nil {
				return nil, fmt.Errorf("parse yaml: %w; parse toml: %w", err, err2)
			}
		}
	}
	return &doc, nil
}
