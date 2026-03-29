package graph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"
)

func TestParseDocument_ValidJSON(t *testing.T) {
	jsonData := []byte(`{
		"apiVersion": "omnigraph/graph/v1",
		"kind": "Graph",
		"metadata": {
			"generatedAt": "2024-01-01T00:00:00Z",
			"project": "test-project",
			"environment": "dev"
		},
		"spec": {
			"phase": "plan",
			"nodes": [
				{
					"id": "node1",
					"kind": "host",
					"label": "host1",
					"state": "active",
					"attributes": {
						"ip": "192.168.1.1"
					}
				},
				{
					"id": "node2",
					"kind": "host",
					"label": "host2",
					"state": "active",
					"attributes": {
						"ip": "192.168.1.2"
					}
				}
			],
			"edges": [
				{
					"from": "node1",
					"to": "node2",
					"kind": "connects"
				}
			]
		}
	}`)

	doc, err := ParseDocument(jsonData)
	if err != nil {
		t.Fatalf("ParseDocument failed: %v", err)
	}

	if doc.APIVersion != "omnigraph/graph/v1" {
		t.Errorf("expected apiVersion %q, got %q", "omnigraph/graph/v1", doc.APIVersion)
	}
	if doc.Kind != "Graph" {
		t.Errorf("expected kind %q, got %q", "Graph", doc.Kind)
	}
	if len(doc.Spec.Nodes) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(doc.Spec.Nodes))
	}
	if len(doc.Spec.Edges) != 1 {
		t.Errorf("expected 1 edge, got %d", len(doc.Spec.Edges))
	}
}

func TestParseDocument_InvalidJSON(t *testing.T) {
	jsonData := []byte(`{
		"apiVersion": "omnigraph/graph/v1",
		"kind": "Graph",
		"metadata": {
			"generatedAt": "2024-01-01T00:00:00Z"
		},
		"spec": {
			"phase": "plan",
			"nodes": [],
			"edges": []
		}
	}`)

	_, err := ParseDocument(jsonData)
	if err == nil {
		t.Errorf("expected error for empty nodes, got nil")
	}
}

func TestParseDocument_MissingFields(t *testing.T) {
	jsonData := []byte(`{
		"apiVersion": "omnigraph/graph/v1",
		"kind": "Graph",
		"metadata": {
			"generatedAt": "2024-01-01T00:00:00Z",
			"project": "test-project"
		},
		"spec": {
			"phase": "plan",
			"nodes": [
				{
					"id": "node1",
					"kind": "host",
					"label": "host1"
				}
			],
			"edges": [
				{
					"from": "node1",
					"to": "node2"
				}
			]
		}
	}`)

	_, err := ParseDocument(jsonData)
	if err == nil {
		t.Errorf("expected error for missing node2, got nil")
	}
}

