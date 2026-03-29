package repo

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/kennetholsenatm-gif/omnigraph/internal/inventory"
)

const manifestMaxFileBytes = 6 << 20 // 6 MiB per file (matches aggregate default)

func readManifestFile(path string) ([]byte, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if fi.Size() > manifestMaxFileBytes {
		return nil, fmt.Errorf("file too large (%d bytes, max %d)", fi.Size(), manifestMaxFileBytes)
	}
	return os.ReadFile(path)
}

func fileUnderRepoRoot(root, file string) error {
	root = filepath.Clean(root)
	file = filepath.Clean(file)
	rel, err := filepath.Rel(root, file)
	if err != nil {
		return err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return fmt.Errorf("path escapes repository root")
	}
	return nil
}

// RoutingPolicy mirrors the Python graph_manifest RoutingPolicy JSON.
type RoutingPolicy struct {
	PolicyRef         string `json:"policy_ref"`
	RequireEncryption bool   `json:"require_encryption"`
}

// DiscoveryProvenance mirrors Python DiscoveryProvenance.
type DiscoveryProvenance struct {
	SourceKind  string         `json:"source_kind"`
	SourcePaths []string       `json:"source_paths"`
	Hints       map[string]any `json:"hints,omitempty"`
}

// GraphManifest is connector-style discovery output (Python omnigraph-agent parity).
type GraphManifest struct {
	GraphID            string              `json:"graph_id"`
	NodeID             string              `json:"node_id"`
	ArtifactsDir       string              `json:"artifacts_dir"`
	Inputs             []string            `json:"inputs"`
	Outputs            []string            `json:"outputs"`
	RoutingPolicy      RoutingPolicy       `json:"routing_policy"`
	BackendConstraints map[string]any      `json:"backend_constraints,omitempty"`
	ExecutionPlanRef   *string             `json:"execution_plan_ref,omitempty"`
	Discovery          *DiscoveryProvenance `json:"discovery,omitempty"`
}

// ManifestDiscoveryResult is scan output including warnings (Python DiscoveryResult).
type ManifestDiscoveryResult struct {
	Manifests []GraphManifest `json:"manifests"`
	Warnings  []string        `json:"warnings,omitempty"`
}

// DiscoverManifestOptions configures deep manifest scanning (Python AutoDiscoverAgent).
type DiscoverManifestOptions struct {
	MaxDepth    int
	IgnoreGlobs []string
}

// DefaultManifestDiscoverOptions matches Python defaults (max_depth=6).
func DefaultManifestDiscoverOptions() DiscoverManifestOptions {
	return DiscoverManifestOptions{
		MaxDepth:    6,
		IgnoreGlobs: []string{".git*", "__pycache__", "*.pyc", ".terraform"},
	}
}

func stableID(parts ...string) string {
	h := sha256.New()
	for _, p := range parts {
		h.Write([]byte(p))
		h.Write([]byte{0x1e})
	}
	sum := h.Sum(nil)
	return hex.EncodeToString(sum)[:14]
}

func artifactsDirFor(absFile string) string {
	dir := filepath.Dir(absFile)
	base := strings.TrimSuffix(filepath.Base(absFile), filepath.Ext(absFile))
	if ext := filepath.Ext(absFile); strings.EqualFold(ext, ".tfstate") {
		base = strings.TrimSuffix(filepath.Base(absFile), ext)
	}
	return filepath.Join(dir, ".omnigraph", base)
}

