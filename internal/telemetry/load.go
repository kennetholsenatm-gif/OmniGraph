package telemetry

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// LoadBundle reads a telemetry JSON file. If apiVersion is set, it must match APIVersion.
func LoadBundle(path string) (*Bundle, error) {
	if path == "" {
		return nil, fmt.Errorf("telemetry: empty path")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var b Bundle
	if err := json.Unmarshal(raw, &b); err != nil {
		return nil, fmt.Errorf("telemetry: %w", err)
	}
	if v := strings.TrimSpace(b.APIVersion); v != "" && v != APIVersion {
		return nil, fmt.Errorf("telemetry: expected apiVersion %q, got %q", APIVersion, b.APIVersion)
	}
	return &b, nil
}
