package graph

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
)

type blastFixtureCase struct {
	Name       string          `json:"name"`
	Incidents  []string        `json:"incidents"`
	Downstream []string        `json:"downstream"`
	Upstream   []string        `json:"upstream"`
	Doc        json.RawMessage `json:"doc"`
}

type blastFixtureFile struct {
	Cases []blastFixtureCase `json:"cases"`
}

func TestBlast_fixturesJSON(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	dir := filepath.Dir(file)
	root := filepath.Clean(filepath.Join(dir, "..", ".."))
	path := filepath.Join(root, "testdata", "graph", "blast_fixtures.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixtures: %v", err)
	}
	var f blastFixtureFile
	if err := json.Unmarshal(raw, &f); err != nil {
		t.Fatalf("decode: %v", err)
	}
	for _, c := range f.Cases {
		t.Run(c.Name, func(t *testing.T) {
			doc, err := ParseDocument(c.Doc)
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			down, err := DownstreamBlast(doc, c.Incidents)
			if err != nil {
				t.Fatalf("DownstreamBlast: %v", err)
			}
			if !slices.Equal(down.DownstreamNodeIDs, c.Downstream) {
				t.Fatalf("downstream got %v want %v", down.DownstreamNodeIDs, c.Downstream)
			}
			up, err := UpstreamBlast(doc, c.Incidents)
			if err != nil {
				t.Fatalf("UpstreamBlast: %v", err)
			}
			if !slices.Equal(up.UpstreamNodeIDs, c.Upstream) {
				t.Fatalf("upstream got %v want %v", up.UpstreamNodeIDs, c.Upstream)
			}
		})
	}
}

func TestUpstreamBlast_fromLeaf(t *testing.T) {
	raw := []byte(`{
  "apiVersion": "omnigraph/graph/v1",
  "kind": "Graph",
  "metadata": { "generatedAt": "2026-01-01T00:00:00Z" },
  "spec": {
    "phase": "plan",
    "nodes": [
      { "id": "a", "kind": "x", "label": "A" },
      { "id": "b", "kind": "x", "label": "B" },
      { "id": "c", "kind": "x", "label": "C" }
    ],
    "edges": [
      { "from": "a", "to": "b" },
      { "from": "b", "to": "c" }
    ]
  }
}`)
	doc, err := ParseDocument(raw)
	if err != nil {
		t.Fatal(err)
	}
	up, err := UpstreamBlast(doc, []string{"c"})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"a", "b", "c"}
	if !slices.Equal(up.UpstreamNodeIDs, want) {
		t.Fatalf("got %v want %v", up.UpstreamNodeIDs, want)
	}
}

func TestValidateDocument_invalidDependencyRole(t *testing.T) {
	raw := []byte(`{
  "apiVersion": "omnigraph/graph/v1",
  "kind": "Graph",
  "metadata": { "generatedAt": "2026-01-01T00:00:00Z" },
  "spec": {
    "phase": "plan",
    "nodes": [
      { "id": "a", "kind": "x", "label": "A" },
      { "id": "b", "kind": "x", "label": "B" }
    ],
    "edges": [
      { "from": "a", "to": "b", "dependencyRole": "maybe" }
    ]
  }
}`)
	_, err := ParseDocument(raw)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrInvalidDependencyRole) && !strings.Contains(err.Error(), "dependencyRole") {
		t.Fatalf("expected invalid dependencyRole error, got %v", err)
	}
}
