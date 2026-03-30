package plan

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMutationAddresses(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "plan.json")
	raw := `{
  "resource_changes": [
    {
      "address": "aws_instance.a",
      "mode": "managed",
      "type": "aws_instance",
      "name": "a",
      "change": { "actions": ["no-op"] }
    },
    {
      "address": "aws_instance.b",
      "mode": "managed",
      "type": "aws_instance",
      "name": "b",
      "change": { "actions": ["update"] }
    },
    {
      "address": "data.aws_ami.x",
      "mode": "data",
      "type": "aws_ami",
      "name": "x",
      "change": { "actions": ["read"] }
    }
  ]
}`
	if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
		t.Fatal(err)
	}
	pj, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	got := MutationAddresses(pj)
	if len(got) != 1 || got[0] != "aws_instance.b" {
		t.Fatalf("MutationAddresses = %v want [aws_instance.b]", got)
	}
}

func TestMutationSeedAddresses_fallback(t *testing.T) {
	pj := &JSON{
		PlannedValues: &PlannedValues{
			RootModule: &RootModule{
				Resources: []Resource{
					{
						Address: "aws_instance.x",
						Mode:    "managed",
						Type:    "aws_instance",
						Name:    "x",
						Values:  map[string]any{"public_ip": "1.2.3.4"},
					},
				},
			},
		},
	}
	got := MutationSeedAddresses(pj)
	if len(got) != 1 || got[0] != "aws_instance.x" {
		t.Fatalf("MutationSeedAddresses = %v", got)
	}
}
