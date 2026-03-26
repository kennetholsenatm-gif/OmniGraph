package policy

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// PolicyTemplate represents a policy template
type PolicyTemplate struct {
	Name        string
	Description string
	Category    string
	Template    string
	Variables   map[string]string
}

// TemplateManager manages policy templates
type TemplateManager struct {
	templates map[string]*PolicyTemplate
}

// NewTemplateManager creates a new template manager
func NewTemplateManager() *TemplateManager {
	tm := &TemplateManager{
		templates: make(map[string]*PolicyTemplate),
	}
	
	// Register built-in templates
	tm.registerBuiltinTemplates()
	
	return tm
}

// registerBuiltinTemplates registers built-in policy templates
func (tm *TemplateManager) registerBuiltinTemplates() {
	// Security templates
	tm.Register(&PolicyTemplate{
		Name:        "security-no-public-ssh",
		Description: "Prevent SSH access from 0.0.0.0/0",
		Category:    "security",
		Template: `package omnigraph.security

# Prevent SSH access from anywhere
deny[msg] {
    input.spec.components[i].componentType == "omnigraph.network.security_group"
    ingress := input.spec.components[i].config.ingress[j]
    ingress.port == {{.port}}
    ingress.cidr == "0.0.0.0/0"
    msg := sprintf("Security group '%s' allows SSH from 0.0.0.0/0", [input.spec.components[i].id])
}`,
		Variables: map[string]string{
			"port": "22",
		},
	})

	tm.Register(&PolicyTemplate{
		Name:        "security-require-encryption",
		Description: "Require encryption at rest for storage resources",
		Category:    "security",
		Template: `package omnigraph.security

# Require encryption for storage volumes
deny[msg] {
    input.spec.components[i].componentType == "{{.resourceType}}"
    not input.spec.components[i].config.encrypted
    msg := sprintf("Storage volume '%s' must have encryption enabled", [input.spec.components[i].id])
}`,
		Variables: map[string]string{
			"resourceType": "omnigraph.storage.volume",
		},
	})

	// Compliance templates
	tm.Register(&PolicyTemplate{
		Name:        "compliance-require-tags",
		Description: "Require mandatory tags on all resources",
		Category:    "compliance",
		Template: `package omnigraph.compliance

# Require specific tags
required_tags := {{.tags}}

deny[msg] {
    input.spec.components[i].componentType
    provided_tags := {tag | input.spec.components[i].config.tags[tag]}
    missing_tags := required_tags - provided_tags
    count(missing_tags) > 0
    msg := sprintf("Component '%s' is missing required tags: %v", [input.spec.components[i].id, missing_tags])
}`,
		Variables: map[string]string{
			"tags": `{"environment", "owner", "cost-center"}`,
		},
	})

	tm.Register(&PolicyTemplate{
		Name:        "compliance-naming-convention",
		Description: "Enforce naming conventions for resources",
		Category:    "compliance",
		Template: `package omnigraph.compliance

# Enforce naming convention
deny[msg] {
    input.spec.components[i].componentType == "{{.resourceType}}"
    name := input.spec.components[i].id
    not regex.match("{{.pattern}}", name)
    msg := sprintf("Resource '%s' must match naming convention: %s", [name, "{{.pattern}}"])
}`,
		Variables: map[string]string{
			"resourceType": "omnigraph.compute.instance",
			"pattern":      "^[a-z][a-z0-9-]*[a-z0-9]$",
		},
	})

	// Cost templates
	tm.Register(&PolicyTemplate{
		Name:        "cost-limit-instance-size",
		Description: "Limit instance sizes to control costs",
		Category:    "cost",
		Template: `package omnigraph.cost

# Limit instance CPU
deny[msg] {
    input.spec.components[i].componentType == "omnigraph.compute.instance"
    cpu := input.spec.components[i].config.cpu
    cpu > {{.maxCpu}}
    msg := sprintf("Instance '%s' requests %d CPUs, maximum allowed is %d", [input.spec.components[i].id, cpu, {{.maxCpu}}])
}

# Limit instance memory
deny[msg] {
    input.spec.components[i].componentType == "omnigraph.compute.instance"
    memory_gb := input.spec.components[i].config.memory_gb
    memory_gb > {{.maxMemory}}
    msg := sprintf("Instance '%s' requests %d GB memory, maximum allowed is %d GB", [input.spec.components[i].id, memory_gb, {{.maxMemory}}])
}`,
		Variables: map[string]string{
			"maxCpu":    "16",
			"maxMemory": "64",
		},
	})

	// Network templates
	tm.Register(&PolicyTemplate{
		Name:        "network-require-private-subnet",
		Description: "Require resources to use private subnets",
		Category:    "network",
		Template: `package omnigraph.network

# Require private subnet
deny[msg] {
    input.spec.components[i].config.subnet
    subnet := input.spec.components[i].config.subnet
    not startswith(subnet, "10.")
    not startswith(subnet, "192.168.")
    not startswith(subnet, "172.")
    msg := sprintf("Component '%s' must use private subnet, got: %s", [input.spec.components[i].id, subnet])
}`,
		Variables: map[string]string{},
	})
}

