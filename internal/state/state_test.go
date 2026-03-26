package state

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestExtractHosts(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("testdata", "minimal.state.json"))
	if err != nil {
		t.Fatal(err)
	}
	st, err := Parse(b)
	if err != nil {
		t.Fatal(err)
	}
	got := ExtractHosts(st)
	want := map[string]string{
		"output.bastion_ip": "203.0.113.10",
		"aws_instance.web":  "10.0.5.21",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("hosts mismatch\n got: %#v\nwant: %#v", got, want)
	}
}
