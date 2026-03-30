package enclave

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// Manager handles enclave lifecycle operations
type Manager struct {
	mu        sync.RWMutex
	enclaves  map[string]*Enclave
	providers map[string]Provider
	baseDir   string
}

// Provider interface for enclave backends
type Provider interface {
	Create(ctx context.Context, enclave *Enclave) error
	Start(ctx context.Context, name string) error
	Stop(ctx context.Context, name string) error
	Delete(ctx context.Context, name string) error
	Status(ctx context.Context, name string) (*EnclaveStatus, error)
	Metrics(ctx context.Context, name string) (*RuntimeMetrics, error)
}

// NewManager creates a new enclave manager
func NewManager(baseDir string) *Manager {
	return &Manager{
		enclaves:  make(map[string]*Enclave),
		providers: make(map[string]Provider),
		baseDir:   baseDir,
	}
}

// RegisterProvider registers an enclave provider
func (m *Manager) RegisterProvider(name string, provider Provider) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.providers[name] = provider
}

// Load loads an enclave configuration from file
func (m *Manager) Load(path string) (*Enclave, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var enclave Enclave
	ext := filepath.Ext(path)
	if ext == ".yaml" || ext == ".yml" {
		if err := yaml.Unmarshal(data, &enclave); err != nil {
			return nil, fmt.Errorf("failed to parse YAML: %w", err)
		}
	} else {
		if err := json.Unmarshal(data, &enclave); err != nil {
			return nil, fmt.Errorf("failed to parse JSON: %w", err)
		}
	}

	return &enclave, nil
}

// Validate validates an enclave configuration
func (m *Manager) Validate(enclave *Enclave) error {
	if enclave.APIVersion != "omnigraph/enclave/v1" {
		return fmt.Errorf("unsupported apiVersion: %s", enclave.APIVersion)
	}
	if enclave.Kind != "WasmEnclave" {
		return fmt.Errorf("unsupported kind: %s", enclave.Kind)
	}
	if enclave.Metadata.Name == "" {
		return fmt.Errorf("metadata.name is required")
	}
	if enclave.Spec.Runtime.Engine == "" {
		return fmt.Errorf("spec.runtime.engine is required")
	}
	if enclave.Spec.TrustBoundary.Enrollment == "" {
		return fmt.Errorf("spec.trustBoundary.enrollment is required")
	}
	if enclave.Spec.CognitivePayload.SourceURI == "" {
		return fmt.Errorf("spec.cognitivePayload.sourceUri is required")
	}
	return ValidateContract(enclave)
}

// Deploy deploys an enclave
func (m *Manager) Deploy(ctx context.Context, enclave *Enclave) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Store enclave
	key := enclave.Metadata.Name
	m.enclaves[key] = enclave

	// Create status
	enclave.Status = &EnclaveStatus{
		Phase: "pending",
		Conditions: []Condition{
			{
				Type:               "Deployed",
				Status:             "False",
				LastTransitionTime: time.Now(),
				Reason:             "Deploying",
				Message:            "Enclave deployment initiated",
			},
		},
	}

	return nil
}

// Enroll performs ZTEE enrollment
func (m *Manager) Enroll(ctx context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	enclave, exists := m.enclaves[name]
	if !exists {
		return fmt.Errorf("enclave not found: %s", name)
	}

	if enclave.Status == nil {
		enclave.Status = &EnclaveStatus{}
	}

	enclave.Status.Phase = "enrolling"

	// Simulate enrollment
	time.Sleep(100 * time.Millisecond)

	enclave.Status.EnrollmentStatus = &EnrollmentStatus{
		Enrolled:          true,
		AttestedAt:        time.Now(),
		CertificateExpiry: time.Now().Add(24 * time.Hour),
	}

	enclave.Status.Phase = "running"
	enclave.Status.Conditions = []Condition{
		{
			Type:               "Enrolled",
			Status:             "True",
			LastTransitionTime: time.Now(),
			Reason:             "ZTEEComplete",
			Message:            "Zero-Trust Enclave Enrollment successful",
		},
	}

	return nil
}

// GetStatus returns the status of an enclave
func (m *Manager) GetStatus(name string) (*EnclaveStatus, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	enclave, exists := m.enclaves[name]
	if !exists {
		return nil, fmt.Errorf("enclave not found: %s", name)
	}

	return enclave.Status, nil
}

// List returns all managed enclaves
func (m *Manager) List() []*Enclave {
	m.mu.RLock()
	defer m.mu.RUnlock()

	enclaves := make([]*Enclave, 0, len(m.enclaves))
	for _, e := range m.enclaves {
		enclaves = append(enclaves, e)
	}
	return enclaves
}

// Delete removes an enclave
func (m *Manager) Delete(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.enclaves[name]; !exists {
		return fmt.Errorf("enclave not found: %s", name)
	}

	delete(m.enclaves, name)
	return nil
}

// UpdateMetrics updates runtime metrics for an enclave
func (m *Manager) UpdateMetrics(name string, metrics RuntimeMetrics) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	enclave, exists := m.enclaves[name]
	if !exists {
		return fmt.Errorf("enclave not found: %s", name)
	}

	if enclave.Status == nil {
		enclave.Status = &EnclaveStatus{}
	}

	enclave.Status.RuntimeMetrics = &metrics
	return nil
}
