package repo

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kennetholsenatm-gif/omnigraph/internal/safepath"
	"github.com/kennetholsenatm-gif/omnigraph/internal/state"
)

const defaultMaxStateBytes = 6 << 20 // 6 MiB per state file

// StateHostRow is one host extracted from Terraform/OpenTofu JSON state under a repo root.
type StateHostRow struct {
	Name        string `json:"name"`
	AnsibleHost string `json:"ansibleHost"`
	Origin      string `json:"origin"` // relative to scan root, slash-separated
}

// AggregateStateHosts discovers state files under root and extracts host rows from each.
// maxFiles caps how many .tfstate files are read; maxBytes caps each file size.
func AggregateStateHosts(root string, maxFiles int, maxBytes int64) ([]StateHostRow, []string, error) {
	if maxFiles <= 0 {
		maxFiles = 32
	}
	if maxBytes <= 0 {
		maxBytes = defaultMaxStateBytes
	}
	res, err := Discover(root)
	if err != nil {
		return nil, nil, err
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, nil, err
	}
	var rows []StateHostRow
	var errs []string
	n := 0
	for _, f := range res.Files {
		if f.Kind != KindTerraformState {
			continue
		}
		if n >= maxFiles {
			break
		}
		full, pathErr := safepath.UnderRoot(absRoot, f.Path)
		if pathErr != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", f.Path, pathErr))
			continue
		}
		b, rerr := readFileLimited(full, maxBytes)
		if rerr != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", f.Path, rerr))
			continue
		}
		st, perr := state.Parse(b)
		if perr != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", f.Path, perr))
			continue
		}
		for name, host := range state.ExtractHosts(st) {
			rows = append(rows, StateHostRow{
				Name:        name,
				AnsibleHost: host,
				Origin:      f.Path,
			})
		}
		n++
	}
	return rows, errs, nil
}

func readFileLimited(path string, maxBytes int64) ([]byte, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if fi.Size() > maxBytes {
		return nil, fmt.Errorf("file too large (%d bytes, max %d)", fi.Size(), maxBytes)
	}
	return os.ReadFile(path)
}
