package enclave

import "testing"

func TestValidateContract_requiresProvides(t *testing.T) {
	e := &Enclave{
		APIVersion: "omnigraph/enclave/v1",
		Kind:       "WasmEnclave",
		Metadata:   EnclaveMetadata{Name: "x"},
		Spec: EnclaveSpec{
			Requires:           []string{"api.in"},
			Provides:           []string{"events.out"},
			DeploymentStrategy: "standalone",
			Replicas:           1,
			Runtime: RuntimeConfig{
				Engine: "wasmedge", MemoryLimitMb: 128, DeterministicExecution: true,
				NetworkAccess: false, FilesystemAccess: "none", MaxInstances: 1,
			},
			TrustBoundary: TrustBoundary{
				Enrollment: "open", CertificateRotation: "24h", AuditLog: true,
				AllowedPeers: []string{"api.in"},
			},
			CognitivePayload: CognitivePayload{
				SourceURI: "file:///m.wasm", WeightFormat: "f32", InferenceMode: "batch",
			},
		},
	}
	if err := ValidateContract(e); err != nil {
		t.Fatal(err)
	}
}

func TestValidateContract_peerURI(t *testing.T) {
	e := &Enclave{
		APIVersion: "omnigraph/enclave/v1",
		Kind:       "WasmEnclave",
		Metadata:   EnclaveMetadata{Name: "x"},
		Spec: EnclaveSpec{
			Requires:           []string{"upstream"},
			Provides:           []string{"downstream"},
			DeploymentStrategy: "standalone",
			Replicas:           1,
			Runtime: RuntimeConfig{
				Engine: "wasmedge", MemoryLimitMb: 128, DeterministicExecution: true,
				NetworkAccess: false, FilesystemAccess: "none", MaxInstances: 1,
			},
			TrustBoundary:    TrustBoundary{Enrollment: "open", CertificateRotation: "24h", AuditLog: false},
			CognitivePayload: CognitivePayload{SourceURI: "file:///m.wasm", WeightFormat: "f32", InferenceMode: "batch"},
			Environment: map[string]string{
				"X": "peer://upstream/path",
			},
		},
	}
	if err := ValidateContract(e); err != nil {
		t.Fatal(err)
	}
	e.Spec.Environment["X"] = "peer://stranger/path"
	if err := ValidateContract(e); err == nil {
		t.Fatal("expected error for undeclared peer")
	}
}
