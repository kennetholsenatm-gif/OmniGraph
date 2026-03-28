package reconciler

import (
	"context"
	"errors"
	"fmt"
)

// ErrNotImplemented is returned by backend stubs until an emitter is completed.
var ErrNotImplemented = errors.New("reconciler: emitter not implemented")

// Artifact is one generated file or blob (logical path + content).
type Artifact struct {
	Path        string
	MediaType   string
	Description string
	Content     []byte
}

// Backend translates IR into concrete artifacts for a single format.
type Backend interface {
	Format() string
	Emit(ctx context.Context, doc *Document) ([]Artifact, error)
}

// Registry holds named backends.
type Registry struct {
	byFormat map[string]Backend
}

// NewRegistry returns an empty registry.
func NewRegistry() *Registry {
	return &Registry{byFormat: make(map[string]Backend)}
}

// Register adds a backend; panics if format is empty or duplicate.
func (r *Registry) Register(b Backend) {
	if b == nil {
		panic("reconciler: Register nil backend")
	}
	f := b.Format()
	if f == "" {
		panic("reconciler: Register empty format")
	}
	if _, ok := r.byFormat[f]; ok {
		panic("reconciler: duplicate backend " + f)
	}
	r.byFormat[f] = b
}

// Get returns a backend or nil.
func (r *Registry) Get(format string) Backend {
	if r == nil {
		return nil
	}
	return r.byFormat[format]
}

// Emit runs a single backend by format id.
func (r *Registry) Emit(ctx context.Context, format string, doc *Document) ([]Artifact, error) {
	if doc == nil {
		return nil, fmt.Errorf("reconciler: nil document")
	}
	b := r.Get(format)
	if b == nil {
		return nil, fmt.Errorf("reconciler: unknown backend %q", format)
	}
	return b.Emit(ctx, doc)
}

// stubBackend registers a placeholder for every known format until implemented.
type stubBackend struct {
	format string
}

func (stubBackend) Emit(context.Context, *Document) ([]Artifact, error) {
	return nil, ErrNotImplemented
}

func (s stubBackend) Format() string { return s.format }

// DefaultRegistry registers real emitters where implemented and stubs for the rest.
func DefaultRegistry() *Registry {
	r := NewRegistry()
	implemented := map[string]Backend{
		AnsibleInventoryINI: ansibleInventoryBackend{},
	}
	for _, f := range AllFormats() {
		if b, ok := implemented[f]; ok {
			r.Register(b)
			continue
		}
		r.Register(stubBackend{format: f})
	}
	return r
}
