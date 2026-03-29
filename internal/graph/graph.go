// Package graph defines omnigraph/graph/v1 documents and helpers. Document, Graph, and GraphSpec
// are immutable snapshots for JSON; use ConcurrentGraph for goroutine-safe incremental mutation.
package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/kennetholsenatm-gif/omnigraph/internal/coerce"
	"github.com/kennetholsenatm-gif/omnigraph/internal/plan"
	"github.com/kennetholsenatm-gif/omnigraph/internal/project"
	"github.com/kennetholsenatm-gif/omnigraph/internal/state"
	"github.com/kennetholsenatm-gif/omnigraph/internal/telemetry"
	"golang.org/x/sync/errgroup"
)

const apiVersion = "omnigraph/graph/v1"
const kind = "Graph"

// EdgeDependencyRole values for Edge.DependencyRole (JSON dependencyRole).
const (
	EdgeDependencyNecessary  = "necessary"
	EdgeDependencySufficient = "sufficient"
)

// Document is the versioned graph payload for UI and PR comments.
type Document struct {
	APIVersion string    `json:"apiVersion"`
	Kind       string    `json:"kind"`
	Metadata   Metadata  `json:"metadata"`
	Spec       GraphSpec `json:"spec"`
}

// Metadata identifies the graph emission.
type Metadata struct {
	GeneratedAt string `json:"generatedAt"`
	Project     string `json:"project,omitempty"`
	Environment string `json:"environment,omitempty"`
}

// GraphSpec holds nodes and edges for the visualizer.
type GraphSpec struct {
	Phase   string      `json:"phase"`
	Nodes   []Node      `json:"nodes"`
	Edges   []Edge      `json:"edges"`
	Phases  []PhaseInfo `json:"phases,omitempty"`
	Summary *RunSummary `json:"summary,omitempty"`
}

// Node is a vertex in the dependency / topology graph.
type Node struct {
	ID         string         `json:"id"`
	Kind       string         `json:"kind"`
	Label      string         `json:"label"`
	State      string         `json:"state,omitempty"`
	Attributes map[string]any `json:"attributes,omitempty"`
}

// Edge links two nodes.
type Edge struct {
	From string `json:"from"`
	To   string `json:"to"`
	Kind string `json:"kind,omitempty"`
	// DependencyRole classifies the edge for blast-radius and triage: "necessary" (hard / critical path)
	// or "sufficient" (optimization or fallback). Empty omitempty JSON means necessary (backward compatible).
	DependencyRole string `json:"dependencyRole,omitempty"`
}

// PhaseInfo records lifecycle progress.
type PhaseInfo struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Detail string `json:"detail,omitempty"`
}

// RunSummary captures coarse tool outcomes (for PR annotations).
type RunSummary struct {
	ValidateOK bool   `json:"validateOk"`
	CoerceOK   bool   `json:"coerceOk"`
	Inventory  string `json:"inventoryPreview,omitempty"`
}

// Graph represents the internal graph structure.
type Graph struct {
	APIVersion string    `json:"apiVersion"`
	Kind       string    `json:"kind"`
	Metadata   Metadata  `json:"metadata"`
	Spec       GraphSpec `json:"spec"`
}

// EmitOptions configures optional plan/state enrichment.
type EmitOptions struct {
	PlanJSONPath   string
	TerraformState *state.TerraformState
	// TelemetryPath loads omnigraph/telemetry/v1 JSON and merges nodes/edges (see internal/telemetry).
	TelemetryPath string
}

