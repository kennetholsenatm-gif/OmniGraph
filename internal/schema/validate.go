package schema

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/kennetholsenatm-gif/omnigraph/schemas"
	"github.com/pelletier/go-toml/v2"
	"github.com/santhosh-tekuri/jsonschema/v6"
	"gopkg.in/yaml.v3"
)

// ValidateOmniGraph checks instance (JSON-like map) against the embedded OmniGraph JSON Schema.
func ValidateOmniGraph(instance map[string]any) error {
	if instance == nil {
		return fmt.Errorf("nil document")
	}
	// Round-trip through JSON so YAML-decoded integers/slices match what the schema expects (avoids jsonschema panics on mixed types).
	norm, err := json.Marshal(instance)
	if err != nil {
		return fmt.Errorf("normalize document: %w", err)
	}
	var normalized map[string]any
	if err := json.Unmarshal(norm, &normalized); err != nil {
		return fmt.Errorf("normalize document: %w", err)
	}
	var schemaDoc map[string]any
	if err := json.Unmarshal(schemas.OmniGraphSchemaJSON, &schemaDoc); err != nil {
		return fmt.Errorf("parse embedded omnigraph schema: %w", err)
	}
	c := jsonschema.NewCompiler()
	const resID = "https://github.com/kennetholsenatm-gif/omnigraph/schemas/omnigraph.schema.json"
	if err := c.AddResource(resID, schemaDoc); err != nil {
		return fmt.Errorf("load schema resource: %w", err)
	}
	sch, err := c.Compile(resID)
	if err != nil {
		return fmt.Errorf("compile schema: %w", err)
	}
	if err := sch.Validate(normalized); err != nil {
		return fmt.Errorf("validate: %w", err)
	}
	return nil
}

// ValidateRawDocument validates raw JSON, YAML, or TOML bytes by decoding to a generic map first.
// It accepts the same encodings as the Schema Contract tab; TOML is the recommended human encoding.
// Downstream machine artifacts (for example omnigraph/graph/v1) remain JSON per their contracts.
// For non-JSON payloads, YAML is tried first (backward compatible with existing .omnigraph.schema files);
// if YAML fails, TOML is attempted.
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
			if err2 := toml.Unmarshal(trim, &instance); err2 != nil {
				return nil, fmt.Errorf("parse yaml: %w; parse toml: %w", err, err2)
			}
		}
	}
	if err := ValidateOmniGraph(instance); err != nil {
		return instance, err
	}
	return instance, nil
}
