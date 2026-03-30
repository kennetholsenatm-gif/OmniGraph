package syncdaemon

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kennetholsenatm-gif/omnigraph/internal/omnistate"
	"github.com/kennetholsenatm-gif/omnigraph/internal/repo"
)

const maxScanFileBytes = 6 << 20

// scanTrack remembers node and edge keys last pushed so the next patch can remove stale entries.
type scanTrack struct {
	nodeIDs map[string]struct{}
	edgeMap map[string]omnistate.EdgeKey
}

func newScanTrack() *scanTrack {
	return &scanTrack{
		nodeIDs: make(map[string]struct{}),
		edgeMap: make(map[string]omnistate.EdgeKey),
	}
}

func edgeKeyOf(e omnistate.StateEdge) omnistate.EdgeKey {
	return omnistate.EdgeKey{From: e.From, To: e.To, Kind: e.Kind}
}

func edgeKeyString(k omnistate.EdgeKey) string {
	return k.From + "\x00" + k.To + "\x00" + k.Kind
}

// buildPatchFromRoots walks writable roots (same discovery rules as repo.Discover), normalizes
// Terraform state and Ansible inventory files, merges into one state, and returns a hub patch that
// removes nodes/edges no longer present and upserts the current scan.
func buildPatchFromRoots(ctx context.Context, roots []string, prev *scanTrack) (omnistate.StatePatch, *scanTrack, bool, error) {
	if prev == nil {
		prev = newScanTrack()
	}
	var frags []omnistate.OmniGraphStateFragment
	seenAbs := make(map[string]struct{})

	for _, root := range roots {
		root = strings.TrimSpace(root)
		if root == "" {
			continue
		}
		absRoot, err := filepath.Abs(root)
		if err != nil {
			continue
		}
		res, err := repo.Discover(absRoot)
		if err != nil {
			continue
		}
		for _, f := range res.Files {
			if f.Kind != repo.KindTerraformState && f.Kind != repo.KindAnsibleInventory {
				continue
			}
			absFile := filepath.Join(absRoot, filepath.FromSlash(f.Path))
			if _, dup := seenAbs[absFile]; dup {
				continue
			}
			seenAbs[absFile] = struct{}{}

			raw, err := readFileLimited(absFile, maxScanFileBytes)
			if err != nil {
				continue
			}
			rel := filepath.ToSlash(f.Path)
			base := filepath.Base(f.Path)
			ct := contentTypeForRepoKind(f.Kind)
			ref := omnistate.SourceRef{
				Type:     omnistate.DetectNormalizer(ct, base).Kind(),
				Name:     rel,
				PathHint: absFile,
			}
			n := omnistate.DetectNormalizer(ct, base)
			frag, err := n.Normalize(ctx, omnistate.NormalizerInput{
				Data:        raw,
				ContentType: ct,
				Name:        rel,
				Ref:         ref,
			})
			if err != nil {
				continue
			}
			frags = append(frags, frag)
		}
	}

	merged := omnistate.MergeFragments("syncdaemon-scan", frags...)
	next := newScanTrack()
	for _, n := range merged.Nodes {
		next.nodeIDs[n.ID] = struct{}{}
	}
	for _, e := range merged.Edges {
		ek := edgeKeyOf(e)
		next.edgeMap[edgeKeyString(ek)] = ek
	}

	var patch omnistate.StatePatch
	for id := range prev.nodeIDs {
		if _, ok := next.nodeIDs[id]; !ok {
			patch.RemoveNodes = append(patch.RemoveNodes, id)
		}
	}
	for ekStr, ek := range prev.edgeMap {
		if _, ok := next.edgeMap[ekStr]; !ok {
			patch.RemoveEdges = append(patch.RemoveEdges, ek)
		}
	}
	if scanTrackEqual(prev, next) {
		return omnistate.StatePatch{}, prev, false, nil
	}

	patch.UpsertNodes = merged.Nodes
	patch.UpsertEdges = merged.Edges
	return patch, next, true, nil
}

func scanTrackEqual(a, b *scanTrack) bool {
	if len(a.nodeIDs) != len(b.nodeIDs) || len(a.edgeMap) != len(b.edgeMap) {
		return false
	}
	for id := range a.nodeIDs {
		if _, ok := b.nodeIDs[id]; !ok {
			return false
		}
	}
	for k := range a.edgeMap {
		if _, ok := b.edgeMap[k]; !ok {
			return false
		}
	}
	return true
}

func contentTypeForRepoKind(k repo.FileKind) string {
	switch k {
	case repo.KindTerraformState:
		return "application/json"
	default:
		return "text/plain"
	}
}

func readFileLimited(path string, max int64) ([]byte, error) {
	st, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if st.Size() > max {
		return nil, fmt.Errorf("file exceeds limit (%d bytes)", max)
	}
	return os.ReadFile(path)
}
