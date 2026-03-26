package schema

import (
	"encoding/json"
	"fmt"

	"github.com/kennetholsenatm-gif/omnigraph/schemas"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

const irV1SchemaURL = "https://github.com/kennetholsenatm-gif/omnigraph/schemas/ir.v1.schema.json"

// ValidateIRV1 checks a decoded JSON object against omnigraph/ir/v1.
func ValidateIRV1(instance map[string]any) error {
	if instance == nil {
		return fmt.Errorf("nil document")
	}
	norm, err := json.Marshal(instance)
	if err != nil {
		return fmt.Errorf("normalize document: %w", err)
	}
	var normalized map[string]any
	if err := json.Unmarshal(norm, &normalized); err != nil {
		return fmt.Errorf("normalize document: %w", err)
	}
	var schemaDoc map[string]any
	if err := json.Unmarshal(schemas.IrV1SchemaJSON, &schemaDoc); err != nil {
		return fmt.Errorf("parse embedded ir schema: %w", err)
	}
	c := jsonschema.NewCompiler()
	if err := c.AddResource(irV1SchemaURL, schemaDoc); err != nil {
		return fmt.Errorf("load schema resource: %w", err)
	}
	sch, err := c.Compile(irV1SchemaURL)
	if err != nil {
		return fmt.Errorf("compile schema: %w", err)
	}
	if err := sch.Validate(normalized); err != nil {
		return fmt.Errorf("validate: %w", err)
	}
	return nil
}
