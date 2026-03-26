package policy

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/storage/inmem"
	"gopkg.in/yaml.v3"
)

// PolicySet represents a collection of policies
type PolicySet struct {
	APIVersion string   `json:"apiVersion" yaml:"apiVersion"`
	Kind       string   `json:"kind" yaml:"kind"`
	Metadata   Metadata `json:"metadata" yaml:"metadata"`
	Spec       Spec     `json:"spec" yaml:"spec"`
}

// Metadata contains policy set metadata
type Metadata struct {
	Name        string            `json:"name" yaml:"name"`
	Version     string            `json:"version,omitempty" yaml:"version,omitempty"`
	Description string            `json:"description,omitempty" yaml:"description,omitempty"`
	Labels      map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
}

// Spec contains policy set specification
type Spec struct {
	TargetKinds []string `json:"targetKinds,omitempty" yaml:"targetKinds,omitempty"`
	Enforcement string   `json:"enforcement,omitempty" yaml:"enforcement,omitempty"` // warn, deny
	Policies    []Policy `json:"policies" yaml:"policies"`
}

// Policy represents a single policy
type Policy struct {
	Name        string   `json:"name" yaml:"name"`
	Description string   `json:"description,omitempty" yaml:"description,omitempty"`
	Severity    string   `json:"severity,omitempty" yaml:"severity,omitempty"` // info, warning, error, critical
	Rego        string   `json:"rego,omitempty" yaml:"rego,omitempty"`
	RegoFile    string   `json:"regoFile,omitempty" yaml:"regoFile,omitempty"`
	InputSchema string   `json:"inputSchema,omitempty" yaml:"inputSchema,omitempty"`
	Tags        []string `json:"tags,omitempty" yaml:"tags,omitempty"`
}