// Emit builds a Graph v1 document from a validated project document and coercion artifacts.
func Emit(doc *project.Document, art *coerce.Artifacts, opts EmitOptions) (*Document, error) {
	if doc == nil {
		return nil, fmt.Errorf("nil project document")
	}
	now := time.Now().UTC().Format(time.RFC3339)
	g := &Document{
		APIVersion: apiVersion,
		Kind:       kind,
		Metadata: Metadata{
			GeneratedAt: now,
			Project:     doc.Metadata.Name,
			Environment: doc.Metadata.Environment,
		},
		Spec: GraphSpec{
			Phase: "plan",
			Nodes: []Node{
				{ID: "project", Kind: "project", Label: doc.Metadata.Name, State: "active"},
				{ID: "tf", Kind: "tool", Label: "OpenTofu/Terraform", State: "pending"},
				{ID: "ansible", Kind: "tool", Label: "Ansible", State: "pending"},
			},
			Edges: []Edge{
				{From: "project", To: "tf", Kind: "provisions"},
				{From: "tf", To: "ansible", Kind: "configures"},
			},
			Phases: []PhaseInfo{
				{Name: "validate", Status: "ok"},
				{Name: "coerce", Status: "ok"},
				{Name: "plan", Status: "pending"},
				{Name: "apply", Status: "pending"},
			},
			Summary: &RunSummary{ValidateOK: true, CoerceOK: art != nil},
		},
	}
	if opts.PlanJSONPath != "" {
		pj, err := plan.Load(opts.PlanJSONPath)
		if err != nil {
			return nil, err
		}
		seenPlanned := make(map[string]struct{})
		hosts := plan.ProjectedHosts(pj)
		g.Spec.Nodes = slices.Grow(g.Spec.Nodes, len(hosts)+8)
		g.Spec.Edges = slices.Grow(g.Spec.Edges, len(hosts)+8)
		for _, addr := range sortedStringKeys(hosts) {
			ip := hosts[addr]
			id := PlannedResourceNodeID(addr)
			seenPlanned[id] = struct{}{}
			g.Spec.Nodes = append(g.Spec.Nodes, Node{
				ID:    id,
				Kind:  "host",
				Label: addr,
				State: "planned",
				Attributes: map[string]any{
					"ansible_host": ip,
				},
			})
			g.Spec.Edges = append(g.Spec.Edges, Edge{From: "tf", To: id, Kind: "creates"})
		}
		for _, addr := range plan.MutationSeedAddresses(pj) {
			id := PlannedResourceNodeID(addr)
			if _, ok := seenPlanned[id]; ok {
				continue
			}
			seenPlanned[id] = struct{}{}
			g.Spec.Nodes = append(g.Spec.Nodes, Node{
				ID:    id,
				Kind:  "resource",
				Label: addr,
				State: "planned",
				Attributes: map[string]any{
					"terraform_address": addr,
				},
			})
			g.Spec.Edges = append(g.Spec.Edges, Edge{From: "tf", To: id, Kind: "mutates"})
		}
	}
	if opts.TerraformState != nil {
		hosts := state.ExtractHosts(opts.TerraformState)
		g.Spec.Nodes = slices.Grow(g.Spec.Nodes, len(hosts))
		g.Spec.Edges = slices.Grow(g.Spec.Edges, len(hosts))
		for _, addr := range sortedStringKeys(hosts) {
			ip := hosts[addr]
			id := "live-" + AddressSlug(addr)
			g.Spec.Nodes = append(g.Spec.Nodes, Node{
				ID:    id,
				Kind:  "host",
				Label: addr,
				State: "live",
				Attributes: map[string]any{
					"ansible_host": ip,
				},
			})
			g.Spec.Edges = append(g.Spec.Edges, Edge{From: "tf", To: id, Kind: "managed"})
		}
	}
	if opts.TelemetryPath != "" {
		bun, err := telemetry.LoadBundle(opts.TelemetryPath)
		if err != nil {
			return nil, err
		}
		mergeTelemetry(g, bun)
	}
	return g, nil
}

