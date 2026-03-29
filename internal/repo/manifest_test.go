package repo

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverManifestsTerraformAndINI(t *testing.T) {
	dir := t.TempDir()
	_ = os.MkdirAll(filepath.Join(dir, "tf"), 0o755)
	state := []byte(`{"version":4,"values":{"root_module":{"resources":[{"address":"aws_instance.x","mode":"managed","type":"aws_instance","name":"x"}]}}}`)
	if err := os.WriteFile(filepath.Join(dir, "tf", "terraform.tfstate"), state, 0o600); err != nil {
		t.Fatal(err)
	}
	inv := []byte("[web]\napp1\n")
	if err := os.WriteFile(filepath.Join(dir, "hosts.ini"), inv, 0o600); err != nil {
		t.Fatal(err)
	}
	res, err := DiscoverManifests([]string{dir}, DefaultManifestDiscoverOptions())
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Manifests) != 2 {
		t.Fatalf("manifests %d: %+v", len(res.Manifests), res.Manifests)
	}
}