// Violation represents a policy violation
type Violation struct {
	Policy      string `json:"policy" yaml:"policy"`
	Severity    string `json:"severity" yaml:"severity"`
	Message     string `json:"message" yaml:"message"`
	Path        string `json:"path,omitempty" yaml:"path,omitempty"`
	Resource    string `json:"resource,omitempty" yaml:"resource,omitempty"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

// PolicyReport represents a policy evaluation report
type PolicyReport struct {
	Timestamp   time.Time   `json:"timestamp" yaml:"timestamp"`
	PolicySet   string      `json:"policySet" yaml:"policySet"`
	Enforcement string      `json:"enforcement" yaml:"enforcement"`
	Violations  []Violation `json:"violations" yaml:"violations"`
	Passed      int         `json:"passed" yaml:"passed"`
	Failed      int         `json:"failed" yaml:"failed"`
	Warnings    int         `json:"warnings" yaml:"warnings"`
}

// Engine manages policy evaluation
type Engine struct {
	policySets map[string]*PolicySet
	compiled   map[string]*ast.Compiler
}

// NewEngine creates a new policy engine
func NewEngine() *Engine {
	return &Engine{
		policySets: make(map[string]*PolicySet),
		compiled:   make(map[string]*ast.Compiler),
	}
}

// EngineKey returns the internal registry key for a policy set (name, or name:version when version is set).
func EngineKey(ps *PolicySet) string {
	if ps.Metadata.Version != "" {
		return fmt.Sprintf("%s:%s", ps.Metadata.Name, ps.Metadata.Version)
	}
	return ps.Metadata.Name
}

func (e *Engine) resolvePolicySetKey(nameOrKey string) (string, error) {
	if _, ok := e.policySets[nameOrKey]; ok {
		return nameOrKey, nil
	}
	for key, ps := range e.policySets {
		if ps.Metadata.Name == nameOrKey {
			return key, nil
		}
	}
	return "", fmt.Errorf("policy set not found: %s", nameOrKey)
}

// LoadPolicySet loads a policy set from a file
func (e *Engine) LoadPolicySet(path string) (*PolicySet, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read policy file: %w", err)
	}

	var policySet PolicySet
	ext := filepath.Ext(path)
	if ext == ".yaml" || ext == ".yml" {
		if err := yaml.Unmarshal(data, &policySet); err != nil {
			return nil, fmt.Errorf("failed to parse YAML: %w", err)
		}
	} else {
		if err := json.Unmarshal(data, &policySet); err != nil {
			return nil, fmt.Errorf("failed to parse JSON: %w", err)
		}
	}

	// Validate policy set
	if err := e.validatePolicySet(&policySet); err != nil {
		return nil, fmt.Errorf("invalid policy set: %w", err)
	}

	// Load external Rego files
	for i := range policySet.Spec.Policies {
		if policySet.Spec.Policies[i].RegoFile != "" {
			regoPath := filepath.Join(filepath.Dir(path), policySet.Spec.Policies[i].RegoFile)
			regoData, err := os.ReadFile(regoPath)
			if err != nil {
				return nil, fmt.Errorf("failed to read Rego file %s: %w", regoPath, err)
			}
			policySet.Spec.Policies[i].Rego = string(regoData)
		}
	}

	// Compile policies
	if err := e.compilePolicies(&policySet); err != nil {
		return nil, fmt.Errorf("failed to compile policies: %w", err)
	}

	key := EngineKey(&policySet)
	e.policySets[key] = &policySet

	return &policySet, nil
}

// validatePolicySet validates a policy set
func (e *Engine) validatePolicySet(ps *PolicySet) error {
	if ps.APIVersion != "omnigraph/policy/v1" {
		return fmt.Errorf("unsupported apiVersion: %s", ps.APIVersion)
	}
	if ps.Kind != "PolicySet" {
		return fmt.Errorf("unsupported kind: %s", ps.Kind)
	}
	if ps.Metadata.Name == "" {
		return fmt.Errorf("metadata.name is required")
	}
	if len(ps.Spec.Policies) == 0 {
		return fmt.Errorf("spec.policies must contain at least one policy")
	}

	// Validate enforcement level
	if ps.Spec.Enforcement != "" {
		if ps.Spec.Enforcement != "warn" && ps.Spec.Enforcement != "deny" {
			return fmt.Errorf("invalid enforcement level: %s (must be 'warn' or 'deny')", ps.Spec.Enforcement)
		}
	}

	// Validate policies
	for i, policy := range ps.Spec.Policies {
		if policy.Name == "" {
			return fmt.Errorf("policy[%d].name is required", i)
		}
		if policy.Rego == "" && policy.RegoFile == "" {
			return fmt.Errorf("policy[%d] must have either rego or regoFile", i)
		}
		if policy.Severity != "" {
			validSeverities := []string{"info", "warning", "error", "critical"}
			valid := false
			for _, s := range validSeverities {
				if policy.Severity == s {
					valid = true
					break
				}
			}
			if !valid {
				return fmt.Errorf("policy[%d].severity must be one of: %v", i, validSeverities)
			}
		}
	}

	return nil
}

// compilePolicies compiles all Rego policies in a policy set
func (e *Engine) compilePolicies(ps *PolicySet) error {
	modules := make(map[string]string)

	for _, policy := range ps.Spec.Policies {
		if policy.Rego == "" {
			continue
		}

		moduleName := fmt.Sprintf("%s/%s", ps.Metadata.Name, policy.Name)
		modules[moduleName] = policy.Rego
	}

	if len(modules) == 0 {
		return nil
	}

	compiler, err := ast.CompileModules(modules)
	if err != nil {
		return fmt.Errorf("failed to compile Rego modules: %w", err)
	}

	e.compiled[EngineKey(ps)] = compiler

	return nil
}

// Evaluate evaluates input against a policy set
func (e *Engine) Evaluate(ctx context.Context, policySetName string, input interface{}) (*PolicyReport, error) {
	key, err := e.resolvePolicySetKey(policySetName)
	if err != nil {
		return nil, err
	}

	ps, ok := e.policySets[key]
	if !ok {
		return nil, fmt.Errorf("policy set not found: %s", policySetName)
	}

	compiler, ok := e.compiled[key]
	if !ok {
		return nil, fmt.Errorf("compiled policies not found: %s", policySetName)
	}

	report := &PolicyReport{
		Timestamp:   time.Now(),
		PolicySet:   policySetName,
		Enforcement: ps.Spec.Enforcement,
		Violations:  []Violation{},
	}

	// Convert input to map for OPA
	inputMap, err := convertToMap(input)
	if err != nil {
		return nil, fmt.Errorf("failed to convert input: %w", err)
	}

	// Evaluate each policy
	for _, policy := range ps.Spec.Policies {
		violations, err := e.evaluatePolicy(ctx, compiler, policy, inputMap)
		if err != nil {
			return nil, fmt.Errorf("failed to evaluate policy %s: %w", policy.Name, err)
		}

		if len(violations) > 0 {
			report.Violations = append(report.Violations, violations...)
			report.Failed++

			// Count warnings
			for _, v := range violations {
				if v.Severity == "warning" {
					report.Warnings++
				}
			}
		} else {
			report.Passed++
		}
	}

	return report, nil
}

// evaluatePolicy evaluates input against a single policy
func (e *Engine) evaluatePolicy(ctx context.Context, compiler *ast.Compiler, policy Policy, input map[string]interface{}) ([]Violation, error) {
	// Extract package name from Rego (path after `package`, without `data.` prefix).
	packageName := extractPackageName(policy.Rego)
	if packageName == "" {
		packageName = "omnigraph.policy"
	}
	if strings.HasPrefix(packageName, "data.") {
		packageName = strings.TrimPrefix(packageName, "data.")
	}

	// OPA requires the `data.` document root; `omnigraph.compliance.deny` is parsed as vars.
	query := "data." + packageName + ".deny"

	// Create Rego query
	r := rego.New(
		rego.Query(query),
		rego.Compiler(compiler),
		rego.Input(input),
		rego.Store(inmem.New()),
	)

	// Evaluate
	rs, err := r.Eval(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate Rego: %w", err)
	}

	var violations []Violation

	// Process results
	for _, result := range rs {
		for _, expr := range result.Expressions {
			value := expr.Value

			// Handle different result types
			switch v := value.(type) {
			case string:
				violations = append(violations, Violation{
					Policy:      policy.Name,
					Severity:    getSeverity(policy.Severity),
					Message:     v,
					Description: policy.Description,
				})
			case []interface{}:
				for _, item := range v {
					if msg, ok := item.(string); ok {
						violations = append(violations, Violation{
							Policy:      policy.Name,
							Severity:    getSeverity(policy.Severity),
							Message:     msg,
							Description: policy.Description,
						})
					}
				}
			case map[string]interface{}:
				if msg, ok := v["msg"].(string); ok {
					violation := Violation{
						Policy:      policy.Name,
						Severity:    getSeverity(policy.Severity),
						Message:     msg,
						Description: policy.Description,
					}
					if path, ok := v["path"].(string); ok {
						violation.Path = path
					}
					if resource, ok := v["resource"].(string); ok {
						violation.Resource = resource
					}
					violations = append(violations, violation)
				}
			}
		}
	}

	return violations, nil
}

// EvaluateFile evaluates a file against a policy set
func (e *Engine) EvaluateFile(ctx context.Context, policySetName, filePath string) (*PolicyReport, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var input interface{}
	ext := filepath.Ext(filePath)
	if ext == ".yaml" || ext == ".yml" {
		if err := yaml.Unmarshal(data, &input); err != nil {
			return nil, fmt.Errorf("failed to parse YAML: %w", err)
		}
	} else {
		if err := json.Unmarshal(data, &input); err != nil {
			return nil, fmt.Errorf("failed to parse JSON: %w", err)
		}
	}

	return e.Evaluate(ctx, policySetName, input)
}

// GetPolicySet returns a policy set by engine key or metadata.name
func (e *Engine) GetPolicySet(name string) (*PolicySet, error) {
	key, err := e.resolvePolicySetKey(name)
	if err != nil {
		return nil, err
	}
	return e.policySets[key], nil
}

// ListPolicySets returns all loaded policy sets
func (e *Engine) ListPolicySets() []*PolicySet {
	sets := make([]*PolicySet, 0, len(e.policySets))
	for _, ps := range e.policySets {
		sets = append(sets, ps)
	}
	return sets
}

// RemovePolicySet removes a policy set
func (e *Engine) RemovePolicySet(name string) {
	delete(e.policySets, name)
	delete(e.compiled, name)
}

// Helper functions

func convertToMap(input interface{}) (map[string]interface{}, error) {
	switch v := input.(type) {
	case map[string]interface{}:
		return v, nil
	default:
		// Marshal to JSON and back to map
		data, err := json.Marshal(input)
		if err != nil {
			return nil, err
		}
		var result map[string]interface{}
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, err
		}
		return result, nil
	}
}

func extractPackageName(rego string) string {
	lines := strings.Split(rego, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "package ") {
			return strings.TrimPrefix(line, "package ")
		}
	}
	return ""
}

func getSeverity(severity string) string {
	if severity == "" {
		return "error"
	}
	return severity
}
