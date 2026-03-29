package state

import (
	"path/filepath"
	"strings"
)

// ExecutionScope optionally restricts mutating Manager operations to an allowlist of absolute paths.
// When AllowedPaths is empty, no extra restriction is applied.
type ExecutionScope struct {
	AllowedPaths map[string]struct{}
	// AuditSummary is optional text stored on new locks for operators (e.g. blast-radius summary).
	AuditSummary string
}

// NewExecutionScopeForBlastRadius returns a scope that permits the primary Terraform/OpenTofu state file
// and its sidecar lock path. Additional paths can be added later as the control plane gains finer writes.
func NewExecutionScopeForBlastRadius(absStatePath string) *ExecutionScope {
	m := make(map[string]struct{})
	p := strings.TrimSpace(absStatePath)
	if p != "" {
		c := filepath.Clean(p)
		m[c] = struct{}{}
	}
	if len(m) == 0 {
		return nil
	}
	return &ExecutionScope{AllowedPaths: m}
}

func (s *ExecutionScope) allowsPath(absPath string) bool {
	if s == nil || len(s.AllowedPaths) == 0 {
		return true
	}
	c := filepath.Clean(absPath)
	if _, ok := s.AllowedPaths[c]; ok {
		return true
	}
	if strings.HasSuffix(c, ".lock") {
		base := strings.TrimSuffix(c, ".lock")
		_, ok := s.AllowedPaths[filepath.Clean(base)]
		return ok
	}
	return false
}
