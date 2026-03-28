package reconciler

import (
	"context"
	"strings"
	"testing"
)

func TestAnsibleInventoryEmit(t *testing.T) {
	doc := &Document{
		APIVersion: "omnigraph/ir/v1",
		Kind:       "InfrastructureIntent",
		Metadata:   Metadata{Name: "t"},
		Spec: Spec{
			Targets: []Target{
				{ID: "web1", AnsibleHost: "10.0.0.5"},
				{ID: "db1", AnsibleHost: "10.0.0.6"},
			},
			Components: []Component{},
			Relations:  []Relation{},
		},
	}
	var b ansibleInventoryBackend
	arts, err := b.Emit(context.Background(), doc)
	if err != nil {
		t.Fatal(err)
	}
	if len(arts) != 1 {
		t.Fatalf("artifacts %d", len(arts))
	}
	body := string(arts[0].Content)
	if !strings.Contains(body, "[omnigraph]") {
		t.Fatal(body)
	}
	if !strings.Contains(body, "ansible_host=10.0.0.5") {
		t.Fatal(body)
	}
}
