package netbox

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_PostWebhook(t *testing.T) {
	var got SyncPayload
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method %s", r.Method)
		}
		b, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(b, &got); err != nil {
			t.Fatal(err)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()
	c := &Client{HTTPClient: srv.Client()}
	if err := c.PostWebhook(context.Background(), srv.URL, SyncPayload{
		Action: "create",
		IP:     "10.0.5.21",
		Role:   "web-server",
	}); err != nil {
		t.Fatal(err)
	}
	if got.IP != "10.0.5.21" || got.Action != "create" || got.Role != "web-server" {
		t.Fatalf("payload %#v", got)
	}
}

func TestClient_PostWebhook_ErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusBadRequest)
	}))
	defer srv.Close()
	c := &Client{HTTPClient: srv.Client()}
	err := c.PostWebhook(context.Background(), srv.URL, SyncPayload{Action: "create", IP: "1.1.1.1"})
	if err == nil {
		t.Fatal("expected error")
	}
}