// Concurrent validation is exercised here; run with CGO and a C compiler for
// go test -race ./internal/graph/... on Windows, or on Linux CI.
func TestValidateDocumentWithOptions_rejectOrphanNode(t *testing.T) {
	doc := &Document{
		APIVersion: apiVersion,
		Kind:       kind,
		Metadata:   Metadata{GeneratedAt: "2024-01-01T00:00:00Z"},
		Spec: GraphSpec{
			Phase: "plan",
			Nodes: []Node{
				{ID: "a", Kind: "host", Label: "a"},
				{ID: "b", Kind: "host", Label: "b"},
				{ID: "c", Kind: "host", Label: "c"},
			},
			Edges: []Edge{{From: "a", To: "b", Kind: "x"}},
		},
	}
	err := ValidateDocumentWithOptions(context.Background(), doc, ValidateDocumentOptions{
		RejectOrphanNodesWhenEdgesExist: true,
	})
	if err == nil {
		t.Fatal("expected orphan error")
	}
	if !strings.Contains(err.Error(), "orphan node") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateDocumentWithOptions_rejectMultipleWeakComponents(t *testing.T) {
	doc := &Document{
		APIVersion: apiVersion,
		Kind:       kind,
		Metadata:   Metadata{GeneratedAt: "2024-01-01T00:00:00Z"},
		Spec: GraphSpec{
			Phase: "plan",
			Nodes: []Node{
				{ID: "a", Kind: "host", Label: "a"},
				{ID: "b", Kind: "host", Label: "b"},
				{ID: "c", Kind: "host", Label: "c"},
				{ID: "d", Kind: "host", Label: "d"},
			},
			Edges: []Edge{
				{From: "a", To: "b", Kind: "x"},
				{From: "c", To: "d", Kind: "x"},
			},
		},
	}
	err := ValidateDocumentWithOptions(context.Background(), doc, ValidateDocumentOptions{
		RejectMultipleWeakComponents: true,
	})
	if err == nil {
		t.Fatal("expected weak component error")
	}
	if !strings.Contains(err.Error(), "weakly connected components") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateDocumentWithOptions_disconnectedAllowedByDefault(t *testing.T) {
	doc := &Document{
		APIVersion: apiVersion,
		Kind:       kind,
		Metadata:   Metadata{GeneratedAt: "2024-01-01T00:00:00Z"},
		Spec: GraphSpec{
			Phase: "plan",
			Nodes: []Node{
				{ID: "a", Kind: "host", Label: "a"},
				{ID: "b", Kind: "host", Label: "b"},
				{ID: "c", Kind: "host", Label: "c"},
				{ID: "d", Kind: "host", Label: "d"},
			},
			Edges: []Edge{
				{From: "a", To: "b", Kind: "x"},
				{From: "c", To: "d", Kind: "x"},
			},
		},
	}
	if err := ValidateDocumentWithOptions(context.Background(), doc, ValidateDocumentOptions{}); err != nil {
		t.Fatal(err)
	}
}

func TestParseDocumentWithContext_strictOpts(t *testing.T) {
	raw := []byte(`{
		"apiVersion": "omnigraph/graph/v1",
		"kind": "Graph",
		"metadata": { "generatedAt": "2024-01-01T00:00:00Z" },
		"spec": {
			"phase": "plan",
			"nodes": [
				{ "id": "a", "kind": "host", "label": "a" },
				{ "id": "b", "kind": "host", "label": "b" },
				{ "id": "c", "kind": "host", "label": "c" }
			],
			"edges": [ { "from": "a", "to": "b" } ]
		}
	}`)
	_, err := ParseDocumentWithContext(context.Background(), raw, ValidateDocumentOptions{
		RejectOrphanNodesWhenEdgesExist: true,
	})
	if err == nil || !strings.Contains(err.Error(), "orphan") {
		t.Fatalf("expected orphan error, got %v", err)
	}
}

func TestValidateDocument_largeGraphParallelChunks(t *testing.T) {
	const n = 2500
	nodes := make([]Node, n)
	edges := make([]Edge, 0, n-1)
	for i := range n {
		id := fmt.Sprintf("n%d", i)
		nodes[i] = Node{ID: id, Kind: "host", Label: id}
		if i > 0 {
			edges = append(edges, Edge{From: fmt.Sprintf("n%d", i-1), To: id, Kind: "next"})
		}
	}
	doc := &Document{
		APIVersion: apiVersion,
		Kind:       kind,
		Metadata:   Metadata{GeneratedAt: "2024-01-01T00:00:00Z"},
		Spec:       GraphSpec{Phase: "plan", Nodes: nodes, Edges: edges},
	}
	if err := validateDocument(doc); err != nil {
		t.Fatal(err)
	}
}

func TestConstructFromDocument_ValidDocument(t *testing.T) {
	doc := &Document{
		APIVersion: "omnigraph/graph/v1",
		Kind:       "Graph",
		Metadata: Metadata{
			GeneratedAt: "2024-01-01T00:00:00Z",
			Project:     "test-project",
			Environment: "dev",
		},
		Spec: GraphSpec{
			Phase: "plan",
			Nodes: []Node{
				{
					ID:         "node1",
					Kind:       "host",
					Label:      "host1",
					State:      "active",
					Attributes: map[string]any{"ip": "192.168.1.1"},
				},
				{
					ID:         "node2",
					Kind:       "host",
					Label:      "host2",
					State:      "active",
					Attributes: map[string]any{"ip": "192.168.1.2"},
				},
			},
			Edges: []Edge{
				{
					From: "node1",
					To:   "node2",
					Kind: "connects",
				},
			},
		},
	}

	graph, err := ConstructFromDocument(doc)
	if err != nil {
		t.Fatalf("ConstructFromDocument failed: %v", err)
	}

	if graph.APIVersion != "omnigraph/graph/v1" {
		t.Errorf("expected apiVersion %q, got %q", "omnigraph/graph/v1", graph.APIVersion)
	}
	if len(graph.Spec.Nodes) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(graph.Spec.Nodes))
	}
	if len(graph.Spec.Edges) != 1 {
		t.Errorf("expected 1 edge, got %d", len(graph.Spec.Edges))
	}
}

