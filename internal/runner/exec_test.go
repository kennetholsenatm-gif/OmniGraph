package runner

import (
	"context"
	"runtime"
	"strings"
	"testing"
)

func TestExecRunner_Run(t *testing.T) {
	var argv []string
	if runtime.GOOS == "windows" {
		argv = []string{"cmd", "/c", "echo", "ok"}
	} else {
		argv = []string{"echo", "ok"}
	}
	r := ExecRunner{}
	res, err := r.Run(context.Background(), Step{Argv: argv})
	if err != nil {
		t.Fatal(err)
	}
	if res.ExitCode != 0 {
		t.Fatalf("exit %d stderr=%q", res.ExitCode, res.Stderr)
	}
	if !strings.Contains(string(res.Stdout), "ok") {
		t.Fatalf("stdout=%q stderr=%q", res.Stdout, res.Stderr)
	}
}
