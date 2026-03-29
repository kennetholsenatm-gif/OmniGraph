package graph

import (
	"errors"
	"fmt"
)

// Sentinel errors for document validation and parsing. Use errors.Is to branch.
var (
	ErrNilDocument     = errors.New("graph: document is nil")
	ErrWrongAPIVersion = errors.New("graph: invalid apiVersion")
	ErrWrongKind       = errors.New("graph: invalid kind")
	ErrEmptyPhase      = errors.New("graph: spec.phase is required")
	ErrEmptyNodes      = errors.New("graph: spec.nodes is required and cannot be empty")

	ErrEmptyNodeID    = errors.New("graph: node ID cannot be empty")
	ErrEmptyNodeKind  = errors.New("graph: node kind cannot be empty")
	ErrEmptyNodeLabel = errors.New("graph: node label cannot be empty")

	ErrEmptyEdgeFrom = errors.New("graph: edge from cannot be empty")
	ErrEmptyEdgeTo   = errors.New("graph: edge to cannot be empty")

	ErrOrphanNode              = errors.New("graph: orphan node has no incident edges")
	ErrMultipleWeakComponents  = errors.New("graph: multiple weakly connected components")
	ErrNilConcurrentGraph      = errors.New("graph: concurrent graph nil receiver")
	ErrConcurrentEmptyNodeID   = errors.New("graph: concurrent graph node id is empty")
	ErrConcurrentEmptyEdgeEnds = errors.New("graph: concurrent graph edge has empty from or to")
)

// ErrCycle is returned via Unwrap from *CycleError for errors.Is(err, graph.ErrCycle).
var ErrCycle = errors.New("graph: directed cycle")

// UnknownNodeError reports an edge or lookup referencing a missing node ID.
type UnknownNodeError struct {
	ID string
}

func (e *UnknownNodeError) Error() string {
	if e == nil {
		return "graph: unknown node"
	}
	return fmt.Sprintf("graph: unknown node %q", e.ID)
}

// TopoNodeError reports invalid node entries during topological ordering.
type TopoNodeError struct {
	Index int
	Err   error
}

func (e *TopoNodeError) Error() string {
	if e == nil {
		return "graph: topological order node error"
	}
	return fmt.Sprintf("topological order: node at index %d: %v", e.Index, e.Err)
}

func (e *TopoNodeError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// TopoDuplicateNodeIDError reports a duplicate node ID in the spec.
type TopoDuplicateNodeIDError struct {
	ID string
}

func (e *TopoDuplicateNodeIDError) Error() string {
	if e == nil {
		return "graph: duplicate node id"
	}
	return fmt.Sprintf("topological order: duplicate node id %q", e.ID)
}

// TopoEdgeError reports invalid edges during topological ordering.
type TopoEdgeError struct {
	Index int
	Err   error
}

func (e *TopoEdgeError) Error() string {
	if e == nil {
		return "graph: topological order edge error"
	}
	return fmt.Sprintf("topological order: edge at index %d: %v", e.Index, e.Err)
}

func (e *TopoEdgeError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}
