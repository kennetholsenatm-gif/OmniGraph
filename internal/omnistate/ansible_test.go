package omnistate

import (
	"context"
	"strings"
	"testing"
)

func TestAnsibleININormalizer_Basic(t *testing.T) {
	const src = `
[web]
h1 ansible_host=1.2.3.4
h2 ansible_host=5.6.7.8
`
	fr, err := AnsibleININormalizer{}.Normalize(context.Background(), NormalizerInput{
		Data: []byte(src),
		Ref:  SourceRef{Type: SourceAnsibleINI, Name: "inv.ini"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(fr.PartialErrors) > 0 {
		t.Fatalf("errors: %+v", fr.PartialErrors)
	}
	var hosts int
	for _, n := range fr.Nodes {
		if n.Kind == "ansible_host" {
			hosts++
		}
	}
	if hosts != 2 {
		t.Fatalf("hosts: %d", hosts)
	}
	if len(fr.Edges) < 2 {
		t.Fatalf("edges: %d", len(fr.Edges))
	}
}

func TestAnsibleYAMLNormalizer_AllChildren(t *testing.T) {
	const src = `
all:
  children:
    web:
      hosts:
        app1:
          ansible_host: 10.0.0.1
`
	fr, err := AnsibleYAMLNormalizer{}.Normalize(context.Background(), NormalizerInput{
		Data: []byte(src),
		Ref:  SourceRef{Type: SourceAnsibleYAML, Name: "inv.yml"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(fr.PartialErrors) > 0 {
		t.Fatalf("errors: %+v", fr.PartialErrors)
	}
	if !strings.Contains(strings.Join(nodeIDs(fr.Nodes), ","), "app1") {
		t.Fatalf("nodes: %+v", fr.Nodes)
	}
}

func TestTerraformNormalizer_ChildModule(t *testing.T) {
	const src = `{
  "values": {
    "root_module": {
      "resources": [],
      "child_modules": [
        {
          "address": "module.child",
          "resources": [
            {
              "address": "module.child.aws_instance.x",
              "mode": "managed",
              "type": "aws_instance",
              "name": "x",
              "values": {
                "depends_on": ["aws_security_group.y"],
                "private_ip": "10.1.1.1"
              }
            }
          ]
        }
      ]
    }
  }
}`
	fr, err := TerraformNormalizer{}.Normalize(context.Background(), NormalizerInput{
		Data: []byte(src),
		Ref:  SourceRef{Type: SourceTerraformState, Name: "s.tfstate"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(fr.PartialErrors) > 0 {
		t.Fatalf("errors: %+v", fr.PartialErrors)
	}
	if len(fr.Nodes) != 1 {
		t.Fatalf("nodes: %d", len(fr.Nodes))
	}
	if fr.Nodes[0].ID != "tf:module.child.aws_instance.x" {
		t.Fatalf("id: %q", fr.Nodes[0].ID)
	}
	if len(fr.Edges) != 1 || fr.Edges[0].Kind != "depends_on" {
		t.Fatalf("edges: %+v", fr.Edges)
	}
}

func nodeIDs(nodes []StateNode) []string {
	var out []string
	for _, n := range nodes {
		out = append(out, n.Label)
	}
	return out
}