func TestConstructFromDocument_InvalidDocument(t *testing.T) {
	doc := &Document{
		APIVersion: "omnigraph/graph/v1",
		Kind:       "Graph",
		Metadata: Metadata{
			GeneratedAt: "2024-01-01T00:00:00Z",
		},
		Spec: GraphSpec{
			Phase: "plan",
			Nodes: []Node{},
			Edges: []Edge{},
		},
	}

	_, err := ConstructFromDocument(doc)
	if err == nil {
		t.Errorf("expected error for empty nodes, got nil")
	}
}

func TestRoundTripJSON(t *testing.T) {
	original := &Document{
		APIVersion: "omnigraph/graph/v1",
		Kind:       "Graph",
		Metadata: Metadata{
			GeneratedAt: "2024-01-01T00:00:00Z",
			Project:     "test-project",
			Environment: "dev",
		},
		Spec: GraphSpec{
			Phase: "plan",
			Nodes: []Node{
				{
					ID:         "node1",
					Kind:       "host",
					Label:      "host1",
					State:      "active",
					Attributes: map[string]any{"ip": "192.168.1.1"},
				},
				{
					ID:         "node2",
					Kind:       "host",
					Label:      "host2",
					State:      "active",
					Attributes: map[string]any{"ip": "192.168.1.2"},
				},
			},
			Edges: []Edge{
				{
					From: "node1",
					To:   "node2",
					Kind: "connects",
				},
			},
		},
	}

	// Convert to JSON and back
	jsonData, err := json.MarshalIndent(original, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal JSON: %v", err)
	}

	parsed, err := ParseDocument(jsonData)
	if err != nil {
		t.Fatalf("ParseDocument failed: %v", err)
	}

	// Convert back to JSON
	roundTripJSON, err := json.MarshalIndent(parsed, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal round-trip JSON: %v", err)
	}

	if string(jsonData) != string(roundTripJSON) {
		t.Errorf("round-trip JSON does not match original")
	}
}

func assertTopologicalConstraints(t *testing.T, spec GraphSpec, order []string) {
	t.Helper()
	pos := make(map[string]int, len(order))
	for i, id := range order {
		pos[id] = i
	}
	seen := make(map[edgePair]struct{})
	for _, e := range spec.Edges {
		if e.From == "" || e.To == "" {
			continue
		}
		p := edgePair{from: e.From, to: e.To}
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		pf, pt := pos[e.From], pos[e.To]
		if pf >= pt {
			t.Fatalf("order violates edge %q -> %q: positions %d, %d", e.From, e.To, pf, pt)
		}
	}
}

func TestTopologicalOrder_Empty(t *testing.T) {
	got, err := TopologicalOrder(GraphSpec{})
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Fatalf("expected nil slice for empty spec, got %#v", got)
	}
}

