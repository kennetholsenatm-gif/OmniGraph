package repo

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscover(t *testing.T) {
	dir := t.TempDir()
	_ = os.MkdirAll(filepath.Join(dir, "opentofu"), 0o755)
	_ = os.WriteFile(filepath.Join(dir, "opentofu", "main.tf"), []byte("#x"), 0o600)
	_ = os.WriteFile(filepath.Join(dir, "opentofu", "terraform.tfstate"), []byte("{}"), 0o600)
	_ = os.WriteFile(filepath.Join(dir, ".omnigraph.schema"), []byte("x"), 0o600)
	_ = os.MkdirAll(filepath.Join(dir, "ansible", "inventory"), 0o755)
	_ = os.WriteFile(filepath.Join(dir, "ansible", "inventory", "hosts.ini"), []byte("[a]\n"), 0o600)

	r, err := Discover(dir)
	if err != nil {
		t.Fatal(err)
	}
	if r.Root == "" {
		t.Fatal("empty root")
	}
	if len(r.Files) < 4 {
		t.Fatalf("want >=4 files, got %d: %+v", len(r.Files), r.Files)
	}
}
