package serve

import (
	"github.com/kennetholsenatm-gif/omnigraph/internal/inventory"
	"github.com/kennetholsenatm-gif/omnigraph/internal/repo"
)

// MergedOmnigraphINI builds a single [omnigraph] INI from state-derived rows (first wins per sanitized key).
func MergedOmnigraphINI(rows []repo.StateHostRow) string {
	out := make(map[string]string)
	seen := make(map[string]struct{})
	for _, r := range rows {
		sk := inventory.SanitizeHostKey(r.Name)
		if _, ok := seen[sk]; ok {
			continue
		}
		seen[sk] = struct{}{}
		out[r.Name] = r.AnsibleHost
	}
	return inventory.BuildINI(out)
}
