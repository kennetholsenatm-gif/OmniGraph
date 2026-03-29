package emitter

import (
	"errors"
	"testing"

	"github.com/kennetholsenatm-gif/omnigraph/internal/graph"
)

func TestRelationsGraphSpec_disjointChainsTopologicalOrders(t *testing.T) {
	doc := &Document{
		APIVersion: "omnigraph/ir/v1",
		Kind:       "InfrastructureIntent",
		Metadata:   Metadata{Name: "x"},
		Spec: Spec{
			Components: []Component{
				{ID: "a", ComponentType: "t"},
				{ID: "b", ComponentType: "t"},
				{ID: "c", ComponentType: "t"},
				{ID: "d", ComponentType: "t"},
			},
			Relations: []Relation{
				{From: "a", To: "b", RelationType: "depends"},
				{From: "c", To: "d", RelationType: "depends"},
			},
		},
	}
	spec := RelationsGraphSpec(doc)
	orders, err := graph.TopologicalOrdersPerWeakComponent(spec)
	if err != nil {
		t.Fatal(err)
	}
	if len(orders) != 2 {
		t.Fatalf("components %d", len(orders))
	}
	if len(orders[0]) != 2 || len(orders[1]) != 2 {
		t.Fatalf("orders %#v", orders)
	}
	// Sorted by min node ID: a-chain before c-chain
	if orders[0][0] != "a" || orders[0][1] != "b" {
		t.Fatalf("first chain %v", orders[0])
	}
	if orders[1][0] != "c" || orders[1][1] != "d" {
		t.Fatalf("second chain %v", orders[1])
	}
	full, err := graph.TopologicalOrder(spec)
	if err != nil {
		t.Fatal(err)
	}
	if len(full) != 4 {
		t.Fatalf("full order %v", full)
	}
}

func TestRelationsGraphSpec_cycleErrorTyped(t *testing.T) {
	doc := &Document{
		Spec: Spec{
			Components: []Component{
				{ID: "a", ComponentType: "t"},
				{ID: "b", ComponentType: "t"},
			},
			Relations: []Relation{
				{From: "a", To: "b"},
				{From: "b", To: "a"},
			},
		},
	}
	spec := RelationsGraphSpec(doc)
	_, err := graph.TopologicalOrder(spec)
	if err == nil {
		t.Fatal("expected cycle")
	}
	if !errors.Is(err, graph.ErrCycle) {
		t.Fatalf("expected ErrCycle: %v", err)
	}
	var ce *graph.CycleError
	if !errors.As(err, &ce) {
		t.Fatalf("expected *CycleError: %T", err)
	}
}
