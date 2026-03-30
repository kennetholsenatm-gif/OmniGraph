package serve

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/kennetholsenatm-gif/omnigraph/internal/runner"
	"github.com/kennetholsenatm-gif/omnigraph/internal/safepath"
)

// integrationRunRequest is the JSON body for POST /api/v1/integrations/run.
// wasmPath must be relative to the server workspace root (no absolute paths; prevents path escape).
// run must validate as omnigraph/integration-run/v1 when encoded to JSON.
type integrationRunRequest struct {
	WasmPath string          `json:"wasmPath"`
	Run      json.RawMessage `json:"run"`
}

const integrationStdioErrPrefix = "serve:integration:stdio"

func integrationStdioErrf(format string, args ...any) error {
	return fmt.Errorf("%s: %s", integrationStdioErrPrefix, fmt.Sprintf(format, args...))
}

func (s *server) postIntegrationRun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, s.maxIngestBodyBytes))
	if err != nil {
		http.Error(w, "read body", http.StatusBadRequest)
		return
	}
	var req integrationRunRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	base := strings.TrimSpace(s.root)
	if base == "" {
		wd, err := os.Getwd()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		base = wd
	}
	absBase, err := filepath.Abs(filepath.Clean(base))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	rel := strings.TrimSpace(req.WasmPath)
	if rel == "" {
		http.Error(w, "wasmPath required", http.StatusBadRequest)
		return
	}
	if filepath.IsAbs(rel) {
		http.Error(w, "wasmPath must be relative to workspace root", http.StatusBadRequest)
		return
	}
	wasmRelSlash := filepath.ToSlash(rel)
	wasmAbs, err := safepath.UnderRoot(absBase, wasmRelSlash)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if st, err := os.Stat(wasmAbs); err != nil || st.IsDir() {
		http.Error(w, "wasmPath not found", http.StatusBadRequest)
		return
	}
	runBytes := req.Run
	if len(runBytes) == 0 {
		http.Error(w, "run required", http.StatusBadRequest)
		return
	}
	var env map[string]any
	if err := json.Unmarshal(runBytes, &env); err != nil {
		http.Error(w, "run must be JSON", http.StatusBadRequest)
		return
	}
	spec, _ := env["spec"].(map[string]any)
	if spec == nil {
		http.Error(w, "run.spec missing", http.StatusBadRequest)
		return
	}
	rawPrefixes, ok := spec["allowedFetchPrefixes"]
	if !ok {
		http.Error(w, "run.spec.allowedFetchPrefixes missing", http.StatusBadRequest)
		return
	}
	arr, ok := rawPrefixes.([]any)
	if !ok {
		http.Error(w, "allowedFetchPrefixes must be array", http.StatusBadRequest)
		return
	}
	var prefixes []string
	for _, v := range arr {
		ps, ok := v.(string)
		if !ok || strings.TrimSpace(ps) == "" {
			http.Error(w, "invalid prefix", http.StatusBadRequest)
			return
		}
		prefixes = append(prefixes, ps)
	}

	ctx := r.Context()
	out, err := runner.RunIntegrationPlugin(ctx, runner.IntegrationHostConfig{
		AllowedFetchPrefixes: prefixes,
		WasmModuleRoot:       absBase,
		WasmModuleRel:        wasmRelSlash,
	}, runBytes, 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(out)
}

// RunIntegrationStdio runs a WASM integration plugin with integration-run/v1 bytes on stdin (no Netbox/Zabbix imports in core).
func RunIntegrationStdio(ctx context.Context, wasmPath string, stdin []byte) ([]byte, error) {
	var env map[string]any
	if err := json.Unmarshal(stdin, &env); err != nil {
		return nil, integrationStdioErrf("stdin json: %v", err)
	}
	spec, _ := env["spec"].(map[string]any)
	if spec == nil {
		return nil, integrationStdioErrf("missing spec")
	}
	rawPrefixes, ok := spec["allowedFetchPrefixes"]
	if !ok {
		return nil, integrationStdioErrf("missing allowedFetchPrefixes")
	}
	arr, ok := rawPrefixes.([]any)
	if !ok {
		return nil, integrationStdioErrf("allowedFetchPrefixes must be array")
	}
	var prefixes []string
	for _, v := range arr {
		ps, ok := v.(string)
		if !ok || strings.TrimSpace(ps) == "" {
			return nil, integrationStdioErrf("invalid prefix")
		}
		prefixes = append(prefixes, ps)
	}
	wasmClean := strings.TrimSpace(wasmPath)
	if filepath.IsAbs(wasmClean) {
		return nil, integrationStdioErrf("--wasm must be relative to the current working directory")
	}
	wd, err := os.Getwd()
	if err != nil {
		return nil, integrationStdioErrf("resolve cwd: %v", err)
	}
	rootAbs, err := filepath.Abs(wd)
	if err != nil {
		return nil, integrationStdioErrf("resolve absolute cwd: %v", err)
	}
	return runner.RunIntegrationPlugin(ctx, runner.IntegrationHostConfig{
		AllowedFetchPrefixes: prefixes,
		WasmModuleRoot:       rootAbs,
		WasmModuleRel:        filepath.ToSlash(wasmClean),
	}, stdin, 0)
}
