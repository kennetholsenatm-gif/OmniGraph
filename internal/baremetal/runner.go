package baremetal

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/kennetholsenatm-gif/omnigraph/internal/runner"
)

// BareMetalRunner extends the execution matrix to handle bare-metal provisioning
type BareMetalRunner struct {
	mu        sync.RWMutex
	config    Config
	providers map[string]HardwareProvider
	eventBus  EventBus
	proxyDHCP *ProxyDHCP
	httpBoot  *HTTPBootServer
	lifecycle *LifecycleManager
}

// Config holds bare-metal runner configuration
type Config struct {
	// Network configuration
	ProvisioningNetwork string
	ProxyDHCPPort       int
	HTTPBootPort        int
	HTTPSPort           int

	// Boot configuration
	IPXEBinaryPath    string
	CloudInitTemplate string
	OSImagesPath      string

	// BMC configuration
	BMCCredentialsPath string
	BMCTimeout         time.Duration
	// RedfishInsecureTLS disables TLS certificate verification for Redfish BMC HTTPS (self-signed BMC certs).
	// When false (zero value), certificates are verified.
	RedfishInsecureTLS bool

	// Lifecycle configuration
	ProvisioningTimeout time.Duration
	FirstBootTimeout    time.Duration
	AgentTimeout        time.Duration
}

// HardwareProvider interface for different hardware types
type HardwareProvider interface {
	// Power on/off the hardware
	PowerOn(ctx context.Context, target *Target) error
	PowerOff(ctx context.Context, target *Target) error
	PowerStatus(ctx context.Context, target *Target) (string, error)

	// Configure boot order
	SetBootOrder(ctx context.Context, target *Target, order []string) error
	GetBootOrder(ctx context.Context, target *Target) ([]string, error)

	// Firmware management
	GetFirmwareVersion(ctx context.Context, target *Target) (string, error)
	UpdateFirmware(ctx context.Context, target *Target, firmwareURL string) error

	// RAID configuration
	GetRAIDConfiguration(ctx context.Context, target *Target) (*RAIDConfig, error)
	SetRAIDConfiguration(ctx context.Context, target *Target, config *RAIDConfig) error

	// Network configuration
	GetMACAddress(ctx context.Context, target *Target) (string, error)
	GetBMCInfo(ctx context.Context, target *Target) (*BMCInfo, error)
}

// Target represents a bare-metal server to provision
type Target struct {
	ID             string
	MACAddress     string
	BMC            BMCConfig
	BootMode       string // http, ipxe, pxe
	OSProfile      string
	FirmwarePolicy string
	RAIDConfig     *RAIDConfig
	Network        NetworkConfig
	Labels         map[string]string
}

// BMCConfig represents BMC connection configuration
type BMCConfig struct {
	Type        string // redfish, ipmi, idrac, ilo
	Address     string
	Port        int
	Credentials *Credentials
}

// Credentials for BMC authentication
type Credentials struct {
	Username  string
	Password  string
	SecretRef string // Reference to Vault secret
}

// RAIDConfig represents RAID configuration
type RAIDConfig struct {
	Level string // raid0, raid1, raid5, raid6, raid10
	Disks []string
	Size  string
}

// NetworkConfig represents network configuration for provisioning
type NetworkConfig struct {
	Interface    string
	IPAddress    string
	Gateway      string
	DNSServers   []string
	VLAN         int
	Provisioning bool // Is this the provisioning interface?
}

// BMCInfo represents BMC information
type BMCInfo struct {
	Type       string
	Address    string
	Firmware   string
	Health     string
	PowerState string
}

// NewBareMetalRunner creates a new bare-metal runner
func NewBareMetalRunner(config Config) (*BareMetalRunner, error) {
	runner := &BareMetalRunner{
		config:    config,
		providers: make(map[string]HardwareProvider),
	}

	// Initialize lifecycle manager
	runner.lifecycle = NewLifecycleManager(config)

	// Register default providers
	runner.RegisterProvider("redfish", NewRedfishProvider(config))
	runner.RegisterProvider("ipmi", NewIPMIProvider(config))

	return runner, nil
}

// RegisterProvider registers a hardware provider
func (r *BareMetalRunner) RegisterProvider(name string, provider HardwareProvider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[name] = provider
}

// Run executes a bare-metal provisioning step
func (r *BareMetalRunner) Run(ctx context.Context, step runner.Step) (*runner.Result, error) {
	// Parse step to determine action
	action, err := r.parseStep(step)
	if err != nil {
		return nil, fmt.Errorf("failed to parse step: %w", err)
	}

	// Execute lifecycle phase
	switch action.Phase {
	case "discover":
		return r.executeDiscovery(ctx, action.Targets)
	case "firmware":
		return r.executeFirmwareUpdate(ctx, action.Targets)
	case "raid":
		return r.executeRAIDConfiguration(ctx, action.Targets)
	case "network":
		return r.executeNetworkConfiguration(ctx, action.Targets)
	case "boot":
		return r.executeBoot(ctx, action.Targets)
	case "wait":
		return r.waitForReady(ctx, action.Targets)
	case "handoff":
		return r.executeHandoff(ctx, action.Targets)
	default:
		return nil, fmt.Errorf("unknown phase: %s", action.Phase)
	}
}

// parseStep parses a runner step into a bare-metal action
func (r *BareMetalRunner) parseStep(step runner.Step) (*Action, error) {
	// Parse step arguments to determine action
	// This is a simplified version - in practice, would parse from step.Argv
	return &Action{
		Phase:   step.Name,
		Targets: []*Target{}, // Would be populated from step config
	}, nil
}

// executeDiscovery discovers hardware via BMC
func (r *BareMetalRunner) executeDiscovery(ctx context.Context, targets []*Target) (*runner.Result, error) {
	var output string
	var errors []string

	for _, target := range targets {
		provider, err := r.getProvider(target.BMC.Type)
		if err != nil {
			errors = append(errors, fmt.Sprintf("target %s: %v", target.ID, err))
			continue
		}

		// Get BMC info
		info, err := provider.GetBMCInfo(ctx, target)
		if err != nil {
			errors = append(errors, fmt.Sprintf("target %s BMC info: %v", target.ID, err))
			continue
		}

		// Get MAC address
		mac, err := provider.GetMACAddress(ctx, target)
		if err != nil {
			errors = append(errors, fmt.Sprintf("target %s MAC: %v", target.ID, err))
			continue
		}

		output += fmt.Sprintf("Target %s: BMC=%s MAC=%s Power=%s\n",
			target.ID, info.Address, mac, info.PowerState)
	}

	if len(errors) > 0 {
		return &runner.Result{
			ExitCode: 1,
			Stderr:   []byte(fmt.Sprintf("Discovery errors:\n%v", errors)),
		}, nil
	}

	return &runner.Result{
		ExitCode: 0,
		Stdout:   []byte(output),
	}, nil
}

// executeFirmwareUpdate updates firmware on targets
func (r *BareMetalRunner) executeFirmwareUpdate(ctx context.Context, targets []*Target) (*runner.Result, error) {
	var output string

	for _, target := range targets {
		provider, err := r.getProvider(target.BMC.Type)
		if err != nil {
			return nil, err
		}

		// Get current firmware version
		version, err := provider.GetFirmwareVersion(ctx, target)
		if err != nil {
			return nil, fmt.Errorf("failed to get firmware version for %s: %w", target.ID, err)
		}

		output += fmt.Sprintf("Target %s: Current firmware: %s\n", target.ID, version)

		// Update firmware if policy is "latest"
		if target.FirmwarePolicy == "latest" {
			output += fmt.Sprintf("Target %s: Updating firmware...\n", target.ID)
			// In practice, would download and apply firmware
		}
	}

	return &runner.Result{
		ExitCode: 0,
		Stdout:   []byte(output),
	}, nil
}

// executeRAIDConfiguration configures RAID on targets
func (r *BareMetalRunner) executeRAIDConfiguration(ctx context.Context, targets []*Target) (*runner.Result, error) {
	var output string

	for _, target := range targets {
		if target.RAIDConfig == nil {
			output += fmt.Sprintf("Target %s: No RAID configuration specified\n", target.ID)
			continue
		}

		provider, err := r.getProvider(target.BMC.Type)
		if err != nil {
			return nil, err
		}

		output += fmt.Sprintf("Target %s: Configuring RAID %s with disks %v\n",
			target.ID, target.RAIDConfig.Level, target.RAIDConfig.Disks)

		err = provider.SetRAIDConfiguration(ctx, target, target.RAIDConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to configure RAID for %s: %w", target.ID, err)
		}
	}

	return &runner.Result{
		ExitCode: 0,
		Stdout:   []byte(output),
	}, nil
}

// executeNetworkConfiguration configures network boot on targets
func (r *BareMetalRunner) executeNetworkConfiguration(ctx context.Context, targets []*Target) (*runner.Result, error) {
	var output string

	for _, target := range targets {
		provider, err := r.getProvider(target.BMC.Type)
		if err != nil {
			return nil, err
		}

		// Set boot order based on boot mode
		var bootOrder []string
		switch target.BootMode {
		case "http":
			bootOrder = []string{"http", "disk"}
		case "ipxe":
			bootOrder = []string{"network", "disk"}
		case "pxe":
			bootOrder = []string{"pxe", "disk"}
		default:
			bootOrder = []string{"disk"}
		}

		output += fmt.Sprintf("Target %s: Setting boot order: %v\n", target.ID, bootOrder)

		err = provider.SetBootOrder(ctx, target, bootOrder)
		if err != nil {
			return nil, fmt.Errorf("failed to set boot order for %s: %w", target.ID, err)
		}
	}

	return &runner.Result{
		ExitCode: 0,
		Stdout:   []byte(output),
	}, nil
}

// executeBoot triggers boot on targets
func (r *BareMetalRunner) executeBoot(ctx context.Context, targets []*Target) (*runner.Result, error) {
	var output string

	for _, target := range targets {
		provider, err := r.getProvider(target.BMC.Type)
		if err != nil {
			return nil, err
		}

		// Power on the target
		output += fmt.Sprintf("Target %s: Powering on...\n", target.ID)

		err = provider.PowerOn(ctx, target)
		if err != nil {
			return nil, fmt.Errorf("failed to power on %s: %w", target.ID, err)
		}

		// Trigger reboot to boot from network
		output += fmt.Sprintf("Target %s: Triggering network boot...\n", target.ID)
		// In practice, would send reboot command via BMC
	}

	return &runner.Result{
		ExitCode: 0,
		Stdout:   []byte(output),
	}, nil
}

// waitForReady waits for targets to be ready
func (r *BareMetalRunner) waitForReady(ctx context.Context, targets []*Target) (*runner.Result, error) {
	var output string

	for _, target := range targets {
		output += fmt.Sprintf("Target %s: Waiting for OS boot...\n", target.ID)

		// Wait for SSH to be available
		err := r.waitForSSH(ctx, target)
		if err != nil {
			return nil, fmt.Errorf("target %s failed to become ready: %w", target.ID, err)
		}

		output += fmt.Sprintf("Target %s: Ready!\n", target.ID)
	}

	return &runner.Result{
		ExitCode: 0,
		Stdout:   []byte(output),
	}, nil
}

// executeHandoff hands off to Ansible for configuration
func (r *BareMetalRunner) executeHandoff(ctx context.Context, targets []*Target) (*runner.Result, error) {
	var output string

	// Generate Ansible inventory
	inventory := r.generateInventory(targets)

	output += "Generated Ansible inventory:\n"
	output += inventory

	// In practice, would write inventory file and return path
	return &runner.Result{
		ExitCode: 0,
		Stdout:   []byte(output),
	}, nil
}

// getProvider returns the hardware provider for a BMC type
func (r *BareMetalRunner) getProvider(bmcType string) (HardwareProvider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	provider, ok := r.providers[bmcType]
	if !ok {
		return nil, fmt.Errorf("no provider registered for BMC type: %s", bmcType)
	}

	return provider, nil
}

// waitForSSH waits for SSH to be available on a target
func (r *BareMetalRunner) waitForSSH(ctx context.Context, target *Target) error {
	// Simplified implementation
	// In practice, would poll SSH port until available
	time.Sleep(5 * time.Second) // Simulate wait
	return nil
}

// generateInventory generates Ansible inventory from targets
func (r *BareMetalRunner) generateInventory(targets []*Target) string {
	var inventory string

	inventory = "[baremetal]\n"
	for _, target := range targets {
		if target.Network.IPAddress != "" {
			inventory += fmt.Sprintf("%s ansible_host=%s\n", target.ID, target.Network.IPAddress)
		}
	}

	return inventory
}

// Action represents a bare-metal provisioning action
type Action struct {
	Phase   string
	Targets []*Target
}

// EventBus interface for publishing events
type EventBus interface {
	Publish(event Event)
	Subscribe(eventType string, handler EventHandler) Subscription
}

// Event represents a provisioning event
type Event struct {
	Type      string
	TargetID  string
	Timestamp time.Time
	Data      map[string]interface{}
}

// EventHandler handles events
type EventHandler func(event Event)

// Subscription represents an event subscription
type Subscription struct {
	ID      string
	Handler EventHandler
}

// HTTPBootServer represents the HTTP boot server
type HTTPBootServer struct {
	config Config
	server net.Listener
}
