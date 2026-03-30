package runner

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/kennetholsenatm-gif/omnigraph/internal/safepath"
	"github.com/kennetholsenatm-gif/omnigraph/internal/schema"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

const (
	ogHostModule   = "omnigraph"
	ogHTTPFetch    = "http_fetch"
	defaultMaxReq  = 1 << 20 // 1 MiB guest → host fetch JSON
	defaultMaxHTTP = 4 << 20 // 4 MiB HTTP response body from upstream
)

var allowedHTTPMethods = map[string]bool{
	http.MethodGet: true, http.MethodPost: true, http.MethodPatch: true,
	http.MethodPut: true, http.MethodDelete: true,
}

// IntegrationHostConfig configures the integration micro-container host.
// AllowedFetchPrefixes is enforced on every http_fetch from the guest (and must match the integration-run envelope).
//
// WasmModuleRoot and WasmModuleRel locate the .wasm file under a trusted root directory (never pass a single
// user-controlled absolute path into the host without this join; see safepath.UnderRoot).
type IntegrationHostConfig struct {
	AllowedFetchPrefixes []string

	// WasmModuleRoot is the trusted base directory (e.g. workspace root or process cwd).
	WasmModuleRoot string
	// WasmModuleRel is a path relative to WasmModuleRoot to the plugin .wasm file.
	WasmModuleRel string

	// MaxGuestFetchJSONBytes caps the JSON request object read from guest linear memory (default defaultMaxReq).
	MaxGuestFetchJSONBytes int
	// MaxHTTPResponseBytes caps bytes read from upstream HTTP responses (default defaultMaxHTTP).
	MaxHTTPResponseBytes int64

	HTTPClient *http.Client
}

type ogFetchRequest struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
	BodyB64 string            `json:"bodyB64,omitempty"`
}

type ogFetchResponse struct {
	StatusCode int               `json:"statusCode"`
	Headers    map[string]string `json:"headers,omitempty"`
	BodyB64    string            `json:"bodyB64,omitempty"`
	Error      string            `json:"error,omitempty"`
}

