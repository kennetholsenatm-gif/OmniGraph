package runner

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

const defaultMaxStdout = 4 << 20 // 4 MiB cap on plugin stdout (deny runaway guests)

// RunWASIParser executes a WASI wasm module with stdin-only input and captures stdout.
// No host filesystem or network is exposed to the guest beyond empty WASI stubs.
func RunWASIParser(ctx context.Context, wasmPath string, stdin []byte) ([]byte, error) {
	return RunWASIParserLimit(ctx, wasmPath, stdin, defaultMaxStdout)
}

// RunWASIParserLimit is like RunWASIParser but caps captured stdout size.
func RunWASIParserLimit(ctx context.Context, wasmPath string, stdin []byte, maxStdout int) ([]byte, error) {
	if maxStdout <= 0 {
		maxStdout = defaultMaxStdout
	}
	abs, err := filepath.Abs(wasmPath)
	if err != nil {
		return nil, err
	}
	wasmBytes, err := os.ReadFile(abs)
	if err != nil {
		return nil, err
	}

	rt := wazero.NewRuntime(ctx)
	defer rt.Close(ctx)

	if _, err := wasi_snapshot_preview1.Instantiate(ctx, rt); err != nil {
		return nil, fmt.Errorf("wasi: %w", err)
	}

	compiled, err := rt.CompileModule(ctx, wasmBytes)
	if err != nil {
		return nil, fmt.Errorf("compile wasm: %w", err)
	}
	defer compiled.Close(ctx)

	var stdout bytes.Buffer
	stdout.Grow(min(4096, maxStdout))
	lw := &limitWriter{w: &stdout, n: maxStdout}

	config := wazero.NewModuleConfig().
		WithStdout(lw).
		WithStdin(bytes.NewReader(stdin)).
		WithArgs(filepath.Base(abs)).
		WithEnv("PATH", "").
		WithEnv("HOME", "")

	mod, err := rt.InstantiateModule(ctx, compiled, config)
	if err != nil {
		return nil, fmt.Errorf("instantiate: %w", err)
	}
	defer mod.Close(ctx)
	if lw.truncated {
		return nil, fmt.Errorf("plugin stdout exceeded %d bytes", maxStdout)
	}
	return stdout.Bytes(), nil
}

type limitWriter struct {
	w         io.Writer
	n         int
	truncated bool
}

func (l *limitWriter) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	if l.n <= 0 {
		l.truncated = true
		return 0, io.ErrShortWrite
	}
	if len(p) > l.n {
		p = p[:l.n]
		l.truncated = true
	}
	n, err := l.w.Write(p)
	l.n -= n
	return n, err
}