func mergeTelemetry(d *Document, b *telemetry.Bundle) {
	if d == nil || b == nil {
		return
	}
	seen := make(map[string]struct{}, len(d.Spec.Nodes))
	for _, n := range d.Spec.Nodes {
		seen[n.ID] = struct{}{}
	}
	d.Spec.Nodes = slices.Grow(d.Spec.Nodes, len(b.Nodes))
	d.Spec.Edges = slices.Grow(d.Spec.Edges, len(b.Edges))
	for _, n := range b.Nodes {
		if n.ID == "" {
			continue
		}
		if _, ok := seen[n.ID]; ok {
			continue
		}
		seen[n.ID] = struct{}{}
		d.Spec.Nodes = append(d.Spec.Nodes, Node{
			ID:         n.ID,
			Kind:       n.Kind,
			Label:      n.Label,
			State:      n.State,
			Attributes: n.Attributes,
		})
	}
	for _, e := range b.Edges {
		if e.From == "" || e.To == "" {
			continue
		}
		d.Spec.Edges = append(d.Spec.Edges, Edge{From: e.From, To: e.To, Kind: e.Kind, DependencyRole: e.DependencyRole})
	}
}

func sortedStringKeys(m map[string]string) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	return ks
}

func truncateString(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

func slug(s string) string {
	return AddressSlug(s)
}

// AddressSlug returns a stable token for a Terraform/OpenTofu resource address, used in graph node IDs.
func AddressSlug(s string) string {
	b := make([]rune, 0, len(s))
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			b = append(b, r)
		default:
			b = append(b, '_')
		}
	}
	if len(b) == 0 {
		return "host"
	}
	return string(b)
}

// PlannedResourceNodeID is the graph node id for a planned resource at the given Terraform address (matches Emit).
func PlannedResourceNodeID(address string) string {
	return "planned-" + AddressSlug(address)
}

// ParseDocument parses JSON bytes into a Document struct.
func ParseDocument(data []byte) (*Document, error) {
	return ParseDocumentWithContext(context.Background(), data, ValidateDocumentOptions{})
}

// ParseDocumentWithContext parses JSON and validates with opts. Context is passed to the errgroup
// used for concurrent validation (deadline/cancel applies to sub-checks that observe ctx).
func ParseDocumentWithContext(ctx context.Context, data []byte, opts ValidateDocumentOptions) (*Document, error) {
	var doc Document
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	if err := validateDocumentWithOptions(ctx, &doc, opts); err != nil {
		return nil, err
	}
	return &doc, nil
}

// ValidateDocumentOptions configures optional topology rules. Zero value matches historical
// behavior (structural checks only); flags enable stricter validation.
type ValidateDocumentOptions struct {
	// RejectOrphanNodesWhenEdgesExist, when true and spec has at least one edge, rejects nodes
	// with no incident edge (neither From nor To).
	RejectOrphanNodesWhenEdgesExist bool
	// RejectMultipleWeakComponents, when true and spec has at least one edge, rejects graphs
	// with more than one weakly connected component (undirected view of edges).
	RejectMultipleWeakComponents bool
}

// ValidateDocumentWithOptions runs validation (including optional topology checks) on an in-memory document.
func ValidateDocumentWithOptions(ctx context.Context, doc *Document, opts ValidateDocumentOptions) error {
	return validateDocumentWithOptions(ctx, doc, opts)
}

// validateDocument validates the Document with default options (parallel structural checks).
func validateDocument(doc *Document) error {
	return validateDocumentWithOptions(context.Background(), doc, ValidateDocumentOptions{})
}

func validateDocumentPreamble(doc *Document) error {
	if doc == nil {
		return fmt.Errorf("%w", ErrNilDocument)
	}
	if doc.APIVersion != apiVersion {
		return fmt.Errorf("%w: expected %q, got %q", ErrWrongAPIVersion, apiVersion, doc.APIVersion)
	}
	if doc.Kind != kind {
		return fmt.Errorf("%w: expected %q, got %q", ErrWrongKind, kind, doc.Kind)
	}
	if doc.Spec.Phase == "" {
		return fmt.Errorf("%w", ErrEmptyPhase)
	}
	if len(doc.Spec.Nodes) == 0 {
		return fmt.Errorf("%w", ErrEmptyNodes)
	}
	return nil
}

