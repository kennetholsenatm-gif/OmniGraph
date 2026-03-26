package ir

import (
	"strings"
	"testing"
)

func TestParseDocumentMinimal(t *testing.T) {
	raw := `apiVersion: omnigraph/ir/v1
kind: InfrastructureIntent
metadata:
  name: demo
spec:
  targets:
    - id: web1
      ansibleHost: "10.0.0.1"
  components:
    - id: net
      componentType: omnigraph.network.vpc
  relations: []
`
	doc, err := ParseDocument([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	if doc.Metadata.Name != "demo" {
		t.Fatalf("name %q", doc.Metadata.Name)
	}
	if len(doc.Spec.Targets) != 1 || doc.Spec.Targets[0].ID != "web1" {
		t.Fatalf("targets %+v", doc.Spec.Targets)
	}
}

func TestParseDocumentRejectWrongVersion(t *testing.T) {
	raw := `{"apiVersion":"wrong","kind":"InfrastructureIntent","metadata":{"name":"x"},"spec":{"targets":[],"components":[],"relations":[]}}`
	_, err := ParseDocument([]byte(raw))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "validate") {
		t.Fatalf("unexpected err: %v", err)
	}
}
