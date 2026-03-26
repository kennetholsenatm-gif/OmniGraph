package runner

import (
	"context"
	"time"
)

// VolumeMount binds a host path into a container (Docker/Podman -v).
type VolumeMount struct {
	HostPath      string
	ContainerPath string
	ReadOnly      bool
}

// Step describes an external tool invocation executed by a Runner.
// For ContainerRunner, set ContainerImage and Mounts; Argv is the command run inside the image (see ADR 003: pass secrets via Env only, not .env files).
type Step struct {
	Name    string
	Argv    []string
	Env     map[string]string
	Dir     string
	Timeout time.Duration

	// Container execution (used by ContainerRunner; ignored by ExecRunner).
	ContainerImage   string
	ContainerWorkdir string // path inside container, default /workspace
	Mounts           []VolumeMount
	ReadOnlyRootFS   bool

	// RedactExtra lists additional substrings to redact from stdout/stderr (values not present in Env).
	RedactExtra []string
}

// Result captures process output and exit status.
type Result struct {
	ExitCode int
	Stdout   []byte
	Stderr   []byte
}

// Runner runs a Step in an isolated or local execution environment.
type Runner interface {
	Run(ctx context.Context, s Step) (*Result, error)
}
