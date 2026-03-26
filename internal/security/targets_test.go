package security

import "testing"

func TestParseAnsibleInventoryINI(t *testing.T) {
	const ini = `
[omnigraph]
web1 ansible_host=10.0.0.5 ansible_user=ubuntu ansible_port=2222
edge
`
	hosts := ParseAnsibleInventoryINI(ini)
	if len(hosts) != 2 {
		t.Fatalf("got %d hosts", len(hosts))
	}
	if hosts[0].Name != "web1" || hosts[0].Host != "10.0.0.5" || hosts[0].User != "ubuntu" || hosts[0].Port != "2222" {
		t.Fatalf("first=%+v", hosts[0])
	}
	if hosts[1].Name != "edge" || hosts[1].Host != "edge" {
		t.Fatalf("second=%+v", hosts[1])
	}
}

func TestSanitizeFilename(t *testing.T) {
	if g := SanitizeFilename("a/b"); g != "a_b" {
		t.Fatalf("got %q", g)
	}
	if SanitizeFilename(":::___") == "" {
		t.Fatal("empty")
	}
}
