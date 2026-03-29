package graph

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"
)

func TestConcurrentGraph_UpsertAddEdgeSnapshot(t *testing.T) {
	g := NewConcurrentGraph()
	if err := g.UpsertNode(Node{ID: "a", Kind: "x", Label: "a"}); err != nil {
		t.Fatal(err)
	}
	if err := g.UpsertNode(Node{ID: "b", Kind: "x", Label: "b"}); err != nil {
		t.Fatal(err)
	}
	if err := g.AddEdge(Edge{From: "a", To: "b"}); err != nil {
		t.Fatal(err)
	}
	if err := g.AddEdge(Edge{From: "a", To: "b"}); err != nil {
		t.Fatal(err)
	}
	spec := g.Snapshot()
	if len(spec.Nodes) != 2 || len(spec.Edges) != 1 {
		t.Fatalf("nodes %d edges %d", len(spec.Nodes), len(spec.Edges))
	}
	if spec.Nodes[0].ID != "a" || spec.Nodes[1].ID != "b" {
		t.Fatalf("order %v %v", spec.Nodes[0].ID, spec.Nodes[1].ID)
	}
}

func TestConcurrentGraph_AddEdgeUnknownNode(t *testing.T) {
	g := NewConcurrentGraph()
	_ = g.UpsertNode(Node{ID: "a", Kind: "x", Label: "a"})
	if err := g.AddEdge(Edge{From: "a", To: "missing"}); err == nil {
		t.Fatal("expected error")
	}
}

func TestConcurrentGraph_GetNodeCopy(t *testing.T) {
	g := NewConcurrentGraph()
	_ = g.UpsertNode(Node{ID: "a", Kind: "x", Label: "a", Attributes: map[string]any{"k": 1}})
	n, ok := g.GetNode("a")
	if !ok {
		t.Fatal("missing")
	}
	n.Attributes["k"] = 2
	n2, _ := g.GetNode("a")
	if n2.Attributes["k"] != 1 {
		t.Fatalf("internal map leaked mutation: %v", n2.Attributes["k"])
	}
}

func TestConcurrentGraph_TopologicalOrder(t *testing.T) {
	g := NewConcurrentGraph()
	for _, id := range []string{"a", "b", "c"} {
		_ = g.UpsertNode(Node{ID: id, Kind: "x", Label: id})
	}
	_ = g.AddEdge(Edge{From: "a", To: "b"})
	_ = g.AddEdge(Edge{From: "b", To: "c"})
	order, err := g.TopologicalOrder()
	if err != nil {
		t.Fatal(err)
	}
	if len(order) != 3 {
		t.Fatalf("%v", order)
	}
}

// Exercise concurrent writers; use: go test -race ./internal/graph/... (needs CGO + C compiler).
func TestConcurrentGraph_parallelUpsert(t *testing.T) {
	const workers = 32
	const perWorker = 100
	g := NewConcurrentGraph()
	var wg sync.WaitGroup
	wg.Add(workers)
	for w := range workers {
		w := w
		go func() {
			defer wg.Done()
			for i := range perWorker {
				id := fmt.Sprintf("n-%d-%d", w, i)
				_ = g.UpsertNode(Node{ID: id, Kind: "host", Label: id})
			}
		}()
	}
	wg.Wait()
	if g.NodeCount() != workers*perWorker {
		t.Fatalf("count %d", g.NodeCount())
	}
}

func TestConcurrentGraph_snapshotConsistent(t *testing.T) {
	g := NewConcurrentGraph()
	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		for i := range 200 {
			id := fmt.Sprintf("w0-%d", i)
			_ = g.UpsertNode(Node{ID: id, Kind: "x", Label: id})
		}
	}()
	go func() {
		defer wg.Done()
		for i := range 200 {
			id := fmt.Sprintf("w1-%d", i)
			_ = g.UpsertNode(Node{ID: id, Kind: "x", Label: id})
		}
	}()
	go func() {
		defer wg.Done()
		for range 500 {
			_ = g.Snapshot()
		}
	}()
	wg.Wait()
	spec := g.Snapshot()
	if len(spec.Nodes) != 400 {
		t.Fatalf("nodes %d", len(spec.Nodes))
	}
	for _, e := range spec.Edges {
		t.Fatalf("unexpected edge %+v", e)
	}
}

