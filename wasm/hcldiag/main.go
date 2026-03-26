//go:build js && wasm

package main

import (
	"encoding/json"
	"syscall/js"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/kennetholsenatm-gif/omnigraph/wasm/tfpattern"
)

type diagnostic struct {
	Severity string `json:"severity"`
	Summary  string `json:"summary"`
	Detail   string `json:"detail"`
	Line     int    `json:"line,omitempty"`
	Column   int    `json:"column,omitempty"`
}

func main() {
	js.Global().Set("omnigraphHclValidate", js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) < 1 {
			return "[]"
		}
		src := args[0].String()
		out := validate([]byte(src))
		return mustJSON(out)
	}))
	js.Global().Set("omnigraphHclStructureLint", js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) < 1 {
			return "[]"
		}
		src := args[0].String()
		out := structureLint([]byte(src))
		return mustJSON(out)
	}))
	js.Global().Set("omnigraphTfPatternLint", js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) < 1 {
			return "[]"
		}
		src := args[0].String()
		out := patternFindingsToDiag(tfpattern.Scan([]byte(src)))
		return mustJSON(out)
	}))
	<-make(chan bool)
}

func patternFindingsToDiag(fs []tfpattern.Finding) []diagnostic {
	out := make([]diagnostic, 0, len(fs))
	for _, f := range fs {
		out = append(out, diagnostic{
			Severity: f.Severity,
			Summary:  f.Summary,
			Detail:   f.Detail,
			Line:     f.Line,
		})
	}
	return out
}

func mustJSON(out []diagnostic) string {
	b, err := json.Marshal(out)
	if err != nil {
		return `[{"severity":"error","summary":"marshal","detail":"internal"}]`
	}
	return string(b)
}

func validate(src []byte) []diagnostic {
	p := hclparse.NewParser()
	_, diags := p.ParseHCL(src, "snippet.tf")
	return diagsToOut(diags)
}

func structureLint(src []byte) []diagnostic {
	p := hclparse.NewParser()
	f, diags := p.ParseHCL(src, "snippet.tf")
	out := diagsToOut(diags)
	if diags.HasErrors() {
		return out
	}
	body, ok := f.Body.(*hclsyntax.Body)
	if !ok {
		return out
	}
	var hasTerraform, hasResource bool
	for _, b := range body.Blocks {
		switch b.Type {
		case "terraform":
			hasTerraform = true
		case "resource":
			hasResource = true
		}
	}
	if hasResource && !hasTerraform {
		out = append(out, diagnostic{
			Severity: "warning",
			Summary:  "missing terraform block",
			Detail:   "resources are declared without a terraform {} block; consider required_version and backend settings (tflint-style heuristic spike, ADR 001)",
		})
	}
	return out
}

func diagsToOut(diags hcl.Diagnostics) []diagnostic {
	var out []diagnostic
	for _, d := range diags {
		item := diagnostic{Summary: d.Summary, Detail: d.Detail}
		switch d.Severity {
		case hcl.DiagError:
			item.Severity = "error"
		case hcl.DiagWarning:
			item.Severity = "warning"
		default:
			item.Severity = "note"
		}
		if d.Subject != nil {
			item.Line = d.Subject.Start.Line
			item.Column = d.Subject.Start.Column
		}
		out = append(out, item)
	}
	return out
}
