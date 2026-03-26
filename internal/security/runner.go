package security

import (
	"context"
	"strings"
	"time"
)

// Filter limits which modules run (empty fields mean no filter).
type Filter struct {
	Tactic    string
	Technique string
	ModuleID  string
}

// Run executes modules against h and returns a Document.
// ansibleHost is optional metadata for graph merge (inventory address); when empty, Target from h.Label() is used for matching.
func Run(ctx context.Context, h Host, transport, profile, ansibleHost string, f Filter, moduleTimeout time.Duration) *Document {
	if moduleTimeout <= 0 {
		moduleTimeout = 90 * time.Second
	}
	target := ""
	if h != nil {
		target = h.Label()
	}
	detectCtx, detectCancel := context.WithTimeout(ctx, 15*time.Second)
	targetLinux := detectLinux(detectCtx, h)
	detectCancel()
	var out []ModuleResult
	for _, m := range All {
		if !matchFilter(m, f) {
			continue
		}
		mctx, cancel := context.WithTimeout(ctx, moduleTimeout)
		res := m.Check(mctx, h, targetLinux)
		cancel()
		if res.ModuleID == "" {
			res.ModuleID = m.ID()
		}
		if res.TechniqueID == "" {
			res.TechniqueID = m.TechniqueID()
		}
		if res.TechniqueName == "" {
			res.TechniqueName = m.TechniqueName()
		}
		if res.Tactic == "" {
			res.Tactic = m.Tactic()
		}
		if res.Severity == "" {
			res.Severity = m.Severity()
		}
		out = append(out, res)
	}
	return NewDocument(target, ansibleHost, transport, profile, out)
}

func matchFilter(m Module, f Filter) bool {
	if f.ModuleID != "" && !strings.EqualFold(m.ID(), f.ModuleID) {
		return false
	}
	if f.Technique != "" && !strings.EqualFold(m.TechniqueID(), f.Technique) {
		return false
	}
	if f.Tactic != "" && !strings.EqualFold(m.Tactic(), f.Tactic) {
		return false
	}
	return true
}

func detectLinux(ctx context.Context, h Host) bool {
	if h == nil {
		return false
	}
	out, _, code, err := h.Run(ctx, []string{"uname", "-s"})
	if err != nil || code != 0 {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(out), "linux")
}