// DiscoverManifests walks scan roots like the Python agent: Terraform state, Ansible INI, YAML inventories.
func DiscoverManifests(roots []string, opt DiscoverManifestOptions) (*ManifestDiscoveryResult, error) {
	if opt.MaxDepth <= 0 {
		opt.MaxDepth = 6
	}
	var manifests []GraphManifest
	var warnings []string
	seenTF := map[string]struct{}{}
	seenInv := map[string]struct{}{}

	for _, root := range roots {
		absRoot, err := filepath.Abs(root)
		if err != nil {
			return nil, err
		}
		st, err := os.Stat(absRoot)
		if err != nil || !st.IsDir() {
			continue
		}
		_ = filepath.WalkDir(absRoot, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() {
				if path != absRoot && skipDir(d.Name()) {
					return fs.SkipDir
				}
				return nil
			}
			if pathInBannedDir(path, absRoot) {
				return nil
			}
			if ignoredName(d.Name(), opt.IgnoreGlobs) {
				return nil
			}
			depth := fileDepth(absRoot, path)
			if depth > opt.MaxDepth {
				return nil
			}

			name := d.Name()
			lower := strings.ToLower(name)

			switch {
			case name == "terraform.tfstate" || strings.HasSuffix(lower, ".tfstate"):
				key, _ := filepath.Abs(path)
				if _, ok := seenTF[key]; ok {
					return nil
				}
				addrs, w := parseTerraformStateFile(path)
				warnings = append(warnings, w...)
				if len(addrs) == 0 {
					return nil
				}
				seenTF[key] = struct{}{}
				m, w2 := graphManifestTerraform(path, absRoot, addrs)
				warnings = append(warnings, w2...)
				if m != nil {
					manifests = append(manifests, *m)
				}

			case name == "inventory" || name == "hosts" || strings.HasSuffix(lower, ".ini"):
				key, _ := filepath.Abs(path)
				if _, ok := seenInv[key]; ok {
					return nil
				}
				hosts, w := parseINIInventoryFile(path)
				warnings = append(warnings, w...)
				if len(hosts) == 0 {
					return nil
				}
				seenInv[key] = struct{}{}
				if m, ok := graphManifestAnsible(path, absRoot, hosts); ok {
					manifests = append(manifests, m)
				}

			case strings.HasSuffix(lower, ".yml") || strings.HasSuffix(lower, ".yaml"):
				if !strings.Contains(lower, "inventory") && !strings.Contains(lower, "hosts") {
					return nil
				}
				key, _ := filepath.Abs(path)
				if _, ok := seenInv[key]; ok {
					return nil
				}
				hosts, w := parseYAMLInventoryFile(path)
				warnings = append(warnings, w...)
				if len(hosts) == 0 {
					return nil
				}
				seenInv[key] = struct{}{}
				if m, ok := graphManifestAnsible(path, absRoot, hosts); ok {
					manifests = append(manifests, m)
				}
			}
			return nil
		})
	}

	return &ManifestDiscoveryResult{Manifests: manifests, Warnings: warnings}, nil
}

func fileDepth(rootAbs, fileAbs string) int {
	rel, err := filepath.Rel(rootAbs, fileAbs)
	if err != nil {
		return 999
	}
	if rel == "." {
		return 0
	}
	return strings.Count(rel, string(filepath.Separator))
}

func pathInBannedDir(path, root string) bool {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return true
	}
	for _, p := range strings.Split(rel, string(filepath.Separator)) {
		switch strings.ToLower(p) {
		case ".git", ".terraform", "__pycache__", "node_modules":
			return true
		}
	}
	return false
}

func ignoredName(name string, globs []string) bool {
	for _, g := range globs {
		g = strings.TrimSpace(g)
		if g == "" {
			continue
		}
		if ok, _ := filepath.Match(g, name); ok {
			return true
		}
	}
	return false
}

func parseTerraformStateFile(path string) ([]string, []string) {
	var w []string
	b, err := readManifestFile(path)
	if err != nil {
		return nil, []string{fmt.Sprintf("%s: %v", path, err)}
	}
	addrs, err := TerraformResourceAddresses(b)
	if err != nil {
		return nil, append(w, fmt.Sprintf("%s: %v", path, err))
	}
	return addrs, w
}

