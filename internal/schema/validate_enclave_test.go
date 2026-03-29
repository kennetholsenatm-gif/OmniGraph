package schema

import "testing"

func TestValidateEnclaveRaw_json(t *testing.T) {
	raw := []byte(`{
  "apiVersion": "omnigraph/enclave/v1",
  "kind": "WasmEnclave",
  "metadata": { "name": "e1" },
  "spec": {
    "requires": ["in.api"],
    "provides": ["out.events"],
    "deploymentStrategy": "standalone",
    "replicas": 1,
    "runtime": {
      "engine": "wasmedge",
      "memoryLimitMb": 64,
      "deterministicExecution": true,
      "networkAccess": false,
      "filesystemAccess": "none",
      "maxInstances": 1
    },
    "trustBoundary": {
      "enrollment": "open",
      "certificateRotation": "24h",
      "auditLog": false
    },
    "cognitivePayload": {
      "sourceUri": "file:///x.wasm",
      "weightFormat": "f32",
      "inferenceMode": "batch"
    }
  }
}`)
	if err := ValidateEnclaveRaw(raw); err != nil {
		t.Fatal(err)
	}
}
