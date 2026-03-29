package graph

import "testing"

func TestComputeBlastRadius_downstreamNecessary(t *testing.T) {
	doc := &Document{
		APIVersion: apiVersion,
		Kind:       kind,
		Spec: GraphSpec{
			Phase: "plan",
			Nodes: []Node{
				{ID: "tf", Kind: "tool", Label: "tf"},
				{ID: "planned-aws_instance_x", Kind: "resource", Label: "aws_instance.x"},
				{ID: "n1", Kind: "service", Label: "one"},
				{ID: "n2", Kind: "service", Label: "two"},
			},
			Edges: []Edge{
				{From: "tf", To: "planned-aws_instance_x", Kind: "mutates"},
				{From: "planned-aws_instance_x", To: "n1", Kind: "depends", DependencyRole: EdgeDependencyNecessary},
				{From: "n1", To: "n2", Kind: "depends", DependencyRole: EdgeDependencyNecessary},
			},
		},
	}
	rep, err := ComputeBlastRadius(doc, []string{"aws_instance.x"}, []string{"planned-aws_instance_x"})
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]struct{}{
		"n1": struct{}{}, "n2": struct{}{}, "planned-aws_instance_x": struct{}{},
	}
	if len(rep.AffectedNodeIDs) != len(want) {
		t.Fatalf("AffectedNodeIDs = %v len %d want %d", rep.AffectedNodeIDs, len(rep.AffectedNodeIDs), len(want))
	}
	for _, id := range rep.AffectedNodeIDs {
		if _, ok := want[id]; !ok {
			t.Fatalf("unexpected id %q", id)
		}
	}
	if rep.AffectedEdgeCount != 2 {
		t.Fatalf("AffectedEdgeCount = %d want 2", rep.AffectedEdgeCount)
	}
	if rep.TotalNodeCount != 4 {
		t.Fatalf("TotalNodeCount = %d", rep.TotalNodeCount)
	}
}

func TestComputeBlastRadius_emptySeeds(t *testing.T) {
	doc := &Document{
		APIVersion: apiVersion,
		Kind:       kind,
		Spec: GraphSpec{
			Phase: "plan",
			Nodes: []Node{{ID: "x", Kind: "k", Label: "l"}},
			Edges: nil,
		},
	}
	rep, err := ComputeBlastRadius(doc, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(rep.AffectedNodeIDs) != 0 {
		t.Fatalf("want empty affected, got %v", rep.AffectedNodeIDs)
	}
}
