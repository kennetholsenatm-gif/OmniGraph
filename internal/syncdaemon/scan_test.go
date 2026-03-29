package syncdaemon

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestBuildPatchFromRoots_TerraformState(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	// Minimal state with one managed resource so normalization emits at least one node.
	stateJSON := []byte(`{
  "values": {
    "root_module": {
      "resources": [
        {
          "address": "aws_instance.x",
          "mode": "managed",
          "type": "aws_instance",
          "name": "x",
          "values": { "private_ip": "10.1.1.1" }
        }
      ]
    }
  }
}`)
	if err := os.WriteFile(filepath.Join(dir, "terraform.tfstate"), stateJSON, 0o600); err != nil {
		t.Fatal(err)
	}

	prev := newScanTrack()
	patch, next, changed, err := buildPatchFromRoots(ctx, []string{dir}, prev)
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("expected first scan to produce a patch")
	}
	if len(patch.UpsertNodes) == 0 {
		t.Fatalf("expected nodes, got %+v", patch)
	}
	if len(next.nodeIDs) == 0 {
		t.Fatal("next track empty")
	}

	patch2, next2, changed2, err := buildPatchFromRoots(ctx, []string{dir}, next)
	if err != nil {
		t.Fatal(err)
	}
	if changed2 {
		t.Fatalf("identical scan should not emit, patch=%+v", patch2)
	}
	if next2 != next {
		t.Fatal("track pointer should be reused when unchanged")
	}
}

func TestBuildPatchFromRoots_RemovesStaleNodes(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	path := filepath.Join(dir, "terraform.tfstate")
	stateJSON := []byte(`{
  "values": {
    "root_module": {
      "resources": [
        {
          "address": "aws_instance.x",
          "mode": "managed",
          "type": "aws_instance",
          "name": "x",
          "values": { "private_ip": "10.1.1.1" }
        }
      ]
    }
  }
}`)
	if err := os.WriteFile(path, stateJSON, 0o600); err != nil {
		t.Fatal(err)
	}

	prev := newScanTrack()
	_, track1, _, err := buildPatchFromRoots(ctx, []string{dir}, prev)
	if err != nil {
		t.Fatal(err)
	}
	_ = os.Remove(path)
	patch2, _, changed, err := buildPatchFromRoots(ctx, []string{dir}, track1)
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("expected patch after file removed")
	}
	if len(patch2.RemoveNodes) == 0 {
		t.Fatalf("expected RemoveNodes after file removed, patch=%+v", patch2)
	}
}
