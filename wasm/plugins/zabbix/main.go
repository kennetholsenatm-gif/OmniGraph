//go:build wasip1

// Package main: Zabbix integration micro-container — stdin integration-run/v1, host import http_fetch only, stdout integration-result/v1.
package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
	"unsafe"
)

//go:wasmimport omnigraph http_fetch
func wasmHTTPFetch(reqPtr, reqLen, respPtr, respCap uint32) int32

const fetchRespCap = 4 << 20

type runEnvelope struct {
	Spec struct {
		Plugin               string         `json:"plugin"`
		AllowedFetchPrefixes []string       `json:"allowedFetchPrefixes"`
		Credentials          map[string]any `json:"credentials"`
	} `json:"spec"`
	Metadata struct {
		IdempotencyKey string `json:"idempotencyKey"`
	} `json:"metadata"`
}

type fetchReqWire struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
	BodyB64 string            `json:"bodyB64,omitempty"`
}

type fetchRespWire struct {
	StatusCode int    `json:"statusCode"`
	BodyB64    string `json:"bodyB64,omitempty"`
	Error      string `json:"error,omitempty"`
}

type zbxRPCResp struct {
	Result []struct {
		HostID string `json:"hostid"`
		Host   string `json:"host"`
	} `json:"result"`
	Error *struct {
		Data string `json:"data"`
	} `json:"error"`
}

func main() {
	stdin, err := io.ReadAll(os.Stdin)
	if err != nil {
		os.Exit(1)
	}
	var env runEnvelope
	if err := json.Unmarshal(stdin, &env); err != nil {
		writeFail("parse stdin: " + err.Error())
		return
	}
	if env.Spec.Plugin != "zabbix" {
		writeFail("plugin must be zabbix")
		return
	}
	if len(env.Spec.AllowedFetchPrefixes) == 0 {
		writeFail("allowedFetchPrefixes required")
		return
	}
	base := strings.TrimSuffix(strings.TrimSpace(env.Spec.AllowedFetchPrefixes[0]), "/")
	tok, _ := env.Spec.Credentials["token"].(string)
	if strings.TrimSpace(tok) == "" {
		writeFail("credentials.token required")
		return
	}

	rpcURL := base + "/api_jsonrpc.php"
	rpcBody := map[string]any{
		"jsonrpc": "2.0",
		"method":  "host.get",
		"params": map[string]any{
			"output": []string{"hostid", "host"},
			"limit":  50,
		},
		"auth": tok,
		"id":   1,
	}
	rpcJSON, err := json.Marshal(rpcBody)
	if err != nil {
		writeFail(err.Error())
		return
	}
	bodyB64 := base64.StdEncoding.EncodeToString(rpcJSON)

	raw, status, ferr := doFetch("POST", rpcURL, map[string]string{
		"Content-Type": "application/json",
	}, bodyB64)
	if ferr != nil {
		writeFail(ferr.Error())
		return
	}
	if status < 200 || status >= 300 {
		writeFail(fmt.Sprintf("zabbix rpc: HTTP %d", status))
		return
	}
	var zr zbxRPCResp
	if err := json.Unmarshal(raw, &zr); err != nil {
		writeFail("zabbix json: " + err.Error())
		return
	}
	if zr.Error != nil {
		writeFail("zabbix: " + zr.Error.Data)
		return
	}

	now := time.Now().UTC().Format(time.RFC3339)
	records := make([]map[string]any, 0, len(zr.Result))
	for _, h := range zr.Result {
		records = append(records, map[string]any{
			"id":         h.HostID,
			"recordType": "host",
			"names":      []string{h.Host},
			"confidence": "high",
			"liveness":   "unknown",
			"links": map[string]string{
				"zabbixHostId": h.HostID,
			},
		})
	}

	snap := map[string]any{
		"apiVersion": "omnigraph/inventory-source/v1",
		"kind":       "InventorySnapshot",
		"metadata": map[string]any{
			"generatedAt": now,
			"source":      "zabbix",
		},
		"spec": map[string]any{
			"records": records,
		},
	}
	if k := strings.TrimSpace(env.Metadata.IdempotencyKey); k != "" {
		snap["metadata"].(map[string]any)["idempotencyKey"] = k
	}

	out := map[string]any{
		"apiVersion": "omnigraph/integration-result/v1",
		"kind":       "IntegrationResult",
		"metadata": map[string]any{
			"generatedAt":    now,
			"plugin":         "zabbix",
			"idempotencyKey": strings.TrimSpace(env.Metadata.IdempotencyKey),
		},
		"spec": map[string]any{
			"status":            "ok",
			"errors":            []string{},
			"inventorySnapshot": snap,
		},
	}
	if strings.TrimSpace(env.Metadata.IdempotencyKey) == "" {
		delete(out["metadata"].(map[string]any), "idempotencyKey")
	}
	b, err := json.Marshal(out)
	if err != nil {
		writeFail(err.Error())
		return
	}
	if _, err := os.Stdout.Write(b); err != nil {
		os.Exit(4)
	}
}

func writeFail(msg string) {
	now := time.Now().UTC().Format(time.RFC3339)
	out := map[string]any{
		"apiVersion": "omnigraph/integration-result/v1",
		"kind":       "IntegrationResult",
		"metadata": map[string]any{
			"generatedAt": now,
			"plugin":      "zabbix",
		},
		"spec": map[string]any{
			"status": "failed",
			"errors": []string{msg},
		},
	}
	b, _ := json.Marshal(out)
	_, _ = os.Stdout.Write(b)
	os.Exit(0)
}

func doFetch(method, url string, headers map[string]string, bodyB64 string) ([]byte, int, error) {
	req := fetchReqWire{Method: method, URL: url, Headers: headers, BodyB64: bodyB64}
	reqJ, err := json.Marshal(req)
	if err != nil {
		return nil, 0, err
	}
	respBuf := make([]byte, fetchRespCap)
	var reqPtr uint32
	if len(reqJ) > 0 {
		reqPtr = uint32(uintptr(unsafe.Pointer(&reqJ[0])))
	}
	respPtr := uint32(uintptr(unsafe.Pointer(&respBuf[0])))
	n := wasmHTTPFetch(reqPtr, uint32(len(reqJ)), respPtr, uint32(len(respBuf)))
	if n < 0 {
		return nil, 0, fmt.Errorf("http_fetch errno %d", n)
	}
	var fr fetchRespWire
	if err := json.Unmarshal(respBuf[:n], &fr); err != nil {
		return nil, 0, err
	}
	if fr.Error != "" && fr.StatusCode == 0 {
		return nil, 0, fmt.Errorf("host: %s", fr.Error)
	}
	raw, err := base64.StdEncoding.DecodeString(fr.BodyB64)
	if err != nil {
		return nil, fr.StatusCode, fmt.Errorf("decode body: %w", err)
	}
	return raw, fr.StatusCode, nil
}
