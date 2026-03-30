package citest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kennetholsenatm-gif/omnigraph/internal/graph"
	"github.com/kennetholsenatm-gif/omnigraph/internal/schema"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found")
		}
		dir = parent
	}
}

func TestSampleOmniGraphSchemaValidates(t *testing.T) {
	root := repoRoot(t)
	p := filepath.Join(root, "testdata", "sample.omnigraph.schema")
	raw, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := schema.ValidateRawDocument(raw); err != nil {
		t.Fatal(err)
	}
}

func TestSampleGraphEmitNonEmpty(t *testing.T) {
	root := repoRoot(t)
	raw, err := os.ReadFile(filepath.Join(root, "testdata", "sample.omnigraph.schema"))
	if err != nil {
		t.Fatal(err)
	}
	gdoc, err := graph.EmitFromProjectRaw(raw, graph.EmitFromProjectRawOptions{
		PlanJSONPath: filepath.Join(root, "internal", "plan", "testdata", "minimal-plan.json"),
		TFStatePath:  filepath.Join(root, "internal", "state", "testdata", "minimal.state.json"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if gdoc.APIVersion != "omnigraph/graph/v1" {
		t.Fatalf("apiVersion: got %q", gdoc.APIVersion)
	}
	b, err := graph.EncodeIndent(gdoc)
	if err != nil {
		t.Fatal(err)
	}
	if len(b) < 400 {
		t.Fatalf("emit output too short: %d bytes", len(b))
	}
	if !strings.Contains(string(b), `"nodes"`) {
		t.Fatal("expected nodes in graph JSON")
	}
}
