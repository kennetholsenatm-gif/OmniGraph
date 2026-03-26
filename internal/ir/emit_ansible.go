package ir

import (
	"context"
	"fmt"
	"strings"

	"github.com/kennetholsenatm-gif/omnigraph/internal/inventory"
)

// ansibleInventoryBackend emits Ansible INI from spec.targets (logical id -> ansible_host).
type ansibleInventoryBackend struct{}

func (ansibleInventoryBackend) Format() string { return AnsibleInventoryINI }

func (ansibleInventoryBackend) Emit(_ context.Context, doc *Document) ([]Artifact, error) {
	if doc == nil {
		return nil, fmt.Errorf("ir: nil document")
	}
	hosts := make(map[string]string)
	for _, t := range doc.Spec.Targets {
		name := strings.TrimSpace(t.ID)
		if name == "" {
			continue
		}
		ah := strings.TrimSpace(t.AnsibleHost)
		if ah == "" {
			ah = name
		}
		hosts[name] = ah
	}
	ini := inventory.BuildINI(hosts)
	return []Artifact{{
		Path:        "inventory.ini",
		MediaType:   "text/plain; charset=utf-8",
		Description: "Ansible INI inventory from omnigraph/ir/v1 targets",
		Content:     []byte(ini),
	}}, nil
}
