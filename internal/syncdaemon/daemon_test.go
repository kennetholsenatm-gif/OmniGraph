package syncdaemon

import (
	"path/filepath"
	"testing"
)

func TestLoadConfigFromEnv(t *testing.T) {
	t.Setenv("OMNIGRAPH_SYNC_WS_URL", "")
	t.Setenv("OMNIGRAPH_SYNC_TOKEN", "")
	t.Setenv("OMNIGRAPH_SYNC_WRITABLE_PATHS", "")
	if _, err := LoadConfigFromEnv(); err == nil {
		t.Fatal("expected error when env empty")
	}

	t.Setenv("OMNIGRAPH_SYNC_WS_URL", "ws://127.0.0.1:1/api/v1/sync/ws")
	t.Setenv("OMNIGRAPH_SYNC_TOKEN", "secret")
	t.Setenv("OMNIGRAPH_SYNC_WRITABLE_PATHS", filepath.Join("a", "b")+","+filepath.Join("c", "d"))
	cfg, err := LoadConfigFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.WSURL != "ws://127.0.0.1:1/api/v1/sync/ws" || cfg.BearerToken != "secret" {
		t.Fatalf("cfg %+v", cfg)
	}
	if len(cfg.WritableRoots) != 2 {
		t.Fatalf("roots %+v", cfg.WritableRoots)
	}
}
