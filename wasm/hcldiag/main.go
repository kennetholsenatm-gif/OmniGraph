//go:build js && wasm

package main

import (
	"encoding/json"
	"syscall/js"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
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
		b, err := json.Marshal(out)
		if err != nil {
			return `[{"severity":"error","summary":"marshal","detail":"internal"}]`
		}
		return string(b)
	}))
	<-make(chan bool)
}

func validate(src []byte) []diagnostic {
	p := hclparse.NewParser()
	_, diags := p.ParseHCL(src, "snippet.tf")
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
