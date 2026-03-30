package omnistate

import (
	"context"
	"testing"
)

func TestCompareIntendedVsRuntime_DegradedAndFractured(t *testing.T) {
	intended := &OmniGraphState{
		APIVersion: APIVersion,
		Nodes: []StateNode{
			{ID: "a", Kind: "ansible_host", Label: "a", Attributes: map[string]any{"ansible_host": "10.0.0.1"}},
			{ID: "b", Kind: "ansible_host", Label: "b", Attributes: map[string]any{"ansible_host": "10.0.0.2"}},
		},
		Edges: []StateEdge{
			{From: "a", To: "g", Kind: "member_of"},
			{From: "b", To: "g", Kind: "member_of"},
		},
	}
	runtime := &OmniGraphState{
		APIVersion: APIVersion,
		Nodes: []StateNode{
			{ID: "a", Kind: "ansible_host", Label: "a", Attributes: map[string]any{"ansible_host": "10.0.0.9"}},
			{ID: "g", Kind: "ansible_group", Label: "g"},
		},
		Edges: []StateEdge{},
	}
	rep, err := CompareIntendedVsRuntime(context.Background(), intended, runtime)
	if err != nil {
		t.Fatal(err)
	}
	if len(rep.DegradedNodes) != 2 {
		t.Fatalf("degraded nodes: got %d want 2", len(rep.DegradedNodes))
	}
	var sawDrift, sawMissing bool
	for _, d := range rep.DegradedNodes {
		for _, r := range d.Reasons {
			if r == "drift" {
				sawDrift = true
			}
			if r == "unresolved_reference" {
				sawMissing = true
			}
		}
	}
	if !sawDrift {
		t.Fatal("expected drift on node a")
	}
	if !sawMissing {
		t.Fatal("expected unresolved_reference for missing node b")
	}
	if len(rep.FracturedEdges) != 1 {
		t.Fatalf("fractured edges: got %d want 1", len(rep.FracturedEdges))
	}
	if rep.FracturedEdges[0].TranslucentTarget == "" {
		t.Fatal("expected translucent target on fractured edge")
	}
}

func TestApplyPatch_UpsertRemove(t *testing.T) {
	st := OmniGraphState{
		APIVersion: APIVersion,
		Nodes: []StateNode{
			{ID: "x", Kind: "k", Label: "x"},
		},
		Edges: []StateEdge{
			{From: "x", To: "y", Kind: "e"},
		},
	}
	next := ApplyPatch(st, StatePatch{
		RemoveNodes: []string{"x"},
		UpsertNodes: []StateNode{{ID: "z", Kind: "k", Label: "z"}},
	})
	var sawZ bool
	for _, n := range next.Nodes {
		if n.ID == "z" {
			sawZ = true
		}
		if n.ID == "x" {
			t.Fatal("node x should be removed")
		}
	}
	if !sawZ {
		t.Fatal("expected upsert z")
	}
}
