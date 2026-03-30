package schema

import (
	"encoding/json"
	"fmt"

	"github.com/kennetholsenatm-gif/omnigraph/schemas"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

const (
	integrationRunV1SchemaURL    = "https://github.com/kennetholsenatm-gif/omnigraph/schemas/integration-run.v1.schema.json"
	integrationResultV1SchemaURL = "https://github.com/kennetholsenatm-gif/omnigraph/schemas/integration-result.v1.schema.json"
	inventorySourceV1SchemaURL   = "https://github.com/kennetholsenatm-gif/omnigraph/schemas/inventory-source.v1.schema.json"
)

// ValidateIntegrationRunV1 validates stdin JSON for integration WASM guests.
func ValidateIntegrationRunV1(instance map[string]any) error {
	return validateAgainstEmbedded(integrationRunV1SchemaURL, schemas.IntegrationRunV1SchemaJSON, instance)
}

// ValidateIntegrationResultV1 validates stdout JSON from integration WASM guests.
// When spec.inventorySnapshot is present, it must also satisfy omnigraph/inventory-source/v1.
func ValidateIntegrationResultV1(instance map[string]any) error {
	if err := validateAgainstEmbedded(integrationResultV1SchemaURL, schemas.IntegrationResultV1SchemaJSON, instance); err != nil {
		return err
	}
	spec, ok := instance["spec"].(map[string]any)
	if !ok || spec == nil {
		return nil
	}
	raw, ok := spec["inventorySnapshot"]
	if !ok || raw == nil {
		return nil
	}
	snap, ok := raw.(map[string]any)
	if !ok {
		return fmt.Errorf("spec.inventorySnapshot: expected object")
	}
	return validateAgainstEmbedded(inventorySourceV1SchemaURL, schemas.InventorySourceV1SchemaJSON, snap)
}

func validateAgainstEmbedded(schemaURL string, schemaBytes []byte, instance map[string]any) error {
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
	if err := json.Unmarshal(schemaBytes, &schemaDoc); err != nil {
		return fmt.Errorf("parse embedded schema: %w", err)
	}
	c := jsonschema.NewCompiler()
	if err := c.AddResource(schemaURL, schemaDoc); err != nil {
		return fmt.Errorf("load schema resource: %w", err)
	}
	sch, err := c.Compile(schemaURL)
	if err != nil {
		return fmt.Errorf("compile schema: %w", err)
	}
	if err := sch.Validate(normalized); err != nil {
		return fmt.Errorf("validate: %w", err)
	}
	return nil
}
