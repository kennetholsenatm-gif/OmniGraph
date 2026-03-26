package enclave

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// GraphConnector manages the bridge between OmniGraph and QminiWasm-core
type GraphConnector struct {
	mu          sync.RWMutex
	manager     *Manager
	graphs      map[string]*EnclaveGraph
	subscribers map[string][]GraphEventHandler
	connected   bool
}

// EnclaveGraph represents a graph of interconnected enclaves
type EnclaveGraph struct {
	GraphID   string            `json:"graphId"`
	Nodes     []GraphNode       `json:"nodes"`
	Edges     []GraphEdge       `json:"edges"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	UpdatedAt time.Time         `json:"updatedAt"`
}

// GraphNode represents a node in the enclave graph
type GraphNode struct {
	ID           string            `json:"id"`
	Type         string            `json:"type"` // enclave, target, service
	Properties   map[string]string `json:"properties,omitempty"`
	Dependencies []string          `json:"dependencies,omitempty"`
}

// GraphEdge represents a connection between nodes
type GraphEdge struct {
	From         string            `json:"from"`
	To           string            `json:"to"`
	Relationship string            `json:"relationship"` // depends_on, communicates_with, routes_to
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// GraphEventHandler handles graph events
type GraphEventHandler func(event GraphEvent) error

// GraphEvent represents a change in the enclave graph
type GraphEvent struct {
	Type      string                 `json:"type"` // node_added, node_removed, edge_added, edge_removed, status_changed
	GraphID   string                 `json:"graphId"`
	NodeID    string                 `json:"nodeId,omitempty"`
	Edge      *GraphEdge             `json:"edge,omitempty"`
	Status    string                 `json:"status,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// NewGraphConnector creates a new graph connector
func NewGraphConnector(manager *Manager) *GraphConnector {
	return &GraphConnector{
		manager:     manager,
		graphs:      make(map[string]*EnclaveGraph),
		subscribers: make(map[string][]GraphEventHandler),
	}
}

// Connect establishes connection to the enclave manager
func (gc *GraphConnector) Connect(ctx context.Context) error {
	gc.mu.Lock()
	defer gc.mu.Unlock()

	gc.connected = true
	return nil
}

// Disconnect closes the connection
func (gc *GraphConnector) Disconnect() error {
	gc.mu.Lock()
	defer gc.mu.Unlock()

	gc.connected = false
	return nil
}

// SyncGraph synchronizes a graph with the enclave manager
func (gc *GraphConnector) SyncGraph(ctx context.Context, graph *EnclaveGraph) error {
	gc.mu.Lock()
	defer gc.mu.Unlock()

	if !gc.connected {
		return fmt.Errorf("not connected to enclave manager")
	}

	// Store the graph
	gc.graphs[graph.GraphID] = graph

	// Process nodes
	for _, node := range graph.Nodes {
		if node.Type == "enclave" {
			// Check if enclave exists
			_, err := gc.manager.GetStatus(node.ID)
			if err != nil {
				// Enclave doesn't exist, would need to create it
				gc.emitEvent(GraphEvent{
					Type:      "node_added",
					GraphID:   graph.GraphID,
					NodeID:    node.ID,
					Timestamp: time.Now(),
					Metadata: map[string]interface{}{
						"nodeType": node.Type,
						"action":   "create_required",
					},
				})
			}
		}
	}

	return nil
}

// GetGraph retrieves a graph by ID
func (gc *GraphConnector) GetGraph(graphID string) (*EnclaveGraph, error) {
	gc.mu.RLock()
	defer gc.mu.RUnlock()

	graph, exists := gc.graphs[graphID]
	if !exists {
		return nil, fmt.Errorf("graph not found: %s", graphID)
	}

	return graph, nil
}

// ListGraphs returns all managed graphs
func (gc *GraphConnector) ListGraphs() []*EnclaveGraph {
	gc.mu.RLock()
	defer gc.mu.RUnlock()

	graphs := make([]*EnclaveGraph, 0, len(gc.graphs))
	for _, g := range gc.graphs {
		graphs = append(graphs, g)
	}
	return graphs
}

// DeleteGraph removes a graph
func (gc *GraphConnector) DeleteGraph(graphID string) error {
	gc.mu.Lock()
	defer gc.mu.Unlock()

	if _, exists := gc.graphs[graphID]; !exists {
		return fmt.Errorf("graph not found: %s", graphID)
	}

	delete(gc.graphs, graphID)
	return nil
}

// Subscribe registers an event handler for graph events
func (gc *GraphConnector) Subscribe(eventType string, handler GraphEventHandler) error {
	gc.mu.Lock()
	defer gc.mu.Unlock()

	gc.subscribers[eventType] = append(gc.subscribers[eventType], handler)
	return nil
}

// Unsubscribe removes an event handler
func (gc *GraphConnector) Unsubscribe(eventType string, handler GraphEventHandler) error {
	gc.mu.Lock()
	defer gc.mu.Unlock()

	handlers := gc.subscribers[eventType]
	for i, h := range handlers {
		// Compare function pointers (simplified)
		if fmt.Sprintf("%p", h) == fmt.Sprintf("%p", handler) {
			gc.subscribers[eventType] = append(handlers[:i], handlers[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("handler not found")
}

// emitEvent sends an event to all subscribers
func (gc *GraphConnector) emitEvent(event GraphEvent) {
	handlers := gc.subscribers[event.Type]
	for _, handler := range handlers {
		go func(h GraphEventHandler, e GraphEvent) {
			if err := h(e); err != nil {
				fmt.Printf("Error in graph event handler: %v\n", err)
			}
		}(handler, event)
	}

	// Also emit to "all" subscribers
	allHandlers := gc.subscribers["all"]
	for _, handler := range allHandlers {
		go func(h GraphEventHandler, e GraphEvent) {
			if err := h(e); err != nil {
				fmt.Printf("Error in graph event handler: %v\n", err)
			}
		}(handler, event)
	}
}

// GetNodeStatus returns the status of a node in the graph
func (gc *GraphConnector) GetNodeStatus(nodeID string) (map[string]interface{}, error) {
	gc.mu.RLock()
	defer gc.mu.RUnlock()

	// Check if node is an enclave
	status, err := gc.manager.GetStatus(nodeID)
	if err == nil {
		return map[string]interface{}{
			"nodeId": nodeID,
			"type":   "enclave",
			"status": status,
		}, nil
	}

	return nil, fmt.Errorf("node not found: %s", nodeID)
}

// ExportGraph exports a graph as JSON
func (gc *GraphConnector) ExportGraph(graphID string) ([]byte, error) {
	gc.mu.RLock()
	defer gc.mu.RUnlock()

	graph, exists := gc.graphs[graphID]
	if !exists {
		return nil, fmt.Errorf("graph not found: %s", graphID)
	}

	return json.MarshalIndent(graph, "", "  ")
}

// ImportGraph imports a graph from JSON
func (gc *GraphConnector) ImportGraph(ctx context.Context, data []byte) error {
	var graph EnclaveGraph
	if err := json.Unmarshal(data, &graph); err != nil {
		return fmt.Errorf("failed to unmarshal graph: %w", err)
	}

	return gc.SyncGraph(ctx, &graph)
}
