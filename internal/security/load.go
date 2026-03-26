package security

import (
	"encoding/json"
	"fmt"
	"os"
)

// LoadDocument reads omnigraph/security/v1 JSON from a file.
func LoadDocument(path string) (*Document, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseDocument(b)
}

// ParseDocument unmarshals and validates apiVersion/kind.
func ParseDocument(b []byte) (*Document, error) {
	var d Document
	if err := json.Unmarshal(b, &d); err != nil {
		return nil, err
	}
	if d.APIVersion != apiVersion {
		return nil, fmt.Errorf("security: expected apiVersion %q, got %q", apiVersion, d.APIVersion)
	}
	if d.Kind != kind {
		return nil, fmt.Errorf("security: expected kind %q, got %q", kind, d.Kind)
	}
	return &d, nil
}