func TestConcurrentGraph_parallelEdgesAfterNodes(t *testing.T) {
	g := NewConcurrentGraph()
	for _, id := range []string{"a", "b", "c", "d"} {
		_ = g.UpsertNode(Node{ID: id, Kind: "x", Label: id})
	}
	var wg sync.WaitGroup
	for _, e := range []Edge{{From: "a", To: "b"}, {From: "b", To: "c"}, {From: "c", To: "d"}} {
		wg.Add(1)
		go func(e Edge) {
			defer wg.Done()
			for range 50 {
				_ = g.AddEdge(e)
			}
		}(e)
	}
	wg.Wait()
	spec := g.Snapshot()
	if len(spec.Edges) != 3 {
		t.Fatalf("edges %d", len(spec.Edges))
	}
}

func TestConcurrentGraph_SnapshotDocument(t *testing.T) {
	g := NewConcurrentGraph()
	_ = g.UpsertNode(Node{ID: "x", Kind: "x", Label: "x"})
	doc := g.SnapshotDocument(Metadata{GeneratedAt: "t", Project: "p"})
	if doc.APIVersion != apiVersion || doc.Kind != kind || len(doc.Spec.Nodes) != 1 {
		t.Fatalf("%+v", doc)
	}
}

func TestConcurrentGraph_BatchVsSequentialSnapshot(t *testing.T) {
	const n = 500
	nodes := make([]Node, 0, n+1)
	nodes = append(nodes, Node{ID: "hub", Kind: "x", Label: "hub"})
	for i := range n {
		id := fmt.Sprintf("n%04d", i)
		nodes = append(nodes, Node{ID: id, Kind: "host", Label: id})
	}
	edges := make([]Edge, 0, n+n/10)
	for i := range n {
		id := fmt.Sprintf("n%04d", i)
		edges = append(edges, Edge{From: "hub", To: id, Kind: "links"})
	}
	for i := 0; i < n; i += 10 {
		id := fmt.Sprintf("n%04d", i)
		edges = append(edges, Edge{From: "hub", To: id, Kind: "links"})
	}

	gSeq := NewConcurrentGraph()
	for _, no := range nodes {
		if err := gSeq.UpsertNode(no); err != nil {
			t.Fatal(err)
		}
	}
	for _, e := range edges {
		_ = gSeq.AddEdge(e)
	}
	want := gSeq.Snapshot()

	gBat := NewConcurrentGraph()
	if err := gBat.BatchUpsertNodes(nodes); err != nil {
		t.Fatal(err)
	}
	if err := gBat.BatchAddEdges(edges); err != nil {
		t.Fatal(err)
	}
	got := gBat.Snapshot()

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("batch vs sequential snapshot differ: nodes %d vs %d edges %d vs %d",
			len(got.Nodes), len(want.Nodes), len(got.Edges), len(want.Edges))
	}
}

func TestConcurrentGraph_BatchUpsertNodes_validationNoMutation(t *testing.T) {
	g := NewConcurrentGraph()
	err := g.BatchUpsertNodes([]Node{{ID: "ok", Kind: "x", Label: "x"}, {ID: "", Kind: "x", Label: "bad"}})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if g.NodeCount() != 0 {
		t.Fatalf("graph mutated: count %d", g.NodeCount())
	}
}

func TestConcurrentGraph_BatchAddEdges_empty(t *testing.T) {
	g := NewConcurrentGraph()
	if err := g.BatchAddEdges(nil); err != nil {
		t.Fatal(err)
	}
	if err := g.BatchAddEdges([]Edge{}); err != nil {
		t.Fatal(err)
	}
}

func benchmarkConcurrentGraphNodes(b *testing.B, batch bool) {
	const N = 10000
	nodes := make([]Node, N)
	for i := range N {
		id := fmt.Sprintf("n%05d", i)
		nodes[i] = Node{ID: id, Kind: "host", Label: id}
	}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		g := NewConcurrentGraph()
		if batch {
			if err := g.BatchUpsertNodes(nodes); err != nil {
				b.Fatal(err)
			}
		} else {
			for _, n := range nodes {
				if err := g.UpsertNode(n); err != nil {
					b.Fatal(err)
				}
			}
		}
	}
}

func BenchmarkConcurrentGraph_UpsertSingle(b *testing.B) {
	benchmarkConcurrentGraphNodes(b, false)
}

