package security

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"time"
)

// Host runs read-only diagnostic commands for posture modules.
type Host interface {
	// Run executes argv[0] with argv[1:] on the target. Working directory is undefined for remote hosts.
	Run(ctx context.Context, argv []string) (stdout, stderr string, exitCode int, err error)
	Label() string
}

// LocalHost uses os/exec on the machine running OmniGraph.
type LocalHost struct {
	LabelOverride string
}

// Label implements Host.
func (l LocalHost) Label() string {
	if l.LabelOverride != "" {
		return l.LabelOverride
	}
	return "localhost"
}

// Run implements Host.
func (l LocalHost) Run(ctx context.Context, argv []string) (stdout, stderr string, exitCode int, err error) {
	if len(argv) == 0 {
		return "", "", -1, fmt.Errorf("empty argv")
	}
	cctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(cctx, argv[0], argv[1:]...)
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	runErr := cmd.Run()
	return outb.String(), errb.String(), exitCodeOf(runErr), nil
}

func exitCodeOf(err error) int {
	if err == nil {
		return 0
	}
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return ee.ExitCode()
	}
	return -1
}
