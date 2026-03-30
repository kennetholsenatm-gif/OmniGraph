package serve

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/kennetholsenatm-gif/omnigraph/internal/inventory"
	"github.com/kennetholsenatm-gif/omnigraph/internal/repo"
)

const inventoryAPIVersion = "omnigraph/inventory-api/v1"

// workspaceInventoryResponse is the default JSON envelope for GET /api/v1/inventory.
type workspaceInventoryResponse struct {
	APIVersion string  `json:"apiVersion"`
	Kind       string  `json:"kind"`
	Metadata   invMeta `json:"metadata"`
	Spec       invSpec `json:"spec"`
}

type invMeta struct {
	Root        string `json:"root"`
	GeneratedAt string `json:"generatedAt"`
}

type invSpec struct {
	Hosts       []repo.StateHostRow `json:"hosts"`
	StateErrors []string            `json:"stateErrors,omitempty"`
}

func ansibleDynamicInventoryJSON(rows []repo.StateHostRow) map[string]any {
	hostvars := make(map[string]map[string]string)
	seen := make(map[string]struct{})
	for _, r := range rows {
		h := inventory.SanitizeHostKey(r.Name)
		if _, ok := seen[h]; ok {
			continue
		}
		seen[h] = struct{}{}
		hostvars[h] = map[string]string{"ansible_host": r.AnsibleHost}
	}
	hostnames := make([]string, 0, len(hostvars))
	for h := range hostvars {
		hostnames = append(hostnames, h)
	}
	sort.Strings(hostnames)
	return map[string]any{
		"omnigraph": map[string]any{
			"hosts": hostnames,
		},
		"_meta": map[string]any{
			"hostvars": hostvars,
		},
	}
}

func (s *server) getInventory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	root, err := resolveWorkspacePath(s.root, r.URL.Query().Get("path"))
	if err != nil {
		writeAPIErrorJSON(w, "INVENTORY_PATH_INVALID", "inventory: "+err.Error(), http.StatusBadRequest)
		return
	}
	rows, stateErrs, err := repo.AggregateStateHosts(root, 32, 0)
	if err != nil {
		writeAPIErrorJSON(w, "INVENTORY_AGGREGATE_FAILED", "inventory: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if s.audit != nil {
		detail := root
		if sub := auditSubjectDetail(r); sub != "" {
			detail = sub + " " + root
		}
		s.audit.Append("inventory_get", detail)
	}

	format := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("format")))
	if format == "" {
		format = "json"
	}

	switch format {
	case "ini":
		ini := MergedOmnigraphINI(rows)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte(ini))
		return
	case "ansible-json", "ansible":
		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		_ = enc.Encode(ansibleDynamicInventoryJSON(rows))
		return
	case "json":
		resp := workspaceInventoryResponse{
			APIVersion: inventoryAPIVersion,
			Kind:       "WorkspaceInventory",
			Metadata: invMeta{
				Root:        root,
				GeneratedAt: time.Now().UTC().Format(time.RFC3339),
			},
			Spec: invSpec{
				Hosts:       rows,
				StateErrors: stateErrs,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		if err := enc.Encode(resp); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	default:
		writeAPIErrorJSON(w, "INVENTORY_FORMAT_UNKNOWN", fmt.Sprintf("inventory: unknown format %q (use json, ini, ansible-json)", format), http.StatusBadRequest)
	}
}
