package hostops

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/kennetholsenatm-gif/omnigraph/internal/remotecmd"
	"golang.org/x/crypto/ssh"
)

var unitNameRe = regexp.MustCompile(`^[a-zA-Z0-9@._-]+\.(service|socket|timer|mount|path)$`)

// ListServiceUnits returns a bounded list of systemd service units (name + load + active + sub).
func ListServiceUnits(ctx context.Context, c *ssh.Client) (string, error) {
	if c == nil {
		return "", fmt.Errorf("nil ssh client")
	}
	sess, err := c.NewSession()
	if err != nil {
		return "", err
	}
	defer sess.Close()
	attachCtx(ctx, sess)
	var out strings.Builder
	sess.Stdout = &out
	cmd := remotecmd.RemoteShC([]string{"sh", "-c", "systemctl list-units --type=service --no-pager --no-legend 2>/dev/null | head -n 200"})
	if err := sess.Run(cmd); err != nil {
		var ee *ssh.ExitError
		if errors.As(err, &ee) && ee.ExitStatus() != 0 {
			return out.String(), fmt.Errorf("systemctl exited %d", ee.ExitStatus())
		}
		return out.String(), err
	}
	return out.String(), nil
}

// JournalTail returns recent journal lines for a unit (capped).
func JournalTail(ctx context.Context, c *ssh.Client, unit string, lines int) (string, error) {
	if c == nil {
		return "", fmt.Errorf("nil ssh client")
	}
	if !unitNameRe.MatchString(strings.TrimSpace(unit)) {
		return "", fmt.Errorf("invalid unit name")
	}
	if lines <= 0 {
		lines = 100
	}
	if lines > 500 {
		lines = 500
	}
	sess, err := c.NewSession()
	if err != nil {
		return "", err
	}
	defer sess.Close()
	attachCtx(ctx, sess)
	var out strings.Builder
	sess.Stdout = &out
	argv := []string{"journalctl", "-u", unit, "-n", fmt.Sprintf("%d", lines), "--no-pager"}
	cmd := remotecmd.RemoteShC(argv)
	if err := sess.Run(cmd); err != nil {
		var ee *ssh.ExitError
		if errors.As(err, &ee) {
			return out.String(), fmt.Errorf("journalctl exited %d", ee.ExitStatus())
		}
		return out.String(), err
	}
	return out.String(), nil
}

// RestartService runs systemctl restart on a unit (write-capable operations).
func RestartService(ctx context.Context, c *ssh.Client, unit string) (string, error) {
	if c == nil {
		return "", fmt.Errorf("nil ssh client")
	}
	if !unitNameRe.MatchString(strings.TrimSpace(unit)) {
		return "", fmt.Errorf("invalid unit name")
	}
	sess, err := c.NewSession()
	if err != nil {
		return "", err
	}
	defer sess.Close()
	attachCtx(ctx, sess)
	var out strings.Builder
	sess.Stdout = &out
	sess.Stderr = &out
	cmd := remotecmd.RemoteShC([]string{"systemctl", "restart", strings.TrimSpace(unit)})
	if err := sess.Run(cmd); err != nil {
		var ee *ssh.ExitError
		if errors.As(err, &ee) {
			return out.String(), fmt.Errorf("systemctl restart exited %d", ee.ExitStatus())
		}
		return out.String(), err
	}
	return out.String(), nil
}

func attachCtx(ctx context.Context, sess *ssh.Session) {
	if ctx == nil {
		return
	}
	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-ctx.Done():
			_ = sess.Close()
		case <-done:
		}
	}()
}
