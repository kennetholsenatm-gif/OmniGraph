package runner

import (
	"context"
	"time"
)

// Step describes an external tool invocation executed by a Runner.
type Step struct {
	Name    string
	Argv    []string
	Env     map[string]string
	Dir     string
	Timeout time.Duration
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
