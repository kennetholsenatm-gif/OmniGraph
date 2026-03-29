package orchestrate

import (
	"fmt"
	"testing"

	"github.com/kennetholsenatm-gif/omnigraph/internal/graph"
)

func TestShouldAbortGraphPipeline(t *testing.T) {
	if ShouldAbortGraphPipeline(nil) {
		t.Fatal("nil should not abort")
	}
	if !ShouldAbortGraphPipeline(fmt.Errorf("wrap: %w", graph.ErrWrongAPIVersion)) {
		t.Fatal("wrong api version should abort")
	}
	if !ShouldAbortGraphPipeline(&graph.CycleError{Path: []string{"a", "a"}}) {
		t.Fatal("cycle should abort")
	}
	if !ShouldAbortGraphPipeline(&graph.UnknownNodeError{ID: "x"}) {
		t.Fatal("unknown node should abort")
	}
	if ShouldAbortGraphPipeline(fmt.Errorf("transient network error")) {
		t.Fatal("generic error should not abort via graph policy")
	}
}
