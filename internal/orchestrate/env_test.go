package orchestrate

import (
	"testing"

	"github.com/kennetholsenatm-gif/omnigraph/internal/coerce"
)

func TestMergeExecutionEnv(t *testing.T) {
	art := &coerce.Artifacts{
		Env: map[string]string{"OMNIGRAPH_PROJECT_NAME": "p"},
		TerraformTfvarsJSON: map[string]any{
			"project_name": "demo",
			"public_ports": []any{80.0, 443.0},
		},
	}
	m := MergeExecutionEnv(art)
	if m["OMNIGRAPH_PROJECT_NAME"] != "p" {
		t.Fatalf("omnigraph env: %v", m)
	}
	if m["TF_VAR_project_name"] != "demo" {
		t.Fatalf("TF_VAR_project_name: %v", m)
	}
	if m["TF_VAR_public_ports"] == "" {
		t.Fatalf("expected JSON list for public_ports: %v", m)
	}
}
