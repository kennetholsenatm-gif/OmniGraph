package policycheck

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kennetholsenatm-gif/omnigraph/internal/policy"
	"github.com/kennetholsenatm-gif/omnigraph/internal/project"
	"github.com/kennetholsenatm-gif/omnigraph/internal/schema"
	"gopkg.in/yaml.v3"
)

// LoadPoliciesFromDir loads *.yaml, *.yml, *.json policy sets from a directory into the engine.
func LoadPoliciesFromDir(engine *policy.Engine, dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read policy dir: %w", err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := filepath.Ext(entry.Name())
		if ext != ".yaml" && ext != ".yml" && ext != ".json" {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		if _, err := engine.LoadPolicySet(path); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to load %s: %v\n", path, err)
		}
	}
	return nil
}

// ValidateOmniGraphSchema validates one .omnigraph.schema file; optional policyDir with enforce flag.
func ValidateOmniGraphSchema(path string, policyDir string, enforce bool) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if _, err := schema.ValidateRawDocument(raw); err != nil {
		return fmt.Errorf("%s: schema: %w", path, err)
	}
	doc, err := project.ParseProjectIntent(raw)
	if err != nil {
		return fmt.Errorf("%s: parse: %w", path, err)
	}
	if doc.Metadata.Name == "" {
		return fmt.Errorf("%s: missing metadata.name", path)
	}
	if policyDir == "" {
		return nil
	}
	ctx := context.Background()
	engine := policy.NewEngine()
	if err := LoadPoliciesFromDir(engine, policyDir); err != nil {
		return fmt.Errorf("%s: load policies: %w", path, err)
	}
	sets := engine.ListPolicySets()
	if len(sets) == 0 {
		fmt.Fprintf(os.Stderr, "warning: no policy sets found in %s\n", policyDir)
		return nil
	}
	var input interface{}
	trim := bytes.TrimSpace(raw)
	if len(trim) == 0 {
		return fmt.Errorf("%s: empty document", path)
	}
	switch trim[0] {
	case '{', '[':
		if err := json.Unmarshal(trim, &input); err != nil {
			return fmt.Errorf("%s: policy input json: %w", path, err)
		}
	default:
		if err := yaml.Unmarshal(trim, &input); err != nil {
			return fmt.Errorf("%s: policy input yaml: %w", path, err)
		}
	}
	for _, ps := range sets {
		report, err := engine.Evaluate(ctx, ps.Metadata.Name, input)
		if err != nil {
			return fmt.Errorf("%s: evaluate %s: %w", path, ps.Metadata.Name, err)
		}
		if len(report.Violations) > 0 && enforce && report.Enforcement == "deny" {
			return fmt.Errorf("%s: policy violations in %s", path, ps.Metadata.Name)
		}
	}
	return nil
}

// ValidatePolicyManifestFile loads a single policy manifest (must contain omnigraph/policy/v1 marker).
func ValidatePolicyManifestFile(path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if !strings.Contains(string(b), "omnigraph/policy/v1") {
		return nil
	}
	engine := policy.NewEngine()
	if _, err := engine.LoadPolicySet(path); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	return nil
}
