package graph

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/kennetholsenatm-gif/omnigraph/internal/coerce"
	"github.com/kennetholsenatm-gif/omnigraph/internal/project"
	"github.com/kennetholsenatm-gif/omnigraph/internal/schema"
	"github.com/kennetholsenatm-gif/omnigraph/internal/state"
	"github.com/kennetholsenatm-gif/omnigraph/schemas"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

func TestEmit_ConformsToGraphSchema(t *testing.T) {
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
	doc, err := project.ParseDocument(raw)
	if err != nil {
		t.Fatal(err)
	}
	art, err := coerce.FromDocument(doc)
	if err != nil {
		t.Fatal(err)
	}
	st, err := state.Load(filepath.Join(root, "internal", "state", "testdata", "minimal.state.json"))
	if err != nil {
		t.Fatal(err)
	}
	planPath := filepath.Join(root, "internal", "plan", "testdata", "minimal-plan.json")
	gdoc, err := Emit(doc, art, EmitOptions{
		PlanJSONPath:   planPath,
		TerraformState: st,
	})
	if err != nil {
		t.Fatal(err)
	}
	b, err := EncodeIndent(gdoc)
	if err != nil {
		t.Fatal(err)
	}
	c := jsonschema.NewCompiler()
	const graphID = "https://github.com/kennetholsenatm-gif/omnigraph/schemas/graph.v1.schema.json"
	if err := c.AddResource(graphID, bytes.NewReader(schemas.GraphV1SchemaJSON)); err != nil {
		t.Fatal(err)
	}
	sch, err := c.Compile(graphID)
	if err != nil {
		t.Fatal(err)
	}
	var instance map[string]any
	if err := json.Unmarshal(b, &instance); err != nil {
		t.Fatal(err)
	}
	if err := sch.Validate(instance); err != nil {
		t.Fatal(err)
	}
}
