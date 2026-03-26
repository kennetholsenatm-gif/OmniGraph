package serve

import (
	"sync"
	"time"
)

// AuditEntry records a privileged API action.
type AuditEntry struct {
	Time   string `json:"time"`
	Action string `json:"action"`
	Detail string `json:"detail,omitempty"`
}

// AuditLog is a fixed-size in-memory ring of recent entries.
type AuditLog struct {
	mu      sync.Mutex
	entries []AuditEntry
	cap     int
}

// NewAuditLog returns an audit buffer with at most cap entries retained.
func NewAuditLog(cap int) *AuditLog {
	if cap <= 0 {
		cap = 100
	}
	return &AuditLog{cap: cap}
}

// Append adds an entry (newest last in Snapshot order).
func (a *AuditLog) Append(action, detail string) {
	if a == nil {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	e := AuditEntry{Time: time.Now().UTC().Format(time.RFC3339), Action: action, Detail: detail}
	a.entries = append(a.entries, e)
	if len(a.entries) > a.cap {
		a.entries = a.entries[len(a.entries)-a.cap:]
	}
}

// Snapshot returns a copy of stored entries.
func (a *AuditLog) Snapshot() []AuditEntry {
	if a == nil {
		return nil
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	out := make([]AuditEntry, len(a.entries))
	copy(out, a.entries)
	return out
}
