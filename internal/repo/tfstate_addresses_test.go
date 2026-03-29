package repo

import (
	"reflect"
	"testing"
)

func TestTerraformResourceAddressesLegacy(t *testing.T) {
	raw := []byte(`{
  "version": 3,
  "resources": [
    {"module": "", "type": "aws_instance", "name": "x", "mode": "managed"},
    {"type": "aws_vpc", "name": "y", "mode": "data"}
  ]
}`)
	got, err := TerraformResourceAddresses(raw)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"aws_instance.x"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestTerraformResourceAddressesV4(t *testing.T) {
	raw := []byte(`{
  "version": 4,
  "values": {
    "root_module": {
      "resources": [
        {"address": "aws_instance.a", "mode": "managed", "type": "aws_instance", "name": "a"}
      ],
      "child_modules": [
        {
          "resources": [
            {"address": "module.net.aws_subnet.s", "mode": "managed", "type": "aws_subnet", "name": "s"}
          ]
        }
      ]
    }
  }
}`)
	got, err := TerraformResourceAddresses(raw)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"aws_instance.a", "module.net.aws_subnet.s"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}
