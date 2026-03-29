package orchestrate

import (
	"testing"

	"github.com/kennetholsenatm-gif/omnigraph/internal/graph"
)

func TestEvaluateBlastRadiusPolicy_zeroDisabled(t *testing.T) {
	doc := &graph.Document{
		Spec: graph.GraphSpec{
			Phase: "plan",
			Nodes: []graph.Node{{ID: "a", Kind: "k", Label: "l"}},
			Edges: nil,
		},
	}
	rep := &graph.BlastRadiusReport{
		AffectedNodeIDs: []string{"a", "b", "c"},
		TotalNodeCount:  10,
	}
	if err := EvaluateBlastRadiusPolicy(BlastRadiusPolicy{}, rep, doc); err != nil {
		t.Fatal(err)
	}
}

func TestEvaluateBlastRadiusPolicy_maxCount(t *testing.T) {
	doc := &graph.Document{
		Spec: graph.GraphSpec{
			Phase: "plan",
			Nodes: []graph.Node{{ID: "a", Kind: "k", Label: "l"}},
		},
	}
	rep := &graph.BlastRadiusReport{AffectedNodeIDs: []string{"x", "y"}, TotalNodeCount: 2}
	err := EvaluateBlastRadiusPolicy(BlastRadiusPolicy{MaxAffectedNodeCount: 1}, rep, doc)
	if err == nil {
		t.Fatal("expected error")
	}
	if !ShouldAbortGraphPipeline(err) {
		t.Fatal("blast policy errors should abort pipeline")
	}
}
