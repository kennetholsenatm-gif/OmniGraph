package repo

import (
	"io/fs"
	"path/filepath"
	"strings"
)

// FileKind classifies paths under a repository root for infrastructure managers.
type FileKind string

const (
	KindTerraformState   FileKind = "terraform-state"
	KindTerraformHCL     FileKind = "terraform-hcl"
	KindOmnigraphSchema  FileKind = "omnigraph-schema"
	KindAnsibleCfg       FileKind = "ansible-cfg"
	KindAnsiblePlaybook  FileKind = "ansible-playbook"
	KindAnsibleInventory FileKind = "ansible-inventory"
	KindTerraformPlanBin FileKind = "terraform-plan-binary"
)

// Discovered is one artifact found while walking a checkout.
type Discovered struct {
	Path string   `json:"path"` // relative to scan root, slash-separated
	Kind FileKind `json:"kind"`
}

// Result is the full scan of a repository working tree.
type Result struct {
	Root  string       `json:"root"`
	Files []Discovered `json:"files"`
}

func skipDir(name string) bool {
	switch strings.ToLower(name) {
	case ".git", "node_modules", ".terraform", ".venv", "venv", "__pycache__", "vendor",
		"dist", "build", "target", ".idea", ".vscode":
		return true
	default:
		return false
	}
}

// Discover walks root (typically a Git working tree) and records IaC-related paths.
func Discover(root string) (*Result, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	var files []Discovered
	_ = filepath.WalkDir(abs, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if path != abs && skipDir(d.Name()) {
				return fs.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(abs, path)
		if err != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		name := strings.ToLower(d.Name())
		switch {
		case strings.HasSuffix(name, ".tfstate"):
			files = append(files, Discovered{Path: rel, Kind: KindTerraformState})
		case name == ".omnigraph.schema" || strings.HasSuffix(name, ".omnigraph.schema"):
			files = append(files, Discovered{Path: rel, Kind: KindOmnigraphSchema})
		case strings.HasSuffix(name, ".tf") || strings.HasSuffix(name, ".tofu"):
			files = append(files, Discovered{Path: rel, Kind: KindTerraformHCL})
		case name == "ansible.cfg":
			files = append(files, Discovered{Path: rel, Kind: KindAnsibleCfg})
		case name == "site.yml" || name == "site.yaml" || name == "playbook.yml" || name == "playbook.yaml":
			files = append(files, Discovered{Path: rel, Kind: KindAnsiblePlaybook})
		case name == "hosts" || strings.HasSuffix(name, ".ini"):
			if strings.Contains(strings.ToLower(rel), "inventory") || name == "hosts" {
				files = append(files, Discovered{Path: rel, Kind: KindAnsibleInventory})
			}
		case name == "tfplan" || strings.HasSuffix(name, ".tfplan"):
			files = append(files, Discovered{Path: rel, Kind: KindTerraformPlanBin})
		}
		return nil
	})
	return &Result{Root: abs, Files: files}, nil
}
