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
)

// integrationRunRequest is the JSON body for POST /api/v1/integrations/run.
// wasmPath is resolved under the server workspace root (same rules as other workspace paths).
// run must validate as omnigraph/integration-run/v1 when encoded to JSON.
type integrationRunRequest struct {
	WasmPath string          `json:"wasmPath"`
	Run      json.RawMessage `json:"run"`
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
	wasmAbs, err := resolveWorkspacePath(s.root, strings.TrimSpace(req.WasmPath))
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
	}, wasmAbs, runBytes, 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(out)
}

// RunIntegrationCLI runs a WASM integration plugin from the CLI (no Netbox/Zabbix imports).
func RunIntegrationCLI(ctx context.Context, wasmPath string, stdin []byte) ([]byte, error) {
	var env map[string]any
	if err := json.Unmarshal(stdin, &env); err != nil {
		return nil, fmt.Errorf("stdin json: %w", err)
	}
	spec, _ := env["spec"].(map[string]any)
	if spec == nil {
		return nil, fmt.Errorf("missing spec")
	}
	rawPrefixes, ok := spec["allowedFetchPrefixes"]
	if !ok {
		return nil, fmt.Errorf("missing allowedFetchPrefixes")
	}
	arr, ok := rawPrefixes.([]any)
	if !ok {
		return nil, fmt.Errorf("allowedFetchPrefixes must be array")
	}
	var prefixes []string
	for _, v := range arr {
		ps, ok := v.(string)
		if !ok || strings.TrimSpace(ps) == "" {
			return nil, fmt.Errorf("invalid prefix")
		}
		prefixes = append(prefixes, ps)
	}
	abs, err := filepath.Abs(wasmPath)
	if err != nil {
		return nil, err
	}
	return runner.RunIntegrationPlugin(ctx, runner.IntegrationHostConfig{
		AllowedFetchPrefixes: prefixes,
	}, abs, stdin, 0)
}
