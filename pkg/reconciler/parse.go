package reconciler

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/kennetholsenatm-gif/omnigraph/internal/schema"
	"gopkg.in/yaml.v3"
)

// ParseDocument decodes JSON or YAML bytes into Document after schema validation.
func ParseDocument(raw []byte) (*Document, error) {
	trim := bytes.TrimSpace(raw)
	if len(trim) == 0 {
		return nil, fmt.Errorf("reconciler: empty document")
	}
	var m map[string]any
	switch trim[0] {
	case '{', '[':
		if err := json.Unmarshal(trim, &m); err != nil {
			return nil, fmt.Errorf("reconciler: parse json: %w", err)
		}
	default:
		if err := yaml.Unmarshal(trim, &m); err != nil {
			return nil, fmt.Errorf("reconciler: parse yaml: %w", err)
		}
	}
	if err := schema.ValidateIRV1(m); err != nil {
		return nil, err
	}
	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	var doc Document
	if err := json.Unmarshal(b, &doc); err != nil {
		return nil, fmt.Errorf("reconciler: decode document: %w", err)
	}
	return &doc, nil
}
