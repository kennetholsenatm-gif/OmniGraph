package schema

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/kennetholsenatm-gif/omnigraph/internal/enclave"
	"gopkg.in/yaml.v3"
)

// ValidateEnclaveRaw parses JSON or YAML enclave bytes and runs enclave Manager validation (requires/provides, etc.).
func ValidateEnclaveRaw(raw []byte) error {
	trim := bytes.TrimSpace(raw)
	if len(trim) == 0 {
		return fmt.Errorf("empty enclave document")
	}
	var manifest enclave.Enclave
	switch trim[0] {
	case '{', '[':
		if err := json.Unmarshal(trim, &manifest); err != nil {
			return fmt.Errorf("parse enclave json: %w", err)
		}
	default:
		if err := yaml.Unmarshal(trim, &manifest); err != nil {
			return fmt.Errorf("parse enclave yaml: %w", err)
		}
	}
	m := enclave.NewManager(".")
	return m.Validate(&manifest)
}
