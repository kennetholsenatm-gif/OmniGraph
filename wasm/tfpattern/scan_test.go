package tfpattern

import "testing"

func TestScan_QuotedPassword(t *testing.T) {
	src := `resource "x" "y" {
  password = "supersecretvalue"
}`
	fs := Scan([]byte(src))
	if len(fs) != 1 {
		t.Fatalf("got %d findings, want 1: %#v", len(fs), fs)
	}
	if fs[0].Severity != "warning" || fs[0].Line < 2 {
		t.Fatalf("unexpected finding: %+v", fs[0])
	}
}

func TestScan_SkipsInterpolation(t *testing.T) {
	src := `password = "${var.pw}"`
	if len(Scan([]byte(src))) != 0 {
		t.Fatal("expected no findings for interpolation")
	}
}

func TestScan_AKIA(t *testing.T) {
	src := `x = "AKIAIOSFODNN7EXAMPLE"`
	fs := Scan([]byte(src))
	if len(fs) != 1 || fs[0].Severity != "error" {
		t.Fatalf("got %#v", fs)
	}
}
