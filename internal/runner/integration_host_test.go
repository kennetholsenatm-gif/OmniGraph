package runner

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestRunIntegrationPlugin_rejectsWasmPathEscape(t *testing.T) {
	tmp := t.TempDir()
	stdin := []byte(`{
		"apiVersion": "omnigraph/integration-run/v1",
		"kind": "IntegrationRun",
		"spec": {
			"plugin": "netbox",
			"allowedFetchPrefixes": ["http://127.0.0.1:9/"]
		}
	}`)
	_, err := RunIntegrationPlugin(context.Background(), IntegrationHostConfig{
		AllowedFetchPrefixes: []string{"http://127.0.0.1:9/"},
		WasmModuleRoot:       tmp,
		WasmModuleRel:        "../outside.wasm",
	}, stdin, 1024)
	if err == nil {
		t.Fatal("expected error for path escape")
	}
	if !strings.Contains(err.Error(), "escapes") {
		t.Fatalf("expected escape error, got %v", err)
	}
}

func TestURLAllowed(t *testing.T) {
	prefixes := []string{"http://127.0.0.1:9999/"}
	if !urlAllowed("http://127.0.0.1:9999/api/x", prefixes) {
		t.Fatal("expected allowed")
	}
	if urlAllowed("http://evil.example/api", prefixes) {
		t.Fatal("expected denied")
	}
	if urlAllowed("ftp://127.0.0.1:9999/api", prefixes) {
		t.Fatal("expected denied scheme")
	}
}

func TestRunIntegrationPlugin_Netbox(t *testing.T) {
	if testing.Short() {
		t.Skip("builds wasip1 wasm")
	}
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("caller")
	}
	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/dcim/devices/" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"results": []map[string]any{
				{
					"id":   42,
					"name": "edge-router",
					"primary_ip": map[string]any{
						"address": "10.0.0.1/24",
					},
				},
			},
		})
	}))
	defer srv.Close()
	prefix := srv.URL + "/"

	tmpDir := t.TempDir()
	wasmOut := filepath.Join(tmpDir, "netbox.wasm")
	cmd := exec.Command("go", "build", "-o", wasmOut, "./wasm/plugins/netbox")
	cmd.Dir = repoRoot
	cmd.Env = append(filterEnv(), "GOOS=wasip1", "GOARCH=wasm")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go build wasm: %v\n%s", err, out)
	}

	stdin := map[string]any{
		"apiVersion": "omnigraph/integration-run/v1",
		"kind":       "IntegrationRun",
		"spec": map[string]any{
			"plugin":               "netbox",
			"allowedFetchPrefixes": []any{prefix},
			"credentials":          map[string]any{"token": "dummy"},
		},
	}
	stdinJ, err := json.Marshal(stdin)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	got, err := RunIntegrationPlugin(ctx, IntegrationHostConfig{
		AllowedFetchPrefixes: []string{prefix},
		WasmModuleRoot:       tmpDir,
		WasmModuleRel:        "netbox.wasm",
	}, stdinJ, 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(got, &doc); err != nil {
		t.Fatal(err)
	}
	spec := doc["spec"].(map[string]any)
	if spec["status"] != "ok" {
		t.Fatalf("status: %+v", doc)
	}
}

func TestRunIntegrationPlugin_Zabbix(t *testing.T) {
	if testing.Short() {
		t.Skip("builds wasip1 wasm")
	}
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("caller")
	}
	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api_jsonrpc.php" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"jsonrpc": "2.0",
			"result": []map[string]any{
				{"hostid": "100", "host": "web-01"},
			},
			"id": 1,
		})
	}))
	defer srv.Close()
	prefix := srv.URL + "/"

	tmpDir := t.TempDir()
	wasmOut := filepath.Join(tmpDir, "zabbix.wasm")
	cmd := exec.Command("go", "build", "-o", wasmOut, "./wasm/plugins/zabbix")
	cmd.Dir = repoRoot
	cmd.Env = append(filterEnv(), "GOOS=wasip1", "GOARCH=wasm")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go build wasm: %v\n%s", err, out)
	}

	stdin := map[string]any{
		"apiVersion": "omnigraph/integration-run/v1",
		"kind":       "IntegrationRun",
		"spec": map[string]any{
			"plugin":               "zabbix",
			"allowedFetchPrefixes": []any{prefix},
			"credentials":          map[string]any{"token": "apitoken"},
		},
	}
	stdinJ, err := json.Marshal(stdin)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	got, err := RunIntegrationPlugin(ctx, IntegrationHostConfig{
		AllowedFetchPrefixes: []string{prefix},
		WasmModuleRoot:       tmpDir,
		WasmModuleRel:        "zabbix.wasm",
	}, stdinJ, 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "web-01") {
		t.Fatalf("unexpected body %s", got)
	}
}
