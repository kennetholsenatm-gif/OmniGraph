package safepath

import (
	"path/filepath"
	"testing"
)

func TestUnderRoot_rejectsTraversal(t *testing.T) {
	root := t.TempDir()
	_, err := UnderRoot(root, "../outside")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestUnderRoot_ok(t *testing.T) {
	root := t.TempDir()
	got, err := UnderRoot(root, "dir/file.txt")
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(got) != "file.txt" {
		t.Fatalf("got %q", got)
	}
}
