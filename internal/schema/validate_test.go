package schema

import (
	"testing"
)

func TestValidateRawDocument_ValidYAML(t *testing.T) {
	raw := []byte(`apiVersion: omnigraph/v1alpha1
kind: Project
metadata:
  name: x
spec: {}
`)
	if _, err := ValidateRawDocument(raw); err != nil {
		t.Fatal(err)
	}
}

func TestValidateRawDocument_InvalidPortType(t *testing.T) {
	raw := []byte(`{
  "apiVersion": "omnigraph/v1alpha1",
  "kind": "Project",
  "metadata": { "name": "x" },
  "spec": {
    "network": { "publicPorts": ["80"] }
  }
}`)
	if _, err := ValidateRawDocument(raw); err == nil {
		t.Fatal("expected validation error for string port")
	}
}
