package coerce

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kennetholsenatm-gif/omnigraph/internal/project"
	"github.com/kennetholsenatm-gif/omnigraph/internal/schema"
)

func TestFromDocument_Golden(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(root, "testdata", "sample.omnigraph.schema"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := schema.ValidateRawDocument(raw); err != nil {
		t.Fatal(err)
	}
	doc, err := project.ParseProjectIntent(raw)
	if err != nil {
		t.Fatal(err)
	}
	art, err := FromDocument(doc)
	if err != nil {
		t.Fatal(err)
	}
	if art.TerraformTfvarsJSON["project_name"] != "demo" {
		t.Fatalf("project_name: %v", art.TerraformTfvarsJSON["project_name"])
	}
	if art.Env["OMNIGRAPH_PROJECT_NAME"] != "demo" {
		t.Fatalf("env name: %v", art.Env["OMNIGRAPH_PROJECT_NAME"])
	}
	if string(art.GroupVarsAllYAML) == "" {
		t.Fatal("expected group vars yaml")
	}
}