func TestTopologicalOrder_Chain(t *testing.T) {
	spec := GraphSpec{
		Phase: "plan",
		Nodes: []Node{
			{ID: "a", Kind: "x", Label: "a"},
			{ID: "b", Kind: "x", Label: "b"},
			{ID: "c", Kind: "x", Label: "c"},
		},
		Edges: []Edge{{From: "a", To: "b"}, {From: "b", To: "c"}},
	}
	got, err := TopologicalOrder(spec)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(got, []string{"a", "b", "c"}) {
		t.Fatalf("got %v", got)
	}
	assertTopologicalConstraints(t, spec, got)
}

func TestTopologicalOrder_Diamond(t *testing.T) {
	spec := GraphSpec{
		Phase: "plan",
		Nodes: []Node{
			{ID: "a", Kind: "x", Label: "a"},
			{ID: "b", Kind: "x", Label: "b"},
			{ID: "c", Kind: "x", Label: "c"},
			{ID: "d", Kind: "x", Label: "d"},
		},
		Edges: []Edge{
			{From: "a", To: "b"},
			{From: "a", To: "c"},
			{From: "b", To: "d"},
			{From: "c", To: "d"},
		},
	}
	got, err := TopologicalOrder(spec)
	if err != nil {
		t.Fatal(err)
	}
	assertTopologicalConstraints(t, spec, got)
	if got[0] != "a" || got[3] != "d" {
		t.Fatalf("expected a first and d last, got %v", got)
	}
}

func TestTopologicalOrder_Disconnected(t *testing.T) {
	spec := GraphSpec{
		Phase: "plan",
		Nodes: []Node{
			{ID: "a", Kind: "x", Label: "a"},
			{ID: "b", Kind: "x", Label: "b"},
			{ID: "c", Kind: "x", Label: "c"},
			{ID: "d", Kind: "x", Label: "d"},
		},
		Edges: []Edge{{From: "a", To: "b"}, {From: "c", To: "d"}},
	}
	got, err := TopologicalOrder(spec)
	if err != nil {
		t.Fatal(err)
	}
	assertTopologicalConstraints(t, spec, got)
	if len(got) != 4 {
		t.Fatalf("len %d", len(got))
	}
}

func TestTopologicalOrdersPerWeakComponent_twoChains(t *testing.T) {
	spec := GraphSpec{
		Phase: "plan",
		Nodes: []Node{
			{ID: "a", Kind: "x", Label: "a"},
			{ID: "b", Kind: "x", Label: "b"},
			{ID: "c", Kind: "x", Label: "c"},
			{ID: "d", Kind: "x", Label: "d"},
		},
		Edges: []Edge{{From: "a", To: "b"}, {From: "c", To: "d"}},
	}
	orders, err := TopologicalOrdersPerWeakComponent(spec)
	if err != nil {
		t.Fatal(err)
	}
	if len(orders) != 2 {
		t.Fatalf("got %d components", len(orders))
	}
	if !slices.Equal(orders[0], []string{"a", "b"}) || !slices.Equal(orders[1], []string{"c", "d"}) {
		t.Fatalf("orders %#v", orders)
	}
}

func TestTopologicalOrder_ErrCycleIs(t *testing.T) {
	spec := GraphSpec{
		Phase: "plan",
		Nodes: []Node{
			{ID: "a", Kind: "x", Label: "a"},
			{ID: "b", Kind: "x", Label: "b"},
		},
		Edges: []Edge{{From: "a", To: "b"}, {From: "b", To: "a"}},
	}
	_, err := TopologicalOrder(spec)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrCycle) {
		t.Fatalf("errors.Is ErrCycle: %v", err)
	}
}

func TestValidateDocumentWithOptions_errorsIs(t *testing.T) {
	err := ValidateDocumentWithOptions(context.Background(), nil, ValidateDocumentOptions{})
	if !errors.Is(err, ErrNilDocument) {
		t.Fatalf("got %v", err)
	}
}

