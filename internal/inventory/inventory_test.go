package inventory

import (
	"strings"
	"testing"

	"github.com/kennetholsenatm-gif/omnigraph/internal/state"
)

func TestBuildINI(t *testing.T) {
	st, err := state.Parse([]byte(`{
  "values": {
    "root_module": {
      "resources": [
        {
          "address": "aws_instance.web",
          "mode": "managed",
          "type": "aws_instance",
          "name": "web",
          "values": { "public_ip": "1.2.3.4" }
        }
      ]
    }
  }
}`))
	if err != nil {
		t.Fatal(err)
	}
	s := BuildINI(state.ExtractHosts(st))
	if !strings.Contains(s, "[omnigraph]") {
		t.Fatal(s)
	}
	if !strings.Contains(s, "ansible_host=1.2.3.4") {
		t.Fatal(s)
	}
}
