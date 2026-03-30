package state

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestExtractHosts(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("testdata", "minimal.state.json"))
	if err != nil {
		t.Fatal(err)
	}
	st, err := Parse(b)
	if err != nil {
		t.Fatal(err)
	}
	got := ExtractHosts(st)
	want := map[string]string{
		"output.bastion_ip": "203.0.113.10",
		"aws_instance.web":  "10.0.5.21",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("hosts mismatch\n got: %#v\nwant: %#v", got, want)
	}
}

func TestExecutionScopeAcquireLock(t *testing.T) {
	dir := t.TempDir()
	allowed := filepath.Join(dir, "terraform.tfstate")
	other := filepath.Join(dir, "other.tfstate")
	for _, p := range []string{allowed, other} {
		if err := os.WriteFile(p, []byte("{}"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	m, err := NewManager(dir)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	m.SetExecutionScope(NewExecutionScopeForBlastRadius(allowed))
	if _, err := m.AcquireLock(ctx, allowed, "t", "op", time.Minute); err != nil {
		t.Fatal(err)
	}
	if _, err := m.AcquireLock(ctx, other, "t", "op2", time.Minute); err == nil {
		t.Fatal("expected error for path outside execution scope")
	}
}
