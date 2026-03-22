package runner

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
)

// ExecRunner runs argv on the local machine via os/exec (development and CI).
type ExecRunner struct{}

// Run executes the step's Argv[0] with remaining args; merges Env into the process environment.
func (ExecRunner) Run(ctx context.Context, s Step) (*Result, error) {
	if len(s.Argv) == 0 {
		return nil, fmt.Errorf("runner: empty argv")
	}
	if s.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, s.Timeout)
		defer cancel()
	}
	name := s.Argv[0]
	args := s.Argv[1:]
	cmd := exec.CommandContext(ctx, name, args...)
	if s.Dir != "" {
		cmd.Dir = s.Dir
	}
	cmd.Env = mergeEnv(os.Environ(), s.Env)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	res := &Result{
		Stdout: stdout.Bytes(),
		Stderr: stderr.Bytes(),
	}
	if err == nil {
		res.ExitCode = 0
		return res, nil
	}
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		res.ExitCode = ee.ExitCode()
		return res, nil
	}
	return nil, fmt.Errorf("runner: %w", err)
}

func mergeEnv(base []string, extra map[string]string) []string {
	if len(extra) == 0 {
		return base
	}
	out := append([]string(nil), base...)
	for k, v := range extra {
		out = append(out, fmt.Sprintf("%s=%s", k, v))
	}
	return out
}
