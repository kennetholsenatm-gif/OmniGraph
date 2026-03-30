package inventory

import (
	"reflect"
	"testing"
)

func TestParseAnsibleINIHosts(t *testing.T) {
	raw := []byte(`[web]
app1 ansible_host=10.0.0.1
app2
`)
	got, err := ParseAnsibleINIHosts(raw)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"app1", "app2"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestParseAnsibleINIHostsLineFormat(t *testing.T) {
	raw := []byte(`# c
10.0.0.1
`)
	got, err := ParseAnsibleINIHosts(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != "10.0.0.1" {
		t.Fatalf("got %v", got)
	}
}

func TestParseAnsibleYAMLHosts(t *testing.T) {
	raw := []byte(`all:
  hosts:
    - h1
    - h2
`)
	got, err := ParseAnsibleYAMLHosts(raw)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"h1", "h2"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}