func graphManifestTerraform(path, rootAbs string, addresses []string) (*GraphManifest, []string) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, []string{err.Error()}
	}
	if err := fileUnderRepoRoot(rootAbs, abs); err != nil {
		return nil, []string{fmt.Sprintf("%s: %v", path, err)}
	}
	if len(addresses) == 0 {
		return nil, nil
	}
	gid := stableID("tf", abs)
	primary := addresses[0]
	return &GraphManifest{
		GraphID:       "tf-" + gid,
		NodeID:        primary,
		ArtifactsDir:  artifactsDirFor(abs),
		Inputs:        addresses,
		Outputs:       nil,
		RoutingPolicy: RoutingPolicy{PolicyRef: "terraform", RequireEncryption: true},
		BackendConstraints: map[string]any{
			"terraform_state_path": abs,
		},
		Discovery: &DiscoveryProvenance{
			SourceKind:  "terraform_state",
			SourcePaths: []string{abs},
			Hints:       map[string]any{"resource_count": len(addresses)},
		},
	}, nil
}

func parseINIInventoryFile(path string) ([]string, []string) {
	b, err := readManifestFile(path)
	if err != nil {
		return nil, []string{fmt.Sprintf("%s: %v", path, err)}
	}
	hosts, err := inventory.ParseAnsibleINIHosts(b)
	if err != nil {
		return nil, []string{fmt.Sprintf("%s: %v", path, err)}
	}
	return hosts, nil
}

func parseYAMLInventoryFile(path string) ([]string, []string) {
	b, err := readManifestFile(path)
	if err != nil {
		return nil, []string{fmt.Sprintf("%s: %v", path, err)}
	}
	hosts, err := inventory.ParseAnsibleYAMLHosts(b)
	if err != nil {
		return nil, []string{fmt.Sprintf("%s: %v", path, err)}
	}
	return hosts, nil
}

func graphManifestAnsible(path, rootAbs string, hosts []string) (GraphManifest, bool) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return GraphManifest{}, false
	}
	if err := fileUnderRepoRoot(rootAbs, abs); err != nil {
		return GraphManifest{}, false
	}
	gid := stableID("ansible", abs)
	primary := "ansible-root"
	if len(hosts) > 0 {
		primary = hosts[0]
	}
	return GraphManifest{
		GraphID:       "ansible-" + gid,
		NodeID:        primary,
		ArtifactsDir:  artifactsDirFor(abs),
		Inputs:        hosts,
		Outputs:       nil,
		RoutingPolicy: RoutingPolicy{PolicyRef: "ansible", RequireEncryption: true},
		BackendConstraints: map[string]any{
			"ansible_inventory_path": abs,
		},
		Discovery: &DiscoveryProvenance{
			SourceKind:  "ansible_inventory",
			SourcePaths: []string{abs},
			Hints:       map[string]any{"host_count": len(hosts)},
		},
	}, true
}

// GraphDocumentV1 is a minimal omnigraph/graph/v1 document for plugin output validation.
type GraphDocumentV1 struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Metadata   struct {
		GeneratedAt string `json:"generatedAt"`
	} `json:"metadata"`
	Spec struct {
		Nodes []struct {
			ID    string `json:"id"`
			Kind  string `json:"kind"`
			Label string `json:"label"`
		} `json:"nodes"`
		Edges []struct {
			From string `json:"from"`
			To   string `json:"to"`
		} `json:"edges,omitempty"`
	} `json:"spec"`
}

// GraphV1FromAnsibleHosts builds a graph/v1 JSON document from inventory host names.
func GraphV1FromAnsibleHosts(generatedAt string, hosts []string) GraphDocumentV1 {
	var doc GraphDocumentV1
	doc.APIVersion = "omnigraph/graph/v1"
	doc.Kind = "Graph"
	doc.Metadata.GeneratedAt = generatedAt
	for _, h := range hosts {
		doc.Spec.Nodes = append(doc.Spec.Nodes, struct {
			ID    string `json:"id"`
			Kind  string `json:"kind"`
			Label string `json:"label"`
		}{ID: "host:" + h, Kind: "host", Label: h})
	}
	return doc
}
