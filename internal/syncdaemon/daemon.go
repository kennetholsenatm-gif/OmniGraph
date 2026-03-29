// Package syncdaemon implements an optional background agent that maintains a WebSocket
// to the OmniGraph control plane, streams state deltas upstream, and applies file mutations
// to allowed paths. Configuration is environment-driven for deployment (sidecar, systemd, etc.).
package syncdaemon

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/kennetholsenatm-gif/omnigraph/internal/omnistate"
)

// Config is loaded from the environment by LoadConfigFromEnv.
type Config struct {
	WSURL         string
	BearerToken   string
	WritableRoots []string
}

// LoadConfigFromEnv reads OMNIGRAPH_SYNC_WS_URL, OMNIGRAPH_SYNC_TOKEN, OMNIGRAPH_SYNC_WRITABLE_PATHS
// (comma- or path-list-separated roots on the agent host).
func LoadConfigFromEnv() (Config, error) {
	url := strings.TrimSpace(os.Getenv("OMNIGRAPH_SYNC_WS_URL"))
	tok := strings.TrimSpace(os.Getenv("OMNIGRAPH_SYNC_TOKEN"))
	paths := strings.TrimSpace(os.Getenv("OMNIGRAPH_SYNC_WRITABLE_PATHS"))
	if url == "" {
		return Config{}, errors.New("OMNIGRAPH_SYNC_WS_URL is required")
	}
	if tok == "" {
		return Config{}, errors.New("OMNIGRAPH_SYNC_TOKEN is required")
	}
	normalized := strings.ReplaceAll(paths, ",", string(os.PathListSeparator))
	var roots []string
	for _, p := range strings.Split(normalized, string(os.PathListSeparator)) {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		roots = append(roots, p)
	}
	if len(roots) == 0 {
		return Config{}, errors.New("OMNIGRAPH_SYNC_WRITABLE_PATHS must list at least one directory root")
	}
	return Config{WSURL: url, BearerToken: tok, WritableRoots: roots}, nil
}

// Run connects to the control plane WebSocket and processes messages until ctx is cancelled.
func Run(ctx context.Context, cfg Config) error {
	hdr := http.Header{}
	hdr.Set("Authorization", "Bearer "+cfg.BearerToken)
	d := websocket.Dialer{HandshakeTimeout: 15 * time.Second}
	conn, _, err := d.DialContext(ctx, cfg.WSURL, hdr)
	if err != nil {
		return fmt.Errorf("syncdaemon dial: %w", err)
	}
	defer conn.Close()

	go func() {
		<-ctx.Done()
		_ = conn.Close()
	}()

	readErr := make(chan error, 1)
	go func() {
		readErr <- readLoop(conn, cfg)
	}()

	tick := time.NewTicker(20 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			_ = conn.WriteControl(websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, "shutdown"),
				time.Now().Add(time.Second))
			return ctx.Err()
		case err := <-readErr:
			if err != nil {
				return err
			}
			return nil
		case <-tick.C:
			delta, err := json.Marshal(map[string]any{
				"type":   "state_delta",
				"source": "syncdaemon",
				"patch": omnistate.StatePatch{
					UpsertNodes: []omnistate.StateNode{
						{
							ID:         "agent:heartbeat:" + uuid.NewString(),
							Kind:       "sync_agent",
							Label:      "heartbeat",
							Attributes: map[string]any{"ts": time.Now().UTC().Format(time.RFC3339Nano)},
							Provenance: omnistate.SourceRef{Type: omnistate.SourceAgentLocal, Name: "syncdaemon"},
						},
					},
				},
			})
			if err != nil {
				continue
			}
			_ = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if werr := conn.WriteMessage(websocket.TextMessage, delta); werr != nil {
				return werr
			}
		}
	}
}

func readLoop(conn *websocket.Conn, cfg Config) error {
	for {
		_ = conn.SetReadDeadline(time.Now().Add(120 * time.Second))
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return err
		}
		var env struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(msg, &env); err != nil {
			continue
		}
		switch env.Type {
		case "apply_mutation":
			var m applyMutation
			if err := json.Unmarshal(msg, &m); err != nil {
				continue
			}
			_ = handleApplyMutation(conn, cfg, m)
		case "state_ack", "mutation_ack", "pong", "error":
			// control plane responses
		}
	}
}

type applyMutation struct {
	Type       string `json:"type"`
	MutationID string `json:"mutationId"`
	TargetPath string `json:"targetPath"`
	Encoding   string `json:"encoding"`
	Data       string `json:"data"`
}

func handleApplyMutation(conn *websocket.Conn, cfg Config, m applyMutation) error {
	ok, msg := writeMutationTarget(cfg, m.TargetPath, m.Encoding, m.Data)
	resp, _ := json.Marshal(map[string]any{
		"type":       "mutation_result",
		"mutationId": m.MutationID,
		"ok":         ok,
		"message":    msg,
	})
	_ = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	return conn.WriteMessage(websocket.TextMessage, resp)
}

func writeMutationTarget(cfg Config, rel, encoding, data string) (ok bool, msg string) {
	rel = strings.TrimSpace(rel)
	if rel == "" || strings.Contains(rel, "..") {
		return false, "invalid target path"
	}
	absTarget, err := filepath.Abs(rel)
	if err != nil {
		return false, err.Error()
	}
	var allowed bool
	for _, root := range cfg.WritableRoots {
		rootAbs, err := filepath.Abs(root)
		if err != nil {
			continue
		}
		if relAllowed(absTarget, rootAbs) {
			allowed = true
			break
		}
	}
	if !allowed {
		return false, "target outside writable roots"
	}
	var raw []byte
	switch strings.ToLower(strings.TrimSpace(encoding)) {
	case "utf8", "utf-8", "":
		raw = []byte(data)
	case "base64":
		var err error
		raw, err = base64.StdEncoding.DecodeString(data)
		if err != nil {
			return false, err.Error()
		}
	default:
		return false, "unsupported encoding"
	}
	if err := os.MkdirAll(filepath.Dir(absTarget), 0o755); err != nil {
		return false, err.Error()
	}
	if err := os.WriteFile(absTarget, raw, 0o600); err != nil {
		return false, err.Error()
	}
	return true, ""
}

func relAllowed(target, root string) bool {
	t := filepath.Clean(target)
	r := filepath.Clean(root)
	rel, err := filepath.Rel(r, t)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator))
}
