package schema

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/kennetholsenatm-gif/omnigraph/schemas"
	"github.com/santhosh-tekuri/jsonschema/v6"
	"gopkg.in/yaml.v3"
)

// ValidateOmniGraph checks instance (JSON-like map) against the embedded OmniGraph JSON Schema.
func ValidateOmniGraph(instance map[string]any) error {
	c := jsonschema.NewCompiler()
	const resID = "https://github.com/kennetholsenatm-gif/omnigraph/schemas/omnigraph.schema.json"
	if err := c.AddResource(resID, bytes.NewReader(schemas.OmniGraphSchemaJSON)); err != nil {
		return fmt.Errorf("load schema resource: %w", err)
	}
	sch, err := c.Compile(resID)
	if err != nil {
		return fmt.Errorf("compile schema: %w", err)
	}
	if err := sch.Validate(instance); err != nil {
		return fmt.Errorf("validate: %w", err)
	}
	return nil
}

// ValidateRawDocument validates raw JSON or YAML bytes by decoding to a generic map first.
func ValidateRawDocument(raw []byte) (map[string]any, error) {
	var instance map[string]any
	trim := bytes.TrimSpace(raw)
	if len(trim) == 0 {
		return nil, fmt.Errorf("empty document")
	}
	switch trim[0] {
	case '{', '[':
		if err := json.Unmarshal(trim, &instance); err != nil {
			return nil, fmt.Errorf("parse json: %w", err)
		}
	default:
		if err := yaml.Unmarshal(trim, &instance); err != nil {
			return nil, fmt.Errorf("parse yaml: %w", err)
		}
	}
	if err := ValidateOmniGraph(instance); err != nil {
		return instance, err
	}
	return instance, nil
}
