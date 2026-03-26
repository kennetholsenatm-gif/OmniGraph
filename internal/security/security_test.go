package security

import (
	"context"
	"strings"
	"testing"
)

type scriptHost struct {
	label  string
	linux  bool
	script map[string]struct{ out, err string; code int }
}

func (h scriptHost) Label() string { return h.label }

func (h scriptHost) Run(ctx context.Context, argv []string) (stdout, stderr string, exitCode int, err error) {
	key := strings.Join(argv, "\x00")
	if key == "uname\x00-s" {
		if h.linux {
			return "Linux", "", 0, nil
		}
		return "Windows_NT", "", 0, nil
	}
	if r, ok := h.script[key]; ok {
		return r.out, r.err, r.code, nil
	}
	return "", "not in script", 127, nil
}

func TestRun_FilterModule(t *testing.T) {
	ctx := context.Background()
	h := scriptHost{label: "t", linux: true, script: map[string]struct{ out, err string; code int }{
		"uname\x00-a": {out: "Linux test 1", code: 0},
	}}
	doc := Run(ctx, h, "local", "p", "", Filter{ModuleID: "T1082_system_info"}, 0)
	if doc.Spec.Summary.ModulesRun != 1 {
		t.Fatalf("modulesRun=%d", doc.Spec.Summary.ModulesRun)
	}
	if len(doc.Spec.Results) != 1 || doc.Spec.Results[0].ModuleID != "T1082_system_info" {
		t.Fatalf("results=%v", doc.Spec.Results)
	}
}

func TestParseDocument_Sample(t *testing.T) {
	doc, err := LoadDocument("../../testdata/sample.security.json")
	if err != nil {
		t.Fatal(err)
	}
	if doc.Spec.Summary.ModulesRun != 2 {
		t.Fatalf("summary=%+v", doc.Spec.Summary)
	}
}