func TestTopologicalOrder_Cycle(t *testing.T) {
	spec := GraphSpec{
		Phase: "plan",
		Nodes: []Node{
			{ID: "a", Kind: "x", Label: "a"},
			{ID: "b", Kind: "x", Label: "b"},
			{ID: "c", Kind: "x", Label: "c"},
		},
		Edges: []Edge{{From: "a", To: "b"}, {From: "b", To: "c"}, {From: "c", To: "a"}},
	}
	_, err := TopologicalOrder(spec)
	if err == nil {
		t.Fatal("expected cycle error")
	}
	var ce *CycleError
	if !errors.As(err, &ce) {
		t.Fatalf("expected *CycleError, got %T: %v", err, err)
	}
	want := []string{"a", "b", "c", "a"}
	if !slices.Equal(ce.Path, want) {
		t.Fatalf("Path %v want %v", ce.Path, want)
	}
	if !strings.Contains(err.Error(), "a -> b -> c -> a") {
		t.Fatalf("Error() %q", err.Error())
	}
}

func TestTopologicalOrder_SelfLoop(t *testing.T) {
	spec := GraphSpec{
		Phase: "plan",
		Nodes: []Node{{ID: "a", Kind: "x", Label: "a"}},
		Edges: []Edge{{From: "a", To: "a"}},
	}
	_, err := TopologicalOrder(spec)
	if err == nil {
		t.Fatal("expected cycle error")
	}
	var ce *CycleError
	if !errors.As(err, &ce) {
		t.Fatalf("expected *CycleError, got %T: %v", err, err)
	}
	want := []string{"a", "a"}
	if !slices.Equal(ce.Path, want) {
		t.Fatalf("Path %v want %v", ce.Path, want)
	}
	if !strings.Contains(err.Error(), "a -> a") {
		t.Fatalf("Error() %q", err.Error())
	}
}

func TestTopologicalOrder_ParallelEdgesDeduped(t *testing.T) {
	spec := GraphSpec{
		Phase: "plan",
		Nodes: []Node{
			{ID: "a", Kind: "x", Label: "a"},
			{ID: "b", Kind: "x", Label: "b"},
		},
		Edges: []Edge{{From: "a", To: "b", Kind: "k1"}, {From: "a", To: "b", Kind: "k2"}},
	}
	got, err := TopologicalOrder(spec)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(got, []string{"a", "b"}) {
		t.Fatalf("got %v", got)
	}
}

func TestTopologicalOrder_DuplicateNodeID(t *testing.T) {
	spec := GraphSpec{
		Phase: "plan",
		Nodes: []Node{
			{ID: "a", Kind: "x", Label: "a"},
			{ID: "a", Kind: "x", Label: "a2"},
		},
	}
	_, err := TopologicalOrder(spec)
	if err == nil {
		t.Fatal("expected duplicate id error")
	}
}

func TestTopologicalOrder_UnknownEdgeEndpoint(t *testing.T) {
	spec := GraphSpec{
		Phase: "plan",
		Nodes: []Node{{ID: "a", Kind: "x", Label: "a"}},
		Edges: []Edge{{From: "a", To: "missing"}},
	}
	_, err := TopologicalOrder(spec)
	if err == nil {
		t.Fatal("expected unknown node error")
	}
}

func TestTopologicalOrder_MatchesNaiveSmall(t *testing.T) {
	cases := []GraphSpec{
		{
			Phase: "plan",
			Nodes: []Node{
				{ID: "a", Kind: "x", Label: "a"},
				{ID: "b", Kind: "x", Label: "b"},
				{ID: "c", Kind: "x", Label: "c"},
			},
			Edges: []Edge{{From: "a", To: "b"}, {From: "b", To: "c"}},
		},
		{
			Phase: "plan",
			Nodes: []Node{
				{ID: "a", Kind: "x", Label: "a"},
				{ID: "b", Kind: "x", Label: "b"},
				{ID: "c", Kind: "x", Label: "c"},
				{ID: "d", Kind: "x", Label: "d"},
			},
			Edges: []Edge{
				{From: "a", To: "b"},
				{From: "a", To: "c"},
				{From: "b", To: "d"},
				{From: "c", To: "d"},
			},
		},
	}
	for i, spec := range cases {
		fast, err := TopologicalOrder(spec)
		if err != nil {
			t.Fatalf("case %d fast: %v", i, err)
		}
		naive, err := topologicalOrderNaive(spec.Nodes, spec.Edges)
		if err != nil {
			t.Fatalf("case %d naive: %v", i, err)
		}
		if len(fast) != len(naive) {
			t.Fatalf("case %d len fast %d naive %d", i, len(fast), len(naive))
		}
		assertTopologicalConstraints(t, spec, fast)
		assertTopologicalConstraints(t, spec, naive)
	}
}

