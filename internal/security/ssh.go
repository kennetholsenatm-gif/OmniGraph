package security

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/kennetholsenatm-gif/omnigraph/internal/remotecmd"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// SSHDialConfig configures an SSH client for posture scans.
type SSHDialConfig struct {
	Host            string
	Port            string
	User            string
	KeyPath         string
	InsecureHostKey bool
	KnownHostsPath  string
}

// SSHHost runs modules over SSH (Linux targets).
type SSHHost struct {
	client *ssh.Client
	label  string
	mu     sync.Mutex
}

// DialSSH opens an SSH session to the target.
func DialSSH(cfg SSHDialConfig) (*SSHHost, error) {
	if cfg.Host == "" {
		return nil, fmt.Errorf("ssh: empty host")
	}
	if cfg.User == "" {
		return nil, fmt.Errorf("ssh: empty user")
	}
	port := cfg.Port
	if port == "" {
		port = "22"
	}
	keyPath := cfg.KeyPath
	if keyPath == "" {
		h, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("ssh key path: %w", err)
		}
		keyPath = h + string(os.PathSeparator) + ".ssh" + string(os.PathSeparator) + "id_rsa"
	}
	keyBytes, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("read ssh key %q: %w", keyPath, err)
	}
	signer, err := ssh.ParsePrivateKey(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("parse ssh key: %w", err)
	}
	var hostKeyCallback ssh.HostKeyCallback
	switch {
	case cfg.InsecureHostKey:
		hostKeyCallback = ssh.InsecureIgnoreHostKey()
	case strings.TrimSpace(cfg.KnownHostsPath) != "":
		hostKeyCallback, err = knownhosts.New(cfg.KnownHostsPath)
		if err != nil {
			return nil, fmt.Errorf("known_hosts %q: %w", cfg.KnownHostsPath, err)
		}
	default:
		return nil, fmt.Errorf("ssh: set --ssh-known-hosts or --insecure-ignore-host-key (host key verification required)")
	}
	clientConfig := &ssh.ClientConfig{
		User:            cfg.User,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: hostKeyCallback,
		Timeout:         20 * time.Second,
	}
	addr := net.JoinHostPort(cfg.Host, port)
	client, err := ssh.Dial("tcp", addr, clientConfig)
	if err != nil {
		return nil, fmt.Errorf("ssh dial %s: %w", addr, err)
	}
	label := cfg.User + "@" + cfg.Host
	if port != "22" {
		label += ":" + port
	}
	return &SSHHost{client: client, label: label}, nil
}

// Client returns the underlying SSH client for advanced callers (e.g. host operations). Close the SSHHost when finished.
func (h *SSHHost) Client() *ssh.Client {
	if h == nil {
		return nil
	}
	return h.client
}

// Close releases the SSH connection.
func (h *SSHHost) Close() error {
	if h == nil || h.client == nil {
		return nil
	}
	return h.client.Close()
}

// Label implements Host.
func (h *SSHHost) Label() string {
	if h == nil {
		return ""
	}
	return h.label
}

// Run implements Host.
func (h *SSHHost) Run(ctx context.Context, argv []string) (stdout, stderr string, exitCode int, err error) {
	if h == nil || h.client == nil {
		return "", "", -1, fmt.Errorf("ssh: nil client")
	}
	if len(argv) == 0 {
		return "", "", -1, fmt.Errorf("ssh: empty argv")
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	sess, err := h.client.NewSession()
	if err != nil {
		return "", "", -1, err
	}
	defer sess.Close()
	if ctx != nil {
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
	var outb, errb strings.Builder
	sess.Stdout = &outb
	sess.Stderr = &errb
	cmd := remotecmd.RemoteShC(argv)
	runErr := sess.Run(cmd)
	stdout = outb.String()
	stderr = errb.String()
	if runErr == nil {
		return stdout, stderr, 0, nil
	}
	var ee *ssh.ExitError
	if errors.As(runErr, &ee) {
		return stdout, stderr, ee.ExitStatus(), nil
	}
	return stdout, stderr, -1, runErr
}
