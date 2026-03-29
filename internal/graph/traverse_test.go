package graph

import (
	"errors"
	"slices"
	"testing"
)

func TestDescendants_diamond(t *testing.T) {
	spec := GraphSpec{
		Phase: "plan",
		Nodes: []Node{
			{ID: "a", Kind: "x", Label: "a"},
			{ID: "b", Kind: "x", Label: "b"},
			{ID: "c", Kind: "x", Label: "c"},
			{ID: "d", Kind: "x", Label: "d"},
		},
		Edges: []Edge{
			{From: "a", To: "b"},
			{From: "a", To: "c"},
			{From: "b", To: "d"},
			{From: "c", To: "d"},
		},
	}
	got, err := Descendants(spec, "a")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"a", "b", "c", "d"}
	if !slices.Equal(got, want) {
		t.Fatalf("descendants %v want %v", got, want)
	}
}

func TestAncestors_chain(t *testing.T) {
	spec := GraphSpec{
		Phase: "plan",
		Nodes: []Node{
			{ID: "a", Kind: "x", Label: "a"},
			{ID: "b", Kind: "x", Label: "b"},
			{ID: "c", Kind: "x", Label: "c"},
		},
		Edges: []Edge{{From: "a", To: "b"}, {From: "b", To: "c"}},
	}
	got, err := Ancestors(spec, "c")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"c", "b", "a"}
	if !slices.Equal(got, want) {
		t.Fatalf("ancestors %v want %v", got, want)
	}
}

func TestDescendants_unknownRoot(t *testing.T) {
	spec := GraphSpec{
		Phase: "plan",
		Nodes: []Node{{ID: "a", Kind: "x", Label: "a"}},
		Edges: nil,
	}
	_, err := Descendants(spec, "missing")
	if err == nil {
		t.Fatal("expected error")
	}
	var u *UnknownNodeError
	if !errors.As(err, &u) {
		t.Fatalf("expected UnknownNodeError, got %T %v", err, err)
	}
}

func TestDescendantsBFSInto_reusedVisitedNeedsClear(t *testing.T) {
	spec := GraphSpec{
		Phase: "plan",
		Nodes: []Node{
			{ID: "a", Kind: "x", Label: "a"},
			{ID: "b", Kind: "x", Label: "b"},
		},
		Edges: []Edge{{From: "a", To: "b"}},
	}
	fwd, _, err := Adjacencies(spec)
	if err != nil {
		t.Fatal(err)
	}
	visited := make(map[string]struct{}, 8)
	queue := make([]string, 0, 8)
	out := make([]string, 0, 8)
	out, err = DescendantsBFSInto(fwd, "a", visited, queue, out)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(out, []string{"a", "b"}) {
		t.Fatalf("first %v", out)
	}
	// Same visited: second traversal from "a" sees everything already visited.
	out, err = DescendantsBFSInto(fwd, "a", visited, queue, out)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(out, []string{"a"}) {
		t.Fatalf("stale visited: got %v want [a]", out)
	}
	clear(visited)
	out, err = DescendantsBFSInto(fwd, "a", visited, queue, out)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(out, []string{"a", "b"}) {
		t.Fatalf("after clear %v", out)
	}
}

func TestDescendantsBFSInto_bufferReuseLowAllocs(t *testing.T) {
	spec := GraphSpec{
		Phase: "plan",
		Nodes: []Node{
			{ID: "a", Kind: "x", Label: "a"},
			{ID: "b", Kind: "x", Label: "b"},
			{ID: "c", Kind: "x", Label: "c"},
		},
		Edges: []Edge{{From: "a", To: "b"}, {From: "b", To: "c"}},
	}
	fwd, _, err := Adjacencies(spec)
	if err != nil {
		t.Fatal(err)
	}
	visited := make(map[string]struct{}, 16)
	queue := make([]string, 0, 16)
	out := make([]string, 0, 16)
	allocs := testing.AllocsPerRun(50, func() {
		clear(visited)
		var e error
		out, e = DescendantsBFSInto(fwd, "a", visited, queue, out)
		if e != nil {
			t.Fatal(e)
		}
		if len(out) != 3 {
			t.Fatalf("len %d", len(out))
		}
	})
	if allocs > 0 {
		t.Fatalf("expected 0 allocs per run with reused buffers, got %f", allocs)
	}
}

func TestAdjacencies_duplicateNodeID(t *testing.T) {
	spec := GraphSpec{
		Phase: "plan",
		Nodes: []Node{
			{ID: "a", Kind: "x", Label: "a"},
			{ID: "a", Kind: "x", Label: "a2"},
		},
	}
	_, _, err := Adjacencies(spec)
	if err == nil {
		t.Fatal("expected error")
	}
}