// topologicalOrderNaive is O(V²·E) worst-case; kept in tests only for benchmark comparison.
func topologicalOrderNaive(nodes []Node, edges []Edge) ([]string, error) {
	idSet := make(map[string]struct{}, len(nodes))
	for _, n := range nodes {
		if n.ID == "" {
			return nil, fmt.Errorf("empty node id")
		}
		idSet[n.ID] = struct{}{}
	}
	for _, e := range edges {
		if e.From == "" || e.To == "" {
			return nil, fmt.Errorf("empty edge endpoint")
		}
		if _, ok := idSet[e.From]; !ok {
			return nil, fmt.Errorf("unknown from %q", e.From)
		}
		if _, ok := idSet[e.To]; !ok {
			return nil, fmt.Errorf("unknown to %q", e.To)
		}
	}
	placed := make(map[string]bool, len(nodes))
	order := make([]string, 0, len(nodes))
	for len(placed) < len(nodes) {
		var pick string
		found := false
		for _, n := range nodes {
			if placed[n.ID] {
				continue
			}
			blocked := false
			for _, e := range edges {
				if e.To != n.ID {
					continue
				}
				if !placed[e.From] {
					blocked = true
					break
				}
			}
			if !blocked {
				pick = n.ID
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("cycle")
		}
		placed[pick] = true
		order = append(order, pick)
	}
	return order, nil
}

func benchTopoChain(V int) GraphSpec {
	nodes := make([]Node, V)
	edges := make([]Edge, V-1)
	for i := range V {
		id := fmt.Sprintf("n%d", i)
		nodes[i] = Node{ID: id, Kind: "x", Label: id}
		if i > 0 {
			edges[i-1] = Edge{From: fmt.Sprintf("n%d", i-1), To: id}
		}
	}
	return GraphSpec{Phase: "plan", Nodes: nodes, Edges: edges}
}

// benchTopoLayered builds L layers of width W with edges from each node to all nodes in the next layer.
func benchTopoLayered(L, W int) GraphSpec {
	nodes := make([]Node, 0, L*W)
	edges := make([]Edge, 0)
	idx := 0
	layerID := make([][]string, L)
	for l := range L {
		layerID[l] = make([]string, W)
		for w := range W {
			id := fmt.Sprintf("n%d", idx)
			layerID[l][w] = id
			nodes = append(nodes, Node{ID: id, Kind: "x", Label: id})
			idx++
		}
	}
	for l := 0; l < L-1; l++ {
		for _, from := range layerID[l] {
			for _, to := range layerID[l+1] {
				edges = append(edges, Edge{From: from, To: to})
			}
		}
	}
	return GraphSpec{Phase: "plan", Nodes: nodes, Edges: edges}
}

func BenchmarkTopologicalOrder_Chain10k(b *testing.B) {
	spec := benchTopoChain(10000)
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_, err := TopologicalOrder(spec)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTopologicalOrder_Layered(b *testing.B) {
	spec := benchTopoLayered(20, 25)
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_, err := TopologicalOrder(spec)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTopologicalOrderNaive_Chain500(b *testing.B) {
	spec := benchTopoChain(500)
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_, err := topologicalOrderNaive(spec.Nodes, spec.Edges)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTopologicalOrder_Chain500(b *testing.B) {
	spec := benchTopoChain(500)
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_, err := TopologicalOrder(spec)
		if err != nil {
			b.Fatal(err)
		}
	}
}