func buildNodeIDSet(nodes []Node) map[string]struct{} {
	nodeIDs := make(map[string]struct{}, len(nodes))
	for _, node := range nodes {
		nodeIDs[node.ID] = struct{}{}
	}
	return nodeIDs
}

func validateNodeFields(nodes []Node) error {
	for _, node := range nodes {
		if node.ID == "" {
			return fmt.Errorf("%w", ErrEmptyNodeID)
		}
		if node.Kind == "" {
			return fmt.Errorf("node %q: %w", node.ID, ErrEmptyNodeKind)
		}
		if node.Label == "" {
			return fmt.Errorf("node %q: %w", node.ID, ErrEmptyNodeLabel)
		}
	}
	return nil
}

func validateEdgeRefs(edges []Edge, nodeIDs map[string]struct{}) error {
	for _, edge := range edges {
		if edge.From == "" {
			return fmt.Errorf("%w", ErrEmptyEdgeFrom)
		}
		if edge.To == "" {
			return fmt.Errorf("%w", ErrEmptyEdgeTo)
		}
		if _, ok := nodeIDs[edge.From]; !ok {
			return fmt.Errorf("%w", &UnknownNodeError{ID: edge.From})
		}
		if _, ok := nodeIDs[edge.To]; !ok {
			return fmt.Errorf("%w", &UnknownNodeError{ID: edge.To})
		}
		dr := strings.TrimSpace(edge.DependencyRole)
		if dr != "" && dr != EdgeDependencyNecessary && dr != EdgeDependencySufficient {
			return fmt.Errorf("edge %q -> %q: %w", edge.From, edge.To, ErrInvalidDependencyRole)
		}
	}
	return nil
}

// validationParallelChunkCount returns how many goroutines to use for scanning a slice of length n.
func validationParallelChunkCount(n int) int {
	if n <= 1024 {
		return 1
	}
	k := n / 1024
	if k > 8 {
		return 8
	}
	if k < 1 {
		return 1
	}
	return k
}

func nodeChunkValidator(nodes []Node, lo, hi int) func() error {
	return func() error {
		return validateNodeFields(nodes[lo:hi])
	}
}

func edgeChunkValidator(edges []Edge, nodeIDs map[string]struct{}, lo, hi int) func() error {
	return func() error {
		return validateEdgeRefs(edges[lo:hi], nodeIDs)
	}
}

func scheduleNodeValidation(g *errgroup.Group, nodes []Node) {
	n := len(nodes)
	if n == 0 {
		return
	}
	chunks := validationParallelChunkCount(n)
	if chunks <= 1 {
		g.Go(func() error {
			return validateNodeFields(nodes)
		})
		return
	}
	size := (n + chunks - 1) / chunks
	for i := range chunks {
		lo := i * size
		if lo >= n {
			break
		}
		hi := min(lo+size, n)
		g.Go(nodeChunkValidator(nodes, lo, hi))
	}
}

func scheduleEdgeValidation(g *errgroup.Group, edges []Edge, nodeIDs map[string]struct{}) {
	n := len(edges)
	if n == 0 {
		return
	}
	chunks := validationParallelChunkCount(n)
	if chunks <= 1 {
		g.Go(func() error {
			return validateEdgeRefs(edges, nodeIDs)
		})
		return
	}
	size := (n + chunks - 1) / chunks
	for i := range chunks {
		lo := i * size
		if lo >= n {
			break
		}
		hi := min(lo+size, n)
		g.Go(edgeChunkValidator(edges, nodeIDs, lo, hi))
	}
}

func validateOrphanNodesWhenEdgesExist(doc *Document) error {
	if len(doc.Spec.Edges) == 0 {
		return nil
	}
	incident := make(map[string]struct{})
	for _, e := range doc.Spec.Edges {
		if e.From == "" || e.To == "" {
			continue
		}
		incident[e.From] = struct{}{}
		incident[e.To] = struct{}{}
	}
	for _, n := range doc.Spec.Nodes {
		if n.ID == "" {
			continue
		}
		if _, ok := incident[n.ID]; !ok {
			return fmt.Errorf("node %q: %w", n.ID, ErrOrphanNode)
		}
	}
	return nil
}

