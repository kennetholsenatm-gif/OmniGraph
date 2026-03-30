package graph

import (
	"fmt"

	"github.com/kennetholsenatm-gif/omnigraph/internal/coerce"
	"github.com/kennetholsenatm-gif/omnigraph/internal/project"
	"github.com/kennetholsenatm-gif/omnigraph/internal/schema"
	"github.com/kennetholsenatm-gif/omnigraph/internal/security"
	"github.com/kennetholsenatm-gif/omnigraph/internal/state"
)

// EmitFromProjectRawOptions configures optional paths merged during graph v1 emission.
type EmitFromProjectRawOptions struct {
	PlanJSONPath  string
	TelemetryPath string
	TFStatePath   string
	SecurityPath  string
}

// EmitFromProjectRaw validates raw project bytes, parses intent, coerces, and emits omnigraph/graph/v1.
// It is the single pipeline previously used by automation entrypoints: human TOML/YAML/JSON project
// document in, machine graph JSON shape out (plus optional OpenTofu state/plan inputs from disk).
func EmitFromProjectRaw(raw []byte, opts EmitFromProjectRawOptions) (*Document, error) {
	if _, err := schema.ValidateRawDocument(raw); err != nil {
		return nil, err
	}
	doc, err := project.ParseProjectIntent(raw)
	if err != nil {
		return nil, err
	}
	art, err := coerce.FromDocument(doc)
	if err != nil {
		return nil, err
	}
	emitOpts := EmitOptions{PlanJSONPath: opts.PlanJSONPath, TelemetryPath: opts.TelemetryPath}
	if opts.TFStatePath != "" {
		st, err := state.Load(opts.TFStatePath)
		if err != nil {
			return nil, err
		}
		emitOpts.TerraformState = st
	}
	gdoc, err := Emit(doc, art, emitOpts)
	if err != nil {
		return nil, err
	}
	if opts.SecurityPath != "" {
		secdoc, err := security.LoadDocument(opts.SecurityPath)
		if err != nil {
			return nil, fmt.Errorf("security: %w", err)
		}
		MergeSecurity(gdoc, secdoc)
	}
	return gdoc, nil
}
