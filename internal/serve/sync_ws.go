package serve

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/kennetholsenatm-gif/omnigraph/internal/omnistate"
)

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin:     func(*http.Request) bool { return true },
}

// syncHub holds the latest merged omnistate for bi-directional WebSocket sync (in-memory).
type syncHub struct {
	mu    sync.RWMutex
	state omnistate.OmniGraphState
}

func newSyncHub() *syncHub {
	return &syncHub{
		state: omnistate.OmniGraphState{
			APIVersion:  omnistate.APIVersion,
			GeneratedAt: time.Now().UTC().Format(time.RFC3339Nano),
			Nodes:       []omnistate.StateNode{},
			Edges:       []omnistate.StateEdge{},
		},
	}
}

func (h *syncHub) replaceState(st omnistate.OmniGraphState) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.state = st
}

func (h *syncHub) snapshot() omnistate.OmniGraphState {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.state
}

func (h *syncHub) applyPatch(p omnistate.StatePatch) omnistate.OmniGraphState {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.state = omnistate.ApplyPatch(h.state, p)
	return h.state
}

// getSyncWebSocket upgrades to WebSocket and runs a JSON message loop until disconnect or context cancel.
// Client → server: { "type":"state_delta", "patch": StatePatch, "baseRevision": number }
// Server → client: { "type":"state_ack", "revision": number }
// Server → client: { "type":"apply_mutation", "mutationId": string, "targetPath": string, "encoding": "utf8"|"base64", "data": string }
// Client → server: { "type":"mutation_result", "mutationId": string, "ok": bool, "message": string }
func (s *server) getSyncWebSocket(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.syncHub == nil {
		http.Error(w, "sync hub not configured", http.StatusServiceUnavailable)
		return
	}
	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()
	_ = conn.SetReadDeadline(time.Now().Add(90 * time.Second))
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(90 * time.Second))
	})

	ctx := r.Context()
	pingDone := make(chan struct{})
	go wsPingLoop(ctx, conn, pingDone)
	defer close(pingDone)

	_ = s.syncReadLoop(ctx, conn)
}

func (s *server) syncReadLoop(ctx context.Context, conn *websocket.Conn) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return err
		}
		var head wsEnvelope
		if err := json.Unmarshal(msg, &head); err != nil {
			_ = writeWSError(conn, "invalid_json", err.Error())
			continue
		}
		switch head.Type {
		case "ping":
			_ = conn.WriteJSON(map[string]string{"type": "pong"})
		case "state_delta":
			var m wsStateDelta
			if err := json.Unmarshal(msg, &m); err != nil {
				_ = writeWSError(conn, "invalid_payload", err.Error())
				continue
			}
			next := s.syncHub.applyPatch(m.Patch)
			_ = conn.WriteJSON(wsStateAck{Type: "state_ack", Revision: next.Revision})
		case "mutation_result":
			// Audit / logging hook; scaffold accepts results without persistence.
			var mr wsMutationResult
			if err := json.Unmarshal(msg, &mr); err != nil {
				_ = writeWSError(conn, "invalid_payload", err.Error())
				continue
			}
			_ = conn.WriteJSON(wsGeneric{Type: "mutation_ack", MutationID: mr.MutationID, OK: mr.OK})
		default:
			_ = writeWSError(conn, "unknown_type", head.Type)
		}
	}
}

type wsEnvelope struct {
	Type string `json:"type"`
}

type wsStateDelta struct {
	Type         string               `json:"type"`
	BaseRevision int64                `json:"baseRevision,omitempty"`
	Patch        omnistate.StatePatch `json:"patch"`
	Source       string               `json:"source,omitempty"`
}

type wsStateAck struct {
	Type     string `json:"type"`
	Revision int64  `json:"revision"`
}

type wsMutationResult struct {
	Type       string `json:"type"`
	MutationID string `json:"mutationId"`
	OK         bool   `json:"ok"`
	Message    string `json:"message,omitempty"`
}

type wsGeneric struct {
	Type       string `json:"type"`
	MutationID string `json:"mutationId,omitempty"`
	OK         bool   `json:"ok,omitempty"`
}

type wsError struct {
	Type    string `json:"type"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeWSError(conn *websocket.Conn, code, msg string) error {
	return conn.WriteJSON(wsError{Type: "error", Code: code, Message: msg})
}

func wsPingLoop(ctx context.Context, conn *websocket.Conn, done <-chan struct{}) {
	t := time.NewTicker(25 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-done:
			return
		case <-t.C:
			_ = conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(5*time.Second))
		}
	}
}

// PushApplyMutation sends a write instruction to the connected agent (best-effort).
func (h *syncHub) PushApplyMutation(conn *websocket.Conn, m wsApplyMutation) error {
	return conn.WriteJSON(m)
}

// wsApplyMutation is a downstream file write for the sync daemon.
type wsApplyMutation struct {
	Type       string `json:"type"`
	MutationID string `json:"mutationId"`
	TargetPath string `json:"targetPath"`
	Encoding   string `json:"encoding"`
	Data       string `json:"data"`
}