func validateSingleWeakComponent(doc *Document, nodeIDs map[string]struct{}) error {
	if len(doc.Spec.Edges) == 0 {
		return nil
	}
	parent := make(map[string]string, len(nodeIDs))
	for id := range nodeIDs {
		parent[id] = id
	}
	var find func(string) string
	find = func(x string) string {
		p := parent[x]
		if p == "" {
			return x
		}
		if p != x {
			parent[x] = find(p)
		}
		return parent[x]
	}
	union := func(a, b string) {
		ra, rb := find(a), find(b)
		if ra == rb {
			return
		}
		if ra < rb {
			parent[rb] = ra
		} else {
			parent[ra] = rb
		}
	}
	for _, e := range doc.Spec.Edges {
		if e.From == "" || e.To == "" {
			continue
		}
		union(e.From, e.To)
	}
	roots := make(map[string]struct{})
	for id := range nodeIDs {
		roots[find(id)] = struct{}{}
	}
	if len(roots) > 1 {
		return fmt.Errorf("%w: count=%d", ErrMultipleWeakComponents, len(roots))
	}
	return nil
}

func validateDocumentWithOptions(ctx context.Context, doc *Document, opts ValidateDocumentOptions) error {
	if err := validateDocumentPreamble(doc); err != nil {
		return err
	}
	nodeIDs := buildNodeIDSet(doc.Spec.Nodes)

	g, _ := errgroup.WithContext(ctx)
	scheduleNodeValidation(g, doc.Spec.Nodes)
	scheduleEdgeValidation(g, doc.Spec.Edges, nodeIDs)

	if opts.RejectOrphanNodesWhenEdgesExist {
		g.Go(func() error {
			return validateOrphanNodesWhenEdgesExist(doc)
		})
	}
	if opts.RejectMultipleWeakComponents {
		g.Go(func() error {
			return validateSingleWeakComponent(doc, nodeIDs)
		})
	}

	return g.Wait()
}

// ConstructFromDocument builds a graph structure from a parsed Document.
func ConstructFromDocument(doc *Document) (*Graph, error) {
	if doc == nil {
		return nil, fmt.Errorf("%w", ErrNilDocument)
	}
	if err := validateDocument(doc); err != nil {
		return nil, err
	}

	graph := &Graph{
		APIVersion: doc.APIVersion,
		Kind:       doc.Kind,
		Metadata: Metadata{
			GeneratedAt: doc.Metadata.GeneratedAt,
			Project:     doc.Metadata.Project,
			Environment: doc.Metadata.Environment,
		},
		Spec: GraphSpec{
			Phase:   doc.Spec.Phase,
			Nodes:   make([]Node, len(doc.Spec.Nodes)),
			Edges:   make([]Edge, len(doc.Spec.Edges)),
			Phases:  doc.Spec.Phases,
			Summary: doc.Spec.Summary,
		},
	}

	// Copy nodes
	for i, node := range doc.Spec.Nodes {
		graph.Spec.Nodes[i] = Node{
			ID:         node.ID,
			Kind:       node.Kind,
			Label:      node.Label,
			State:      node.State,
			Attributes: node.Attributes,
		}
	}

	// Copy edges
	for i, edge := range doc.Spec.Edges {
		graph.Spec.Edges[i] = Edge{
			From:           edge.From,
			To:             edge.To,
			Kind:           edge.Kind,
			DependencyRole: edge.DependencyRole,
		}
	}

	return graph, nil
}

// EncodeIndent returns indented JSON for human-readable artifacts.
func EncodeIndent(g *Document) ([]byte, error) {
	return json.MarshalIndent(g, "", "  ")
}
