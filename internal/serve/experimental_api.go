package serve

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/kennetholsenatm-gif/omnigraph/internal/hostops"
	"github.com/kennetholsenatm-gif/omnigraph/internal/security"
)

type sshTargetBody struct {
	Host            string `json:"host"`
	Port            string `json:"port"`
	User            string `json:"user"`
	KeyPath         string `json:"keyPath"`
	InsecureHostKey bool   `json:"insecureHostKey"`
	KnownHostsPath  string `json:"knownHostsPath"`
}

type securityScanAPIRequest struct {
	Mode      string `json:"mode"`
	Profile   string `json:"profile"`
	Tactic    string `json:"tactic"`
	Technique string `json:"technique"`
	Module    string `json:"module"`
}

func (s *server) postSecurityScanAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req securityScanAPIRequest
	if r.Body != nil {
		defer r.Body.Close()
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&req); err != nil && !errors.Is(err, io.EOF) {
			http.Error(w, fmt.Sprintf("invalid JSON: %v", err), http.StatusBadRequest)
			return
		}
	}
	mode := strings.ToLower(strings.TrimSpace(req.Mode))
	if mode == "" {
		mode = "local"
	}
	if mode != "local" {
		http.Error(w, "only mode=local is supported for this endpoint", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Minute)
	defer cancel()
	f := security.Filter{Tactic: req.Tactic, Technique: req.Technique, ModuleID: req.Module}
	profile := req.Profile
	if profile == "" {
		profile = "api-local"
	}
	doc := security.Run(ctx, security.LocalHost{}, "local", profile, "", f, 0)
	if s.audit != nil {
		s.audit.Append("security_scan_local", "mode=local")
	}
	w.Header().Set("Content-Type", "application/json")
	b, err := security.EncodeIndent(doc)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, _ = w.Write(append(b, '\n'))
}

func dialSSHHost(body sshTargetBody) (*security.SSHHost, error) {
	cfg := security.SSHDialConfig{
		Host:            strings.TrimSpace(body.Host),
		Port:            strings.TrimSpace(body.Port),
		User:            strings.TrimSpace(body.User),
		KeyPath:         strings.TrimSpace(body.KeyPath),
		InsecureHostKey: body.InsecureHostKey,
		KnownHostsPath:  strings.TrimSpace(body.KnownHostsPath),
	}
	if cfg.Host == "" || cfg.User == "" {
		return nil, fmt.Errorf("host and user are required")
	}
	if cfg.KeyPath == "" {
		return nil, fmt.Errorf("keyPath is required")
	}
	return security.DialSSH(cfg)
}

func (s *server) postHostOpsSystemdUnits(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body sshTargetBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	sh, err := dialSSHHost(body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer sh.Close()
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Minute)
	defer cancel()
	out, err := hostops.ListServiceUnits(ctx, sh.Client())
	if s.audit != nil {
		s.audit.Append("host_ops_systemd_units", body.Host)
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"output": out})
}

type journalAPIRequest struct {
	sshTargetBody
	Unit  string `json:"unit"`
	Lines int    `json:"lines"`
}

func (s *server) postHostOpsJournal(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body journalAPIRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	sh, err := dialSSHHost(body.sshTargetBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer sh.Close()
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Minute)
	defer cancel()
	out, err := hostops.JournalTail(ctx, sh.Client(), body.Unit, body.Lines)
	if s.audit != nil {
		s.audit.Append("host_ops_journal_tail", body.Host+" "+body.Unit)
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"output": out})
}

type restartAPIRequest struct {
	sshTargetBody
	Unit string `json:"unit"`
}

func (s *server) postHostOpsRestart(w http.ResponseWriter, r *http.Request) {
	if !s.hostOpsAllowWrites {
		http.Error(w, "host-ops writes disabled", http.StatusForbidden)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body restartAPIRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	sh, err := dialSSHHost(body.sshTargetBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer sh.Close()
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Minute)
	defer cancel()
	out, err := hostops.RestartService(ctx, sh.Client(), body.Unit)
	if s.audit != nil {
		s.audit.Append("host_ops_systemd_restart", body.Host+" "+body.Unit)
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"output": out})
}

func (s *server) getAudit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.audit == nil {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("[]\n"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(s.audit.Snapshot())
}
