package remotecmd

import (
	"strings"
	"testing"
)

func TestRemoteShC(t *testing.T) {
	s := RemoteShC([]string{"uname", "-a"})
	if !strings.Contains(s, "exec") || !strings.Contains(s, "/bin/sh") {
		t.Fatalf("%q", s)
	}
}