// Register registers a new policy template
func (tm *TemplateManager) Register(tmpl *PolicyTemplate) {
	tm.templates[tmpl.Name] = tmpl
}

// List returns all available templates
func (tm *TemplateManager) List() []*PolicyTemplate {
	templates := make([]*PolicyTemplate, 0, len(tm.templates))
	for _, tmpl := range tm.templates {
		templates = append(templates, tmpl)
	}
	return templates
}

// Get returns a template by name
func (tm *TemplateManager) Get(name string) (*PolicyTemplate, error) {
	tmpl, ok := tm.templates[name]
	if !ok {
		return nil, fmt.Errorf("template not found: %s", name)
	}
	return tmpl, nil
}

// Generate generates a policy from a template
func (tm *TemplateManager) Generate(name string, variables map[string]string) (string, error) {
	tmpl, err := tm.Get(name)
	if err != nil {
		return "", err
	}

	// Merge variables with defaults
	vars := make(map[string]string)
	for k, v := range tmpl.Variables {
		vars[k] = v
	}
	for k, v := range variables {
		vars[k] = v
	}

	// Parse and execute template
	t, err := template.New(name).Parse(tmpl.Template)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var result strings.Builder
	if err := t.Execute(&result, vars); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return result.String(), nil
}

// GenerateToFile generates a policy from a template and writes to file
func (tm *TemplateManager) GenerateToFile(name string, variables map[string]string, outputPath string) error {
	content, err := tm.Generate(name, variables)
	if err != nil {
		return err
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write to file
	if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// Search searches for templates by keyword
func (tm *TemplateManager) Search(keyword string) []*PolicyTemplate {
	var results []*PolicyTemplate
	keyword = strings.ToLower(keyword)
	
	for _, tmpl := range tm.templates {
		if strings.Contains(strings.ToLower(tmpl.Name), keyword) ||
			strings.Contains(strings.ToLower(tmpl.Description), keyword) ||
			strings.Contains(strings.ToLower(tmpl.Category), keyword) {
			results = append(results, tmpl)
		}
	}
	
	return results
}

// GetByCategory returns templates by category
func (tm *TemplateManager) GetByCategory(category string) []*PolicyTemplate {
	var results []*PolicyTemplate
	
	for _, tmpl := range tm.templates {
		if tmpl.Category == category {
			results = append(results, tmpl)
		}
	}
	
	return results
}

// GetCategories returns all unique categories
func (tm *TemplateManager) GetCategories() []string {
	categories := make(map[string]bool)
	
	for _, tmpl := range tm.templates {
		categories[tmpl.Category] = true
	}
	
	result := make([]string, 0, len(categories))
	for cat := range categories {
		result = append(result, cat)
	}
	
	return result
}