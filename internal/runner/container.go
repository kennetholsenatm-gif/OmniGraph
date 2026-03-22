package runner

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"sort"
	"strings"
)

// ContainerRunner runs steps via docker or podman. Step.ContainerImage must be set; Argv is appended after the image name (container entrypoint/command).
// Environment variables are passed with -e only (no --env-file), aligning with memory-only secret injection (ADR 003).
type ContainerRunner struct {
	// Program is the container CLI: "docker" or "podman".
	Program string
}

// Run implements Runner.
func (r ContainerRunner) Run(ctx context.Context, s Step) (*Result, error) {
	if s.ContainerImage == "" {
		return nil, fmt.Errorf("container runner: Step.ContainerImage is required")
	}
	if len(s.Argv) == 0 {
		return nil, fmt.Errorf("container runner: empty Argv")
	}
	prog := r.Program
	if prog == "" {
		prog = "docker"
	}
	if s.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, s.Timeout)
		defer cancel()
	}
	args, err := BuildContainerRunArgs(prog, s)
	if err != nil {
		return nil, err
	}
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
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
	return nil, fmt.Errorf("container runner: %w", err)
}

// BuildContainerRunArgs returns [program, arg1, ...] for `program run ...` (for tests and inspection).
func BuildContainerRunArgs(program string, s Step) ([]string, error) {
	if s.ContainerImage == "" {
		return nil, fmt.Errorf("container runner: ContainerImage required")
	}
	if program == "" {
		program = "docker"
	}
	wd := s.ContainerWorkdir
	if wd == "" {
		wd = "/workspace"
	}
	out := []string{program, "run", "--rm", "-i", "--workdir", wd}
	if s.ReadOnlyRootFS {
		out = append(out, "--read-only")
	}
	for _, m := range s.Mounts {
		if m.HostPath == "" || m.ContainerPath == "" {
			return nil, fmt.Errorf("container runner: invalid mount %+v", m)
		}
		mode := "rw"
		if m.ReadOnly {
			mode = "ro"
		}
		out = append(out, "-v", fmt.Sprintf("%s:%s:%s", m.HostPath, m.ContainerPath, mode))
	}
	keys := sortedEnvKeys(s.Env)
	for _, k := range keys {
		v := s.Env[k]
		if strings.ContainsRune(v, '\x00') {
			return nil, fmt.Errorf("container runner: env value for %q contains NUL", k)
		}
		out = append(out, "-e", fmt.Sprintf("%s=%s", k, v))
	}
	out = append(out, s.ContainerImage)
	out = append(out, s.Argv...)
	return out, nil
}

func sortedEnvKeys(m map[string]string) []string {
	if len(m) == 0 {
		return nil
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// DetectContainerRuntime returns podman or docker if found on PATH (podman preferred).
func DetectContainerRuntime() string {
	for _, name := range []string{"podman", "docker"} {
		path, err := exec.LookPath(name)
		if err == nil && path != "" {
			return name
		}
	}
	return ""
}
