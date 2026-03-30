package schema

import (
	"encoding/json"
	"testing"
	"time"
)

func TestValidateIntegrationRunV1(t *testing.T) {
	raw := `{
		"apiVersion": "omnigraph/integration-run/v1",
		"kind": "IntegrationRun",
		"spec": {
			"plugin": "netbox",
			"allowedFetchPrefixes": ["https://nb.example/"]
		}
	}`
	var m map[string]any
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		t.Fatal(err)
	}
	if err := ValidateIntegrationRunV1(m); err != nil {
		t.Fatal(err)
	}
}

func TestValidateIntegrationResultV1_withInventory(t *testing.T) {
	now := time.Now().UTC().Format(time.RFC3339)
	raw := `{
		"apiVersion": "omnigraph/integration-result/v1",
		"kind": "IntegrationResult",
		"metadata": {
			"generatedAt": "` + now + `",
			"plugin": "netbox"
		},
		"spec": {
			"status": "ok",
			"errors": [],
			"inventorySnapshot": {
				"apiVersion": "omnigraph/inventory-source/v1",
				"kind": "InventorySnapshot",
				"metadata": { "generatedAt": "` + now + `", "source": "netbox" },
				"spec": { "records": [{ "id": "1", "recordType": "host", "confidence": "authoritative" }] }
			}
		}
	}`
	var m map[string]any
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		t.Fatal(err)
	}
	if err := ValidateIntegrationResultV1(m); err != nil {
		t.Fatal(err)
	}
}
