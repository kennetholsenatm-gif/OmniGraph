package repo

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAggregateStateHosts(t *testing.T) {
	dir := t.TempDir()
	_ = os.MkdirAll(filepath.Join(dir, "env"), 0o755)
	stateJSON := []byte(`{"version":4,"values":{"outputs":{"pub":{"value":"203.0.113.5"}},"root_module":{"resources":[]}}}`)
	_ = os.WriteFile(filepath.Join(dir, "env", "terraform.tfstate"), stateJSON, 0o600)

	rows, errs, err := AggregateStateHosts(dir, 10, 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	if len(errs) > 0 {
		t.Fatalf("errs: %v", errs)
	}
	if len(rows) != 1 {
		t.Fatalf("rows: %+v", rows)
	}
	if rows[0].Name != "output.pub" || rows[0].AnsibleHost != "203.0.113.5" {
		t.Fatalf("row: %+v", rows[0])
	}
}