// RunIntegrationPlugin runs a WASI integration plugin with stdin integration-run/v1 and captures stdout.
// The guest may call host import omnigraph.http_fetch for allowlisted HTTP only.
// Returned stdout is validated as omnigraph/integration-result/v1 (including nested inventory snapshot when present).
func RunIntegrationPlugin(ctx context.Context, cfg IntegrationHostConfig, stdin []byte, maxStdout int) ([]byte, error) {
	if maxStdout <= 0 {
		maxStdout = defaultMaxStdout
	}
	if strings.TrimSpace(cfg.WasmModuleRoot) == "" || strings.TrimSpace(cfg.WasmModuleRel) == "" {
		return nil, fmt.Errorf("integration host: WasmModuleRoot and WasmModuleRel are required")
	}
	prefixes := normalizePrefixes(cfg.AllowedFetchPrefixes)
	if len(prefixes) == 0 {
		return nil, fmt.Errorf("integration host: empty AllowedFetchPrefixes")
	}
	cfg.AllowedFetchPrefixes = prefixes

	var env map[string]any
	if err := json.Unmarshal(stdin, &env); err != nil {
		return nil, fmt.Errorf("integration stdin json: %w", err)
	}
	if err := schema.ValidateIntegrationRunV1(env); err != nil {
		return nil, err
	}
	echo, err := extractAllowedPrefixes(env)
	if err != nil {
		return nil, err
	}
	echo = normalizePrefixes(echo)
	if !slices.Equal(echo, prefixes) {
		return nil, fmt.Errorf("integration host: allowedFetchPrefixes mismatch between stdin and host config")
	}

	maxReq := cfg.MaxGuestFetchJSONBytes
	if maxReq <= 0 {
		maxReq = defaultMaxReq
	}
	maxHTTP := cfg.MaxHTTPResponseBytes
	if maxHTTP <= 0 {
		maxHTTP = defaultMaxHTTP
	}
	client := cfg.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 60 * time.Second}
	}

	absWasm, err := safepath.UnderRoot(cfg.WasmModuleRoot, cfg.WasmModuleRel)
	if err != nil {
		return nil, fmt.Errorf("wasm module path: %w", err)
	}
	wasmBytes, err := os.ReadFile(absWasm)
	if err != nil {
		return nil, fmt.Errorf("read wasm module: %w", err)
	}

	rt := wazero.NewRuntime(ctx)
	defer rt.Close(ctx)

	if _, err := wasi_snapshot_preview1.Instantiate(ctx, rt); err != nil {
		return nil, fmt.Errorf("wasi: %w", err)
	}

	h := &ogFetchHost{
		prefixes:   prefixes,
		maxReq:     maxReq,
		maxHTTP:    maxHTTP,
		httpClient: client,
	}
	if _, err := rt.NewHostModuleBuilder(ogHostModule).
		NewFunctionBuilder().
		WithFunc(func(ctx context.Context, m api.Module, reqPtr, reqLen, respPtr, respCap uint32) int32 {
			return h.handleFetch(ctx, m, reqPtr, reqLen, respPtr, respCap)
		}).
		Export(ogHTTPFetch).
		Instantiate(ctx); err != nil {
		return nil, fmt.Errorf("host module omnigraph: %w", err)
	}

	compiled, err := rt.CompileModule(ctx, wasmBytes)
	if err != nil {
		return nil, fmt.Errorf("compile wasm: %w", err)
	}
	defer compiled.Close(ctx)

	var stdout bytes.Buffer
	stdout.Grow(min(4096, maxStdout))
	lw := &limitWriter{w: &stdout, n: maxStdout}

	modCfg := wazero.NewModuleConfig().
		WithStdout(lw).
		WithStdin(bytes.NewReader(stdin)).
		WithArgs(filepath.Base(absWasm)).
		WithEnv("PATH", "").
		WithEnv("HOME", "")

	mod, err := rt.InstantiateModule(ctx, compiled, modCfg)
	if err != nil {
		return nil, fmt.Errorf("instantiate: %w", err)
	}
	defer mod.Close(ctx)
	if lw.truncated {
		return nil, fmt.Errorf("plugin stdout exceeded %d bytes", maxStdout)
	}

	out := stdout.Bytes()
	var result map[string]any
	if err := json.Unmarshal(out, &result); err != nil {
		return nil, fmt.Errorf("integration stdout json: %w", err)
	}
	if err := schema.ValidateIntegrationResultV1(result); err != nil {
		return nil, fmt.Errorf("integration result schema: %w", err)
	}
	return out, nil
}

type ogFetchHost struct {
	prefixes   []string
	maxReq     int
	maxHTTP    int64
	httpClient *http.Client
}

// fetch errno: negative int32 from guest-visible contract
const (
	ogErrBadMemory   int32 = -1
	ogErrReqTooLarge int32 = -2
	ogErrBadJSON     int32 = -3
	ogErrURLDenied   int32 = -4
	ogErrMethod      int32 = -5
	ogErrHTTP        int32 = -6
	ogErrRespTooBig  int32 = -7
)

