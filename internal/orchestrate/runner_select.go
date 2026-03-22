package orchestrate

import (
	"strings"

	"github.com/kennetholsenatm-gif/omnigraph/internal/runner"
)

// NewRunner returns ExecRunner or ContainerRunner based on kind ("exec" or "container").
func NewRunner(kind, containerProgram string) runner.Runner {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "container":
		p := strings.TrimSpace(containerProgram)
		if p == "" {
			p = runner.DetectContainerRuntime()
		}
		if p == "" {
			p = "docker"
		}
		return runner.ContainerRunner{Program: p}
	default:
		return runner.ExecRunner{}
	}
}
