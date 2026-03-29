package graph

import (
	"encoding/json"
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