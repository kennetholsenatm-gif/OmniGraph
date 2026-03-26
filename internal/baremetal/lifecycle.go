package baremetal

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// LifecycleManager manages the provisioning lifecycle for bare-metal targets
type LifecycleManager struct {
	mu       sync.RWMutex
	config   Config
	states   map[string]*TargetState
	eventBus EventBus
	handlers map[LifecyclePhase]LifecycleHandler
}

// LifecyclePhase represents a phase in the provisioning lifecycle
type LifecyclePhase string

const (
	PhaseDiscovery LifecyclePhase = "discovery"
	PhaseFirmware  LifecyclePhase = "firmware"
	PhaseRAID      LifecyclePhase = "raid"
	PhaseNetwork   LifecyclePhase = "network"
	PhaseBoot      LifecyclePhase = "boot"
	PhaseWait      LifecyclePhase = "wait"
	PhaseHandoff   LifecyclePhase = "handoff"
)

// TargetState represents the current state of a target
type TargetState struct {
	TargetID    string
	Phase       LifecyclePhase
	Status      PhaseStatus
	StartedAt   time.Time
	CompletedAt *time.Time
	Error       string
	Metadata    map[string]interface{}
}

// PhaseStatus represents the status of a lifecycle phase
type PhaseStatus string

const (
	PhaseStatusPending PhaseStatus = "pending"
	PhaseStatusRunning PhaseStatus = "running"
	PhaseStatusSuccess PhaseStatus = "success"
	PhaseStatusFailed  PhaseStatus = "failed"
	PhaseStatusSkipped PhaseStatus = "skipped"
)

// LifecycleHandler handles a specific lifecycle phase
type LifecycleHandler interface {
	// Execute executes the phase
	Execute(ctx context.Context, target *Target, state *TargetState) error

	// Validate validates prerequisites for the phase
	Validate(ctx context.Context, target *Target) error

	// Rollback rolls back the phase if needed
	Rollback(ctx context.Context, target *Target, state *TargetState) error
}

// NewLifecycleManager creates a new lifecycle manager
func NewLifecycleManager(config Config) *LifecycleManager {
	m := &LifecycleManager{
		config:   config,
		states:   make(map[string]*TargetState),
		handlers: make(map[LifecyclePhase]LifecycleHandler),
	}

	// Register default handlers
	m.RegisterHandler(PhaseDiscovery, &DiscoveryHandler{config: config})
	m.RegisterHandler(PhaseFirmware, &FirmwareHandler{config: config})
	m.RegisterHandler(PhaseRAID, &RAIDHandler{config: config})
	m.RegisterHandler(PhaseNetwork, &NetworkHandler{config: config})
	m.RegisterHandler(PhaseBoot, &BootHandler{config: config})
	m.RegisterHandler(PhaseWait, &WaitHandler{config: config})
	m.RegisterHandler(PhaseHandoff, &HandoffHandler{config: config})

	return m
}

// RegisterHandler registers a lifecycle handler
func (m *LifecycleManager) RegisterHandler(phase LifecyclePhase, handler LifecycleHandler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handlers[phase] = handler
}

// Execute executes the provisioning lifecycle for a target
func (m *LifecycleManager) Execute(ctx context.Context, target *Target) error {
	// Initialize target state
	state := &TargetState{
		TargetID:  target.ID,
		Phase:     PhaseDiscovery,
		Status:    PhaseStatusPending,
		StartedAt: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	m.mu.Lock()
	m.states[target.ID] = state
	m.mu.Unlock()

	// Define lifecycle phases in order
	phases := []LifecyclePhase{
		PhaseDiscovery,
		PhaseFirmware,
		PhaseRAID,
		PhaseNetwork,
		PhaseBoot,
		PhaseWait,
		PhaseHandoff,
	}

	// Execute each phase
	for _, phase := range phases {
		if err := m.executePhase(ctx, target, state, phase); err != nil {
			return fmt.Errorf("phase %s failed: %w", phase, err)
		}
	}

	// Mark as completed
	now := time.Now()
	state.CompletedAt = &now
	state.Status = PhaseStatusSuccess

	// Publish completion event
	if m.eventBus != nil {
		m.eventBus.Publish(Event{
			Type:      "lifecycle.completed",
			TargetID:  target.ID,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"duration": now.Sub(state.StartedAt),
			},
		})
	}

	return nil
}

// executePhase executes a single lifecycle phase
func (m *LifecycleManager) executePhase(ctx context.Context, target *Target, state *TargetState, phase LifecyclePhase) error {
	m.mu.RLock()
	handler, exists := m.handlers[phase]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("no handler registered for phase: %s", phase)
	}

	// Update state
	m.mu.Lock()
	state.Phase = phase
	state.Status = PhaseStatusRunning
	m.mu.Unlock()

	// Publish phase start event
	if m.eventBus != nil {
		m.eventBus.Publish(Event{
			Type:      "lifecycle.phase.start",
			TargetID:  target.ID,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"phase": string(phase),
			},
		})
	}

	// Validate prerequisites
	if err := handler.Validate(ctx, target); err != nil {
		m.mu.Lock()
		state.Status = PhaseStatusFailed
		state.Error = err.Error()
		m.mu.Unlock()

		// Publish failure event
		if m.eventBus != nil {
			m.eventBus.Publish(Event{
				Type:      "lifecycle.phase.failed",
				TargetID:  target.ID,
				Timestamp: time.Now(),
				Data: map[string]interface{}{
					"phase": string(phase),
					"error": err.Error(),
				},
			})
		}

		return fmt.Errorf("validation failed: %w", err)
	}

	// Execute phase
	if err := handler.Execute(ctx, target, state); err != nil {
		m.mu.Lock()
		state.Status = PhaseStatusFailed
		state.Error = err.Error()
		m.mu.Unlock()

		// Attempt rollback
		if rollbackErr := handler.Rollback(ctx, target, state); rollbackErr != nil {
			log.Printf("Rollback failed for phase %s: %v", phase, rollbackErr)
		}

		// Publish failure event
		if m.eventBus != nil {
			m.eventBus.Publish(Event{
				Type:      "lifecycle.phase.failed",
				TargetID:  target.ID,
				Timestamp: time.Now(),
				Data: map[string]interface{}{
					"phase": string(phase),
					"error": err.Error(),
				},
			})
		}

		return fmt.Errorf("execution failed: %w", err)
	}

	// Mark phase as successful
	m.mu.Lock()
	state.Status = PhaseStatusSuccess
	m.mu.Unlock()

	// Publish phase completion event
	if m.eventBus != nil {
		m.eventBus.Publish(Event{
			Type:      "lifecycle.phase.completed",
			TargetID:  target.ID,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"phase": string(phase),
			},
		})
	}

	return nil
}

// GetState returns the current state of a target
func (m *LifecycleManager) GetState(targetID string) (*TargetState, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	state, exists := m.states[targetID]
	if !exists {
		return nil, fmt.Errorf("no state found for target: %s", targetID)
	}

	return state, nil
}

// GetAllStates returns all target states
func (m *LifecycleManager) GetAllStates() map[string]*TargetState {
	m.mu.RLock()
	defer m.mu.RUnlock()

	states := make(map[string]*TargetState)
	for k, v := range m.states {
		states[k] = v
	}

	return states
}

// SetEventBus sets the event bus
func (m *LifecycleManager) SetEventBus(bus EventBus) {
	m.eventBus = bus
}

// DiscoveryHandler handles hardware discovery
type DiscoveryHandler struct {
	config Config
}

func (h *DiscoveryHandler) Validate(ctx context.Context, target *Target) error {
	if target.BMC.Address == "" {
		return fmt.Errorf("BMC address is required")
	}
	return nil
}

func (h *DiscoveryHandler) Execute(ctx context.Context, target *Target, state *TargetState) error {
	// Simulate hardware discovery
	time.Sleep(2 * time.Second)

	// Store discovered information
	state.Metadata["hardware_discovered"] = true
	state.Metadata["discovery_time"] = time.Now()

	return nil
}

func (h *DiscoveryHandler) Rollback(ctx context.Context, target *Target, state *TargetState) error {
	// Nothing to rollback for discovery
	return nil
}

// FirmwareHandler handles firmware updates
type FirmwareHandler struct {
	config Config
}

func (h *FirmwareHandler) Validate(ctx context.Context, target *Target) error {
	// Firmware update is optional
	return nil
}

func (h *FirmwareHandler) Execute(ctx context.Context, target *Target, state *TargetState) error {
	// Skip if firmware policy is none
	if target.FirmwarePolicy == "none" {
		state.Metadata["firmware_skipped"] = true
		return nil
	}

	// Simulate firmware update
	time.Sleep(5 * time.Second)

	state.Metadata["firmware_updated"] = true
	state.Metadata["firmware_time"] = time.Now()

	return nil
}

func (h *FirmwareHandler) Rollback(ctx context.Context, target *Target, state *TargetState) error {
	// Firmware rollback is complex and vendor-specific
	return nil
}

// RAIDHandler handles RAID configuration
type RAIDHandler struct {
	config Config
}

func (h *RAIDHandler) Validate(ctx context.Context, target *Target) error {
	// RAID configuration is optional
	if target.RAIDConfig == nil {
		return nil
	}

	if target.RAIDConfig.Level == "" {
		return fmt.Errorf("RAID level is required")
	}

	if len(target.RAIDConfig.Disks) == 0 {
		return fmt.Errorf("at least one disk is required for RAID")
	}

	return nil
}

func (h *RAIDHandler) Execute(ctx context.Context, target *Target, state *TargetState) error {
	// Skip if no RAID configuration
	if target.RAIDConfig == nil {
		state.Metadata["raid_skipped"] = true
		return nil
	}

	// Simulate RAID configuration
	time.Sleep(3 * time.Second)

	state.Metadata["raid_configured"] = true
	state.Metadata["raid_level"] = target.RAIDConfig.Level
	state.Metadata["raid_time"] = time.Now()

	return nil
}

func (h *RAIDHandler) Rollback(ctx context.Context, target *Target, state *TargetState) error {
	// RAID rollback is complex
	return nil
}

// NetworkHandler handles network boot configuration
type NetworkHandler struct {
	config Config
}

func (h *NetworkHandler) Validate(ctx context.Context, target *Target) error {
	if target.MACAddress == "" {
		return fmt.Errorf("MAC address is required")
	}

	if target.BootMode == "" {
		return fmt.Errorf("boot mode is required")
	}

	return nil
}

func (h *NetworkHandler) Execute(ctx context.Context, target *Target, state *TargetState) error {
	// Simulate network configuration
	time.Sleep(2 * time.Second)

	state.Metadata["network_configured"] = true
	state.Metadata["boot_mode"] = target.BootMode
	state.Metadata["mac_address"] = target.MACAddress
	state.Metadata["network_time"] = time.Now()

	return nil
}

func (h *NetworkHandler) Rollback(ctx context.Context, target *Target, state *TargetState) error {
	// Reset boot order
	return nil
}

// BootHandler handles boot orchestration
type BootHandler struct {
	config Config
}

func (h *BootHandler) Validate(ctx context.Context, target *Target) error {
	// No specific validation required
	return nil
}

func (h *BootHandler) Execute(ctx context.Context, target *Target, state *TargetState) error {
	// Simulate boot process
	time.Sleep(10 * time.Second)

	state.Metadata["boot_initiated"] = true
	state.Metadata["boot_time"] = time.Now()

	return nil
}

func (h *BootHandler) Rollback(ctx context.Context, target *Target, state *TargetState) error {
	// Power off the target
	return nil
}

// WaitHandler handles waiting for OS boot
type WaitHandler struct {
	config Config
}

func (h *WaitHandler) Validate(ctx context.Context, target *Target) error {
	// No specific validation required
	return nil
}

func (h *WaitHandler) Execute(ctx context.Context, target *Target, state *TargetState) error {
	// Simulate waiting for OS boot
	timeout := h.config.FirstBootTimeout
	if timeout == 0 {
		timeout = 20 * time.Minute
	}
	sleep := 15 * time.Second
	if timeout < sleep {
		sleep = timeout
	}
	time.Sleep(sleep)

	state.Metadata["os_booted"] = true
	state.Metadata["boot_completed"] = time.Now()

	return nil
}

func (h *WaitHandler) Rollback(ctx context.Context, target *Target, state *TargetState) error {
	// Nothing to rollback
	return nil
}

// HandoffHandler handles handoff to Ansible
type HandoffHandler struct {
	config Config
}

func (h *HandoffHandler) Validate(ctx context.Context, target *Target) error {
	// No specific validation required
	return nil
}

func (h *HandoffHandler) Execute(ctx context.Context, target *Target, state *TargetState) error {
	// Simulate handoff to Ansible
	time.Sleep(2 * time.Second)

	state.Metadata["handoff_completed"] = true
	state.Metadata["ansible_ready"] = true
	state.Metadata["handoff_time"] = time.Now()

	return nil
}

func (h *HandoffHandler) Rollback(ctx context.Context, target *Target, state *TargetState) error {
	// Nothing to rollback
	return nil
}
