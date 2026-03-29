package orchestrate

import (
	"errors"

	"github.com/kennetholsenatm-gif/omnigraph/internal/graph"
)

// ShouldAbortGraphPipeline reports whether err is a deterministic graph validation or topology
// problem that should not be retried (fix data or schema instead).
func ShouldAbortGraphPipeline(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, graph.ErrWrongAPIVersion) || errors.Is(err, graph.ErrWrongKind) {
		return true
	}
	if errors.Is(err, graph.ErrEmptyPhase) || errors.Is(err, graph.ErrEmptyNodes) {
		return true
	}
	if errors.Is(err, graph.ErrNilDocument) {
		return true
	}
	var ce *graph.CycleError
	if errors.As(err, &ce) {
		return true
	}
	var u *graph.UnknownNodeError
	if errors.As(err, &u) {
		return true
	}
	return false
}
