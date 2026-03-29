package graph

import (
	"fmt"
	"sort"
)

// BlastRadiusReport is the plan-time closure of nodes affected by proposed mutations along necessary edges.
type BlastRadiusReport struct {
	SeedNodeIDs           []string
	MutationAddresses     []string
	AffectedNodeIDs       []string
	AffectedEdgeCount     int
	TotalNodeCount        int
	DownstreamBlastReport *BlastReport
}

// ComputeBlastRadius builds a blast report from seed node IDs (e.g. planned-* ids for mutation addresses).
func ComputeBlastRadius(doc *Document, mutationAddresses []string, seedNodeIDs []string) (*BlastRadiusReport, error) {
	if doc == nil {
		return nil, fmt.Errorf("%w", ErrNilDocument)
	}
	seeds := append([]string(nil), seedNodeIDs...)
	sort.Strings(seeds)
	uniq := seeds[:0]
	var prev string
	for _, id := range seeds {
		if id == "" {
			return nil, fmt.Errorf("seed node id cannot be empty")
		}
		if len(uniq) > 0 && id == prev {
			continue
		}
		prev = id
		uniq = append(uniq, id)
	}
	seeds = uniq

	br, err := DownstreamBlast(doc, seeds)
	if err != nil {
		return nil, err
	}
	addrs := append([]string(nil), mutationAddresses...)
	sort.Strings(addrs)

	return &BlastRadiusReport{
		SeedNodeIDs:           append([]string(nil), seeds...),
		MutationAddresses:     addrs,
		AffectedNodeIDs:       append([]string(nil), br.DownstreamNodeIDs...),
		AffectedEdgeCount:     len(br.TraversedDownstream),
		TotalNodeCount:        len(doc.Spec.Nodes),
		DownstreamBlastReport: br,
	}, nil
}
