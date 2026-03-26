package policy

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// PolicyVersion represents a versioned policy
type PolicyVersion struct {
	Version     string            `json:"version"`
	Hash        string            `json:"hash"`
	Timestamp   time.Time         `json:"timestamp"`
	Author      string            `json:"author,omitempty"`
	Message     string            `json:"message,omitempty"`
	Changes     []string          `json:"changes,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// PolicyManifest represents a policy with version information
type PolicyManifest struct {
	APIVersion string                 `json:"apiVersion"`
	Kind       string                 `json:"kind"`
	Metadata   PolicyManifestMetadata `json:"metadata"`
	Spec       PolicyManifestSpec     `json:"spec"`
}

// PolicyManifestMetadata contains policy metadata
type PolicyManifestMetadata struct {
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Description string            `json:"description,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

// PolicyManifestSpec contains policy specification
type PolicyManifestSpec struct {
	Rego        string            `json:"rego"`
	Variables   map[string]string `json:"variables,omitempty"`
	Enforcement string            `json:"enforcement,omitempty"`
	Severity    string            `json:"severity,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
}

// VersionManager manages policy versions
type VersionManager struct {
	versionsDir string
	history     map[string][]*PolicyVersion
}

// NewVersionManager creates a new version manager
func NewVersionManager(versionsDir string) (*VersionManager, error) {
	// Create versions directory if it doesn't exist
	if err := os.MkdirAll(versionsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create versions directory: %w", err)
	}

	return &VersionManager{
		versionsDir: versionsDir,
		history:     make(map[string][]*PolicyVersion),
	}, nil
}

// CreateVersion creates a new version of a policy
func (vm *VersionManager) CreateVersion(policyPath string, author, message string) (*PolicyVersion, error) {
	// Read policy file
	content, err := os.ReadFile(policyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read policy: %w", err)
	}

	// Calculate hash
	hash := fmt.Sprintf("%x", sha256.Sum256(content))

	// Get existing versions
	policyName := filepath.Base(policyPath)
	versions := vm.history[policyName]

	// Check if hash already exists
	for _, v := range versions {
		if v.Hash == hash {
			return v, nil // Already versioned
		}
	}

	// Create new version
	version := &PolicyVersion{
		Version:   fmt.Sprintf("%d", len(versions)+1),
		Hash:      hash,
		Timestamp: time.Now(),
		Author:    author,
		Message:   message,
		Changes:   vm.detectChanges(policyName, content),
		Metadata:  make(map[string]string),
	}

	// Add to history
	vm.history[policyName] = append(versions, version)

	// Save version to disk
	if err := vm.saveVersion(policyName, version, content); err != nil {
		return nil, fmt.Errorf("failed to save version: %w", err)
	}

	return version, nil
}

// GetVersion returns a specific version of a policy
func (vm *VersionManager) GetVersion(policyName, version string) (*PolicyVersion, []byte, error) {
	versions := vm.history[policyName]
	if versions == nil {
		return nil, nil, fmt.Errorf("no versions found for policy: %s", policyName)
	}

	// Find version
	for _, v := range versions {
		if v.Version == version {
			// Load content from disk
			content, err := vm.loadVersion(policyName, v)
			if err != nil {
				return nil, nil, err
			}
			return v, content, nil
		}
	}

	return nil, nil, fmt.Errorf("version %s not found for policy %s", version, policyName)
}

// ListVersions returns all versions of a policy
func (vm *VersionManager) ListVersions(policyName string) []*PolicyVersion {
	return vm.history[policyName]
}

// GetLatestVersion returns the latest version of a policy
func (vm *VersionManager) GetLatestVersion(policyName string) (*PolicyVersion, error) {
	versions := vm.history[policyName]
	if len(versions) == 0 {
		return nil, fmt.Errorf("no versions found for policy: %s", policyName)
	}
	return versions[len(versions)-1], nil
}

// CompareVersions compares two versions of a policy
func (vm *VersionManager) CompareVersions(policyName, version1, version2 string) (*VersionDiff, error) {
	v1, content1, err := vm.GetVersion(policyName, version1)
	if err != nil {
		return nil, err
	}

	v2, content2, err := vm.GetVersion(policyName, version2)
	if err != nil {
		return nil, err
	}

	diff := &VersionDiff{
		PolicyName: policyName,
		Version1:   v1,
		Version2:   v2,
		Changes:    vm.computeDiff(content1, content2),
	}

	return diff, nil
}

// VersionDiff represents differences between versions
type VersionDiff struct {
	PolicyName string
	Version1   *PolicyVersion
	Version2   *PolicyVersion
	Changes    []DiffChange
}

// DiffChange represents a single change
type DiffChange struct {
	Type    string // added, removed, modified
	Line    int
	Content string
}

// computeDiff computes differences between two policy contents
func (vm *VersionManager) computeDiff(content1, content2 []byte) []DiffChange {
	lines1 := splitLines(string(content1))
	lines2 := splitLines(string(content2))

	var changes []DiffChange
	maxLen := len(lines1)
	if len(lines2) > maxLen {
		maxLen = len(lines2)
	}

	for i := 0; i < maxLen; i++ {
		var line1, line2 string
		if i < len(lines1) {
			line1 = lines1[i]
		}
		if i < len(lines2) {
			line2 = lines2[i]
		}

		if line1 != line2 {
			if line1 == "" {
				changes = append(changes, DiffChange{
					Type:    "added",
					Line:    i + 1,
					Content: line2,
				})
			} else if line2 == "" {
				changes = append(changes, DiffChange{
					Type:    "removed",
					Line:    i + 1,
					Content: line1,
				})
			} else {
				changes = append(changes, DiffChange{
					Type:    "modified",
					Line:    i + 1,
					Content: line2,
				})
			}
		}
	}

	return changes
}

// detectChanges detects changes from previous version
func (vm *VersionManager) detectChanges(policyName string, content []byte) []string {
	versions := vm.history[policyName]
	if len(versions) == 0 {
		return []string{"Initial version"}
	}

	latest := versions[len(versions)-1]
	latestContent, err := vm.loadVersion(policyName, latest)
	if err != nil {
		return []string{"Unable to detect changes"}
	}

	diff := vm.computeDiff(latestContent, content)
	var changes []string
	for _, d := range diff {
		changes = append(changes, fmt.Sprintf("%s at line %d", d.Type, d.Line))
	}

	return changes
}

// saveVersion saves a version to disk
func (vm *VersionManager) saveVersion(policyName string, version *PolicyVersion, content []byte) error {
	// Create policy directory
	policyDir := filepath.Join(vm.versionsDir, policyName)
	if err := os.MkdirAll(policyDir, 0755); err != nil {
		return err
	}

	// Save version metadata
	metaPath := filepath.Join(policyDir, fmt.Sprintf("%s.json", version.Version))
	metaData, err := json.MarshalIndent(version, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(metaPath, metaData, 0644); err != nil {
		return err
	}

	// Save policy content
	contentPath := filepath.Join(policyDir, fmt.Sprintf("%s.rego", version.Version))
	if err := os.WriteFile(contentPath, content, 0644); err != nil {
		return err
	}

	return nil
}

// loadVersion loads a version from disk
func (vm *VersionManager) loadVersion(policyName string, version *PolicyVersion) ([]byte, error) {
	contentPath := filepath.Join(vm.versionsDir, policyName, fmt.Sprintf("%s.rego", version.Version))
	return os.ReadFile(contentPath)
}

// splitLines splits content into lines
func splitLines(content string) []string {
	var lines []string
	current := ""
	for _, c := range content {
		if c == '\n' {
			lines = append(lines, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

// ExportVersion exports a specific version
func (vm *VersionManager) ExportVersion(policyName, version, outputPath string) error {
	v, content, err := vm.GetVersion(policyName, version)
	if err != nil {
		return err
	}

	// Create manifest
	manifest := PolicyManifest{
		APIVersion: "omnigraph/policy/v1",
		Kind:       "Policy",
		Metadata: PolicyManifestMetadata{
			Name:    policyName,
			Version: v.Version,
		},
		Spec: PolicyManifestSpec{
			Rego: string(content),
		},
	}

	// Write manifest
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(outputPath, data, 0644)
}

// ImportVersion imports a policy version from manifest
func (vm *VersionManager) ImportVersion(manifestPath string) (*PolicyVersion, error) {
	// Read manifest
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, err
	}

	var manifest PolicyManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, err
	}

	// Create version
	version := &PolicyVersion{
		Version:   manifest.Metadata.Version,
		Hash:      fmt.Sprintf("%x", sha256.Sum256([]byte(manifest.Spec.Rego))),
		Timestamp: time.Now(),
		Metadata:  manifest.Metadata.Labels,
	}

	// Add to history
	policyName := manifest.Metadata.Name
	vm.history[policyName] = append(vm.history[policyName], version)

	// Save to disk
	if err := vm.saveVersion(policyName, version, []byte(manifest.Spec.Rego)); err != nil {
		return nil, err
	}

	return version, nil
}