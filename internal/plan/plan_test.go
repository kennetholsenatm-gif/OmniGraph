package plan

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestProjectedHosts(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("testdata", "minimal-plan.json"))
	if err != nil {
		t.Fatal(err)
	}
	pj, err := Parse(b)
	if err != nil {
		t.Fatal(err)
	}
	got := ProjectedHosts(pj)
	want := map[string]string{
		"output.preview_ip": "10.0.5.99",
		"aws_instance.web":   "10.0.5.99",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("projected hosts mismatch\n got: %#v\nwant: %#v", got, want)
	}
}
