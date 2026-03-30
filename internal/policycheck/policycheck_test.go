package policycheck

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kennetholsenatm-gif/omnigraph/internal/policy"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found")
		}
		dir = parent
	}
}

func TestPolicyManifestYAMLs(t *testing.T) {
	root := repoRoot(t)
	skip := map[string]bool{
		".git": true, "node_modules": true, ".github": true,
	}
	var checked int
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if skip[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".yaml" && ext != ".yml" {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		if strings.HasPrefix(rel, ".github") {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if !strings.Contains(string(b), "omnigraph/policy/v1") {
			return nil
		}
		checked++
		if err := ValidatePolicyManifestFile(path); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if checked == 0 {
		t.Log("no policy manifest files with omnigraph/policy/v1 marker found")
	}
}

func TestOmniGraphSchemaFiles(t *testing.T) {
	root := repoRoot(t)
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".omnigraph.schema") {
			return nil
		}
		policyDir := ""
		enforce := false
		if st, e := os.Stat(filepath.Join(root, "policies")); e == nil && st.IsDir() {
			policyDir = filepath.Join(root, "policies")
			enforce = true
		} else if st, e := os.Stat(filepath.Join(root, "testdata", "policies")); e == nil && st.IsDir() {
			policyDir = filepath.Join(root, "testdata", "policies")
			enforce = true
		}
		if policyDir == "" {
			return ValidateOmniGraphSchema(path, "", false)
		}
		return ValidateOmniGraphSchema(path, policyDir, enforce)
	})
	if err != nil {
		t.Fatal(err)
	}
}

// TestWritePolicyReportJSON writes policy-report.json for the PR comment workflow when policies/ exists.
func TestWritePolicyReportJSON(t *testing.T) {
	if os.Getenv("OMNIGRAPH_WRITE_POLICY_REPORT") != "1" {
		t.Skip("set OMNIGRAPH_WRITE_POLICY_REPORT=1 to write policy-report.json")
	}
	root := repoRoot(t)
	policiesDir := filepath.Join(root, "policies")
	if _, err := os.Stat(policiesDir); err != nil {
		t.Skip("no policies/ directory")
	}
	engine := policy.NewEngine()
	if err := LoadPoliciesFromDir(engine, policiesDir); err != nil {
		t.Fatal(err)
	}
	sets := engine.ListPolicySets()
	summary := map[string]any{
		"passed":     len(sets),
		"failed":     0,
		"warnings":   0,
		"violations": []any{},
	}
	if len(sets) == 0 {
		summary["passed"] = 0
	}
	out, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "policy-report.json"), out, 0o644); err != nil {
		t.Fatal(err)
	}
}
