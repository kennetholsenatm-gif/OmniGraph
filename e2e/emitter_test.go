package e2e

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kennetholsenatm-gif/omnigraph/pkg/emitter"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(dir) != "e2e" {
		t.Fatalf("expected test cwd e2e, got %s", dir)
	}
	return filepath.Clean(filepath.Join(dir, ".."))
}

func TestIR_Validate_MinimalFixture(t *testing.T) {
	fix := filepath.Join(repoRoot(t), "e2e", "fixtures", "minimal.ir.json")
	raw, err := os.ReadFile(fix)
	if err != nil {
		t.Fatal(err)
	}
	doc, err := emitter.ParseDocument(raw)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Metadata.Name != "e2e-fixture" {
		t.Fatalf("unexpected metadata.name: %q", doc.Metadata.Name)
	}
}

func TestIR_Emit_AnsibleInventory(t *testing.T) {
	fix := filepath.Join(repoRoot(t), "e2e", "fixtures", "minimal.ir.json")
	raw, err := os.ReadFile(fix)
	if err != nil {
		t.Fatal(err)
	}
	doc, err := emitter.ParseDocument(raw)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	arts, err := emitter.DefaultRegistry().Emit(ctx, "ansible-inventory-ini", doc)
	if err != nil {
		t.Fatal(err)
	}
	if len(arts) == 0 {
		t.Fatal("emitter returned no artifacts")
	}
	body := string(arts[0].Content)
	if !strings.Contains(body, "ansible_host=10.0.0.99") {
		t.Fatalf("expected inventory line in output: %q", body)
	}
}

func TestIR_Validate_InvalidFixtureFails(t *testing.T) {
	fix := filepath.Join(repoRoot(t), "e2e", "fixtures", "invalid.ir.json")
	raw, err := os.ReadFile(fix)
	if err != nil {
		t.Fatal(err)
	}
	_, err = emitter.ParseDocument(raw)
	if err == nil {
		t.Fatal("expected parse error for invalid IR")
	}
}
