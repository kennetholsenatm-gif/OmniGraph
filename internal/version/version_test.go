package version

import "testing"

func TestString(t *testing.T) {
	if got := String(); got == "" {
		t.Fatal("String() returned empty version")
	}
}
