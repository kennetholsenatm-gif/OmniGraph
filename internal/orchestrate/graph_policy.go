package orchestrate

import (
	"errors"
	"fmt"
	"strings"

	"github.com/kennetholsenatm-gif/omnigraph/internal/graph"
)

// ErrBlastRadiusExceeded is returned when an active blast-radius policy rejects a plan.
var ErrBlastRadiusExceeded = errors.New("orchestrate: blast radius exceeds policy")

// BlastRadiusPolicy limits how large a plan-time dependency closure may be before apply.
// Zero value disables all checks.
type BlastRadiusPolicy struct {
	// MaxAffectedNodeCount, if > 0, rejects when len(AffectedNodeIDs) exceeds this.
	MaxAffectedNodeCount int
	// MaxAffectedFraction, if > 0 and <= 1, rejects when affected/total nodes exceeds this.
	MaxAffectedFraction float64
	// MaxAffectedByKind maps node kind -> max allowed affected nodes of that kind (0 skips that kind).
	MaxAffectedByKind map[string]int
	// ForbiddenKindsInBlast rejects if any affected node has one of these kinds.
	ForbiddenKindsInBlast []string
}

// EvaluateBlastRadiusPolicy returns ErrBlastRadiusExceeded when limits are violated.
func EvaluateBlastRadiusPolicy(policy BlastRadiusPolicy, report *graph.BlastRadiusReport, doc *graph.Document) error {
	if report == nil || doc == nil {
		return nil
	}
	p := policy
	if p.MaxAffectedNodeCount > 0 && len(report.AffectedNodeIDs) > p.MaxAffectedNodeCount {
		return fmt.Errorf("%w: affected nodes %d > max %d", ErrBlastRadiusExceeded, len(report.AffectedNodeIDs), p.MaxAffectedNodeCount)
	}
	total := report.TotalNodeCount
	if total <= 0 {
		total = len(doc.Spec.Nodes)
	}
	if p.MaxAffectedFraction > 0 && p.MaxAffectedFraction <= 1 && total > 0 {
		frac := float64(len(report.AffectedNodeIDs)) / float64(total)
		if frac > p.MaxAffectedFraction {
			return fmt.Errorf("%w: affected fraction %.3f > max %.3f (%d/%d nodes)", ErrBlastRadiusExceeded, frac, p.MaxAffectedFraction, len(report.AffectedNodeIDs), total)
		}
	}
	if len(p.ForbiddenKindsInBlast) > 0 {
		forbidden := make(map[string]struct{}, len(p.ForbiddenKindsInBlast))
		for _, k := range p.ForbiddenKindsInBlast {
			k = strings.TrimSpace(k)
			if k != "" {
				forbidden[k] = struct{}{}
			}
		}
		kindByID := make(map[string]string, len(doc.Spec.Nodes))
		for _, n := range doc.Spec.Nodes {
			kindByID[n.ID] = n.Kind
		}
		for _, id := range report.AffectedNodeIDs {
			if k, ok := kindByID[id]; ok {
				if _, bad := forbidden[k]; bad {
					return fmt.Errorf("%w: forbidden kind %q in blast (node %q)", ErrBlastRadiusExceeded, k, id)
				}
			}
		}
	}
	if len(p.MaxAffectedByKind) > 0 {
		counts := make(map[string]int)
		kindByID := make(map[string]string, len(doc.Spec.Nodes))
		for _, n := range doc.Spec.Nodes {
			kindByID[n.ID] = n.Kind
		}
		for _, id := range report.AffectedNodeIDs {
			if k, ok := kindByID[id]; ok {
				counts[k]++
			}
		}
		for kind, maxN := range p.MaxAffectedByKind {
			if maxN <= 0 {
				continue
			}
			if counts[kind] > maxN {
				return fmt.Errorf("%w: affected %s nodes %d > max %d", ErrBlastRadiusExceeded, kind, counts[kind], maxN)
			}
		}
	}
	return nil
}

// ShouldAbortGraphPipeline reports whether err is a deterministic graph validation or topology
// problem that should not be retried (fix data or schema instead).
func ShouldAbortGraphPipeline(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, graph.ErrWrongAPIVersion) || errors.Is(err, graph.ErrWrongKind) {
		return true
	}
	if errors.Is(err, graph.ErrEmptyPhase) || errors.Is(err, graph.ErrEmptyNodes) {
		return true
	}
	if errors.Is(err, graph.ErrNilDocument) {
		return true
	}
	if errors.Is(err, ErrBlastRadiusExceeded) {
		return true
	}
	var ce *graph.CycleError
	if errors.As(err, &ce) {
		return true
	}
	var u *graph.UnknownNodeError
	if errors.As(err, &u) {
		return true
	}
	return false
}