func (h *ogFetchHost) handleFetch(ctx context.Context, m api.Module, reqPtr, reqLen, respPtr, respCap uint32) int32 {
	mem := m.Memory()
	if mem == nil {
		return ogErrBadMemory
	}
	if h.maxReq < 0 || int64(reqLen) > int64(h.maxReq) {
		return ogErrReqTooLarge
	}
	reqBytes, ok := mem.Read(reqPtr, reqLen)
	if !ok {
		return ogErrBadMemory
	}
	var fr ogFetchRequest
	if err := json.Unmarshal(reqBytes, &fr); err != nil {
		return ogErrBadJSON
	}
	method := strings.ToUpper(strings.TrimSpace(fr.Method))
	if !allowedHTTPMethods[method] {
		return ogErrMethod
	}
	u := strings.TrimSpace(fr.URL)
	if !urlAllowed(u, h.prefixes) {
		return ogErrURLDenied
	}

	var bodyReader io.Reader
	if fr.BodyB64 != "" {
		dec, err := base64.StdEncoding.DecodeString(fr.BodyB64)
		if err != nil {
			out, _ := json.Marshal(ogFetchResponse{Error: "invalid bodyB64"})
			return writeResp(mem, respPtr, respCap, out)
		}
		bodyReader = bytes.NewReader(dec)
	}

	req, err := http.NewRequestWithContext(ctx, method, u, bodyReader)
	if err != nil {
		out, _ := json.Marshal(ogFetchResponse{Error: err.Error()})
		return writeResp(mem, respPtr, respCap, out)
	}
	for k, v := range fr.Headers {
		if strings.TrimSpace(k) == "" {
			continue
		}
		req.Header.Set(k, v)
	}

	resp, err := h.httpClient.Do(req)
	if err != nil {
		out, _ := json.Marshal(ogFetchResponse{Error: err.Error()})
		return writeResp(mem, respPtr, respCap, out)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, h.maxHTTP+1))
	if err != nil {
		out, _ := json.Marshal(ogFetchResponse{Error: err.Error()})
		return writeResp(mem, respPtr, respCap, out)
	}
	if int64(len(body)) > h.maxHTTP {
		out, _ := json.Marshal(ogFetchResponse{Error: "upstream body exceeds cap"})
		return writeResp(mem, respPtr, respCap, out)
	}
	hdr := make(map[string]string)
	for k, vv := range resp.Header {
		if len(vv) > 0 {
			hdr[k] = vv[0]
		}
	}
	out, err := json.Marshal(ogFetchResponse{
		StatusCode: resp.StatusCode,
		Headers:    hdr,
		BodyB64:    base64.StdEncoding.EncodeToString(body),
	})
	if err != nil {
		return ogErrHTTP
	}
	return writeResp(mem, respPtr, respCap, out)
}

func writeResp(mem api.Memory, respPtr, respCap uint32, jsonOut []byte) int32 {
	n := len(jsonOut)
	if uint64(n) > uint64(respCap) {
		return ogErrRespTooBig
	}
	if n > math.MaxInt32 {
		return ogErrRespTooBig
	}
	if !mem.Write(respPtr, jsonOut) {
		return ogErrBadMemory
	}
	return int32(n) // n ≤ MaxInt32 and ≤ respCap
}

func normalizePrefixes(in []string) []string {
	var out []string
	for _, p := range in {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	slices.Sort(out)
	return slices.Compact(out)
}

func extractAllowedPrefixes(env map[string]any) ([]string, error) {
	spec, _ := env["spec"].(map[string]any)
	if spec == nil {
		return nil, fmt.Errorf("integration run: missing spec")
	}
	raw, ok := spec["allowedFetchPrefixes"]
	if !ok {
		return nil, fmt.Errorf("integration run: missing allowedFetchPrefixes")
	}
	arr, ok := raw.([]any)
	if !ok {
		return nil, fmt.Errorf("integration run: allowedFetchPrefixes not an array")
	}
	var s []string
	for _, v := range arr {
		str, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("integration run: allowedFetchPrefixes must be strings")
		}
		str = strings.TrimSpace(str)
		if str == "" {
			return nil, fmt.Errorf("integration run: empty prefix")
		}
		s = append(s, str)
	}
	return normalizePrefixes(s), nil
}

func urlAllowed(raw string, prefixes []string) bool {
	u, err := url.Parse(raw)
	if err != nil {
		return false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	full := u.String()
	for _, p := range prefixes {
		if strings.HasPrefix(full, p) {
			return true
		}
	}
	return false
}
