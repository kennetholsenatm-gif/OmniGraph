package netbox

import (
	"encoding/json"
	"testing"
)

func TestWebhookV1_Validate(t *testing.T) {
	w := &WebhookV1{Action: "create", IP: "10.0.5.21"}
	if err := w.Validate(); err != nil {
		t.Fatal(err)
	}
	if _, err := w.MarshalJSON(); err != nil {
		t.Fatal(err)
	}
}

func TestWebhookV1_RequiresAddress(t *testing.T) {
	w := &WebhookV1{Action: "create"}
	if err := w.Validate(); err == nil {
		t.Fatal("expected error")
	}
}

func TestWebhookV1_MarshalSetsAPIVersion(t *testing.T) {
	w := &WebhookV1{Action: "upsert", IP: "1.1.1.1"}
	b, err := w.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	if m["apiVersion"] != WebhookAPIVersion {
		t.Fatalf("apiVersion: %v", m["apiVersion"])
	}
}
