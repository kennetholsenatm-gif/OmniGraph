package tfpattern

import (
	"testing"
)

func FuzzScan(f *testing.F) {
	f.Add([]byte(`password = "notshort"`))
	f.Add([]byte("resource \"x\" \"y\" {}\n# AKIA0123456789012345\n"))
	f.Fuzz(func(t *testing.T, data []byte) {
		_ = Scan(data)
	})
}
