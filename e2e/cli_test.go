package e2e

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
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

func omnigraphBin(t *testing.T) string {
	t.Helper()
	root := repoRoot(t)
	name := "omnigraph"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	out := filepath.Join(t.TempDir(), name)
	cmd := exec.Command("go", "build", "-o", out, "./cmd/omnigraph")
	cmd.Dir = root
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("go build: %v\n%s", err, stderr.String())
	}
	return out
}

func TestCLI_IR_Validate_MinimalFixture(t *testing.T) {
	bin := omnigraphBin(t)
	fix := filepath.Join(repoRoot(t), "e2e", "fixtures", "minimal.ir.json")
	cmd := exec.Command(bin, "ir", "validate", "--file", fix)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		t.Fatalf("ir validate: %v\n%s", err, out.String())
	}
	if !strings.Contains(out.String(), "e2e-fixture") {
		t.Fatalf("unexpected output: %s", out.String())
	}
}

func TestCLI_IR_Emit_AnsibleInventory(t *testing.T) {
	bin := omnigraphBin(t)
	fix := filepath.Join(repoRoot(t), "e2e", "fixtures", "minimal.ir.json")
	cmd := exec.Command(bin, "ir", "emit", "--file", fix, "--format", "ansible-inventory-ini", "--out", "-")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("ir emit: %v\n%s", err, stderr.String())
	}
	body := stdout.String()
	if !strings.Contains(body, "ansible_host=10.0.0.99") {
		t.Fatalf("expected inventory line in output: %q", body)
	}
}

func TestCLI_IR_Validate_InvalidFixtureFails(t *testing.T) {
	bin := omnigraphBin(t)
	fix := filepath.Join(repoRoot(t), "e2e", "fixtures", "invalid.ir.json")
	cmd := exec.Command(bin, "ir", "validate", "--file", fix)
	if err := cmd.Run(); err == nil {
		t.Fatal("expected non-zero exit for invalid IR")
	}
}
