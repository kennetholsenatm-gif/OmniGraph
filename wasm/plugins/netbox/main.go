//go:build wasip1

// Package main: NetBox integration micro-container — reads omnigraph/integration-run/v1 from stdin,
// uses host import omnigraph.http_fetch for allowlisted REST calls, writes omnigraph/integration-result/v1 to stdout.
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

const fetchRespCap = 4 << 20 // 4 MiB response buffer for host JSON

type runEnvelope struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Spec       struct {
		Plugin               string         `json:"plugin"`
		AllowedFetchPrefixes []string       `json:"allowedFetchPrefixes"`
		Credentials          map[string]any `json:"credentials"`
		PluginConfig         map[string]any `json:"pluginConfig"`
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
	StatusCode int               `json:"statusCode"`
	Headers    map[string]string `json:"headers,omitempty"`
	BodyB64    string            `json:"bodyB64,omitempty"`
	Error      string            `json:"error,omitempty"`
}

type nbDeviceList struct {
	Results []struct {
		ID        int    `json:"id"`
		Name      string `json:"name"`
		PrimaryIP *struct {
			Address string `json:"address"`
		} `json:"primary_ip"`
		CustomFields map[string]any `json:"custom_fields"`
	} `json:"results"`
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
	if env.Spec.Plugin != "netbox" {
		writeFail("plugin must be netbox")
		return
	}
	if len(env.Spec.AllowedFetchPrefixes) == 0 {
		writeFail("allowedFetchPrefixes required")
		return
	}
	base := strings.TrimSuffix(strings.TrimSpace(env.Spec.AllowedFetchPrefixes[0]), "/")
	if base == "" {
		writeFail("empty base URL prefix")
		return
	}
	tok, _ := env.Spec.Credentials["token"].(string)
	if strings.TrimSpace(tok) == "" {
		writeFail("credentials.token required")
		return
	}

	devicesURL := base + "/api/dcim/devices/?limit=50"
	body, status, ferr := doFetch("GET", devicesURL, map[string]string{
		"Authorization": "Token " + tok,
		"Accept":        "application/json",
	}, "")
	if ferr != nil {
		writeFail(ferr.Error())
		return
	}
	if status < 200 || status >= 300 {
		writeFail(fmt.Sprintf("netbox devices: HTTP %d", status))
		return
	}
	var list nbDeviceList
	if err := json.Unmarshal(body, &list); err != nil {
		writeFail("netbox json: " + err.Error())
		return
	}

	now := time.Now().UTC().Format(time.RFC3339)
	records := make([]map[string]any, 0, len(list.Results))
	for _, d := range list.Results {
		ansible := ""
		if d.PrimaryIP != nil {
			ansible = strings.Split(d.PrimaryIP.Address, "/")[0]
		}
		names := []string{}
		if strings.TrimSpace(d.Name) != "" {
			names = append(names, d.Name)
		}
		rec := map[string]any{
			"id":          fmt.Sprintf("%d", d.ID),
			"recordType":  "host",
			"names":       names,
			"ansibleHost": ansible,
			"confidence":  "authoritative",
		}
		if d.CustomFields != nil {
			rec["attributes"] = d.CustomFields
		}
		records = append(records, rec)
	}

	snap := map[string]any{
		"apiVersion": "omnigraph/inventory-source/v1",
		"kind":       "InventorySnapshot",
		"metadata": map[string]any{
			"generatedAt": now,
			"source":      "netbox",
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
			"plugin":         "netbox",
			"idempotencyKey": strings.TrimSpace(env.Metadata.IdempotencyKey),
		},
		"spec": map[string]any{
			"status":            "ok",
			"errors":            []string{},
			"inventorySnapshot": snap,
		},
	}
	// omit empty idempotency key from metadata if blank
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
			"plugin":      "netbox",
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
	if len(reqJ) == 0 {
		return nil, 0, fmt.Errorf("empty fetch request")
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
	if int(n) > len(respBuf) {
		return nil, 0, fmt.Errorf("invalid fetch response length")
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
