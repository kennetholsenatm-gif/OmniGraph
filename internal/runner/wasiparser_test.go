package runner

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestRunWASIParserAnsibleINI(t *testing.T) {
	if testing.Short() {
		t.Skip("builds wasip1 wasm")
	}
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("caller")
	}
	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))

	wasmOut := filepath.Join(t.TempDir(), "ansible-ini.wasm")
	cmd := exec.Command("go", "build", "-o", wasmOut, "./wasm/plugins/ansibleini")
	cmd.Dir = repoRoot
	cmd.Env = append(filterEnv(), "GOOS=wasip1", "GOARCH=wasm")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go build wasm: %v\n%s", err, out)
	}

	ctx := context.Background()
	stdin := []byte("[web]\nhost-a\n")
	got, err := RunWASIParserLimit(ctx, wasmOut, stdin, 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		APIVersion string `json:"apiVersion"`
		Kind       string `json:"kind"`
	}
	if err := json.Unmarshal(got, &doc); err != nil {
		t.Fatalf("json: %v body %s", err, got)
	}
	if doc.APIVersion != "omnigraph/graph/v1" || doc.Kind != "Graph" {
		t.Fatalf("unexpected %+v", doc)
	}
}

func filterEnv() []string {
	var out []string
	for _, e := range os.Environ() {
		if strings.HasPrefix(e, "GOOS=") || strings.HasPrefix(e, "GOARCH=") {
			continue
		}
		out = append(out, e)
	}
	return out
}