func BenchmarkConcurrentGraph_BatchUpsertNodes(b *testing.B) {
	benchmarkConcurrentGraphNodes(b, true)
}

func benchmarkConcurrentGraphEdges(b *testing.B, batch bool) {
	const N = 10000
	nodes := make([]Node, N+1)
	nodes[0] = Node{ID: "hub", Kind: "x", Label: "hub"}
	edges := make([]Edge, N)
	for i := range N {
		id := fmt.Sprintf("n%05d", i)
		nodes[i+1] = Node{ID: id, Kind: "host", Label: id}
		edges[i] = Edge{From: "hub", To: id, Kind: "links"}
	}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		g := NewConcurrentGraph()
		if err := g.BatchUpsertNodes(nodes); err != nil {
			b.Fatal(err)
		}
		if batch {
			if err := g.BatchAddEdges(edges); err != nil {
				b.Fatal(err)
			}
		} else {
			for _, e := range edges {
				if err := g.AddEdge(e); err != nil {
					b.Fatal(err)
				}
			}
		}
	}
}

func BenchmarkConcurrentGraph_AddEdgeSingle(b *testing.B) {
	benchmarkConcurrentGraphEdges(b, false)
}

func BenchmarkConcurrentGraph_BatchAddEdges(b *testing.B) {
	benchmarkConcurrentGraphEdges(b, true)
}

func TestConcurrentGraph_degreesDedupeAndScan(t *testing.T) {
	g := NewConcurrentGraph()
	for _, id := range []string{"a", "b", "c"} {
		if err := g.UpsertNode(Node{ID: id, Kind: "x", Label: id}); err != nil {
			t.Fatal(err)
		}
	}
	_ = g.AddEdge(Edge{From: "a", To: "b"})
	_ = g.AddEdge(Edge{From: "a", To: "b"})
	_ = g.AddEdge(Edge{From: "b", To: "c"})
	inA, _ := g.InDegree("a")
	outA, _ := g.OutDegree("a")
	inB, _ := g.InDegree("b")
	outB, _ := g.OutDegree("b")
	if inA != 0 || outA != 1 || inB != 1 || outB != 1 {
		t.Fatalf("degrees a in=%d out=%d b in=%d out=%d", inA, outA, inB, outB)
	}
	spec := g.Snapshot()
	naiveIn := make(map[string]int)
	naiveOut := make(map[string]int)
	for _, e := range spec.Edges {
		naiveOut[e.From]++
		naiveIn[e.To]++
	}
	for _, id := range []string{"a", "b", "c"} {
		ni, _ := g.InDegree(id)
		no, _ := g.OutDegree(id)
		if ni != naiveIn[id] || no != naiveOut[id] {
			t.Fatalf("node %q: cached in=%d out=%d naive in=%d out=%d", id, ni, no, naiveIn[id], naiveOut[id])
		}
	}
}

func TestConcurrentGraph_batchVersusSequentialDegrees(t *testing.T) {
	nodes := []Node{
		{ID: "hub", Kind: "x", Label: "hub"},
		{ID: "n0", Kind: "x", Label: "n0"},
		{ID: "n1", Kind: "x", Label: "n1"},
	}
	edges := []Edge{{From: "hub", To: "n0"}, {From: "hub", To: "n1"}}
	gSeq := NewConcurrentGraph()
	for _, n := range nodes {
		_ = gSeq.UpsertNode(n)
	}
	for _, e := range edges {
		_ = gSeq.AddEdge(e)
	}
	gBat := NewConcurrentGraph()
	_ = gBat.BatchUpsertNodes(nodes)
	_ = gBat.BatchAddEdges(edges)
	for _, id := range []string{"hub", "n0", "n1"} {
		is, _ := gSeq.InDegree(id)
		io, _ := gSeq.OutDegree(id)
		ib, _ := gBat.InDegree(id)
		ob, _ := gBat.OutDegree(id)
		if is != ib || io != ob {
			t.Fatalf("node %q seq in=%d out=%d batch in=%d out=%d", id, is, io, ib, ob)
		}
	}
}

func TestConcurrentGraph_nilReceiverTypedError(t *testing.T) {
	var g *ConcurrentGraph
	if err := g.UpsertNode(Node{ID: "x"}); !errors.Is(err, ErrNilConcurrentGraph) {
		t.Fatalf("got %v", err)
	}
}
