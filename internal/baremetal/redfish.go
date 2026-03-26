package baremetal

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// RedfishProvider implements HardwareProvider for Redfish-compatible BMCs
// This includes Dell iDRAC, HPE iLO, Lenovo XClarity, and Supermicro
type RedfishProvider struct {
	config    Config
	client    *http.Client
	baseURL   string
	sessionID string
	token     string
}

// RedfishServiceRoot represents the Redfish service root
type RedfishServiceRoot struct {
	ODataContext   string   `json:"@odata.context"`
	ODataID        string   `json:"@odata.id"`
	ODataType      string   `json:"@odata.type"`
	ID             string   `json:"Id"`
	Name           string   `json:"Name"`
	RedfishVersion string   `json:"RedfishVersion"`
	UUID           string   `json:"UUID"`
	Systems        ODataRef `json:"Systems"`
	Chassis        ODataRef `json:"Chassis"`
	Managers       ODataRef `json:"Managers"`
	AccountService ODataRef `json:"AccountService"`
	SessionService ODataRef `json:"SessionService"`
}

// ODataRef represents an OData reference
type ODataRef struct {
	ODataID string `json:"@odata.id"`
}

// RedfishSystem represents a computer system
type RedfishSystem struct {
	ODataContext      string           `json:"@odata.context"`
	ODataID           string           `json:"@odata.id"`
	ODataType         string           `json:"@odata.type"`
	ID                string           `json:"Id"`
	Name              string           `json:"Name"`
	SystemType        string           `json:"SystemType"`
	Manufacturer      string           `json:"Manufacturer"`
	Model             string           `json:"Model"`
	SerialNumber      string           `json:"SerialNumber"`
	UUID              string           `json:"UUID"`
	PowerState        string           `json:"PowerState"`
	Status            Status           `json:"Status"`
	ProcessorSummary  ProcessorSummary `json:"ProcessorSummary"`
	MemorySummary     MemorySummary    `json:"MemorySummary"`
	Boot              Boot             `json:"Boot"`
	NetworkInterfaces ODataRef         `json:"NetworkInterfaces"`
}

// Status represents the status of a resource
type Status struct {
	State  string `json:"State"`
	Health string `json:"Health"`
}

// ProcessorSummary represents processor summary
type ProcessorSummary struct {
	Count  int    `json:"Count"`
	Model  string `json:"Model"`
	Status Status `json:"Status"`
}

// MemorySummary represents memory summary
type MemorySummary struct {
	TotalSystemMemoryGiB float64 `json:"TotalSystemMemoryGiB"`
	Status               Status  `json:"Status"`
}

// Boot represents boot configuration
type Boot struct {
	BootSourceOverrideEnabled    string   `json:"BootSourceOverrideEnabled"`
	BootSourceOverrideTarget     string   `json:"BootSourceOverrideTarget"`
	BootSourceOverrideTargetList []string `json:"BootSourceOverrideTarget@Redfish.AllowableValues"`
}

// RedfishManager represents a manager (BMC)
type RedfishManager struct {
	ODataContext       string   `json:"@odata.context"`
	ODataID            string   `json:"@odata.id"`
	ODataType          string   `json:"@odata.type"`
	ID                 string   `json:"Id"`
	Name               string   `json:"Name"`
	ManagerType        string   `json:"ManagerType"`
	FirmwareVersion    string   `json:"FirmwareVersion"`
	Status             Status   `json:"Status"`
	EthernetInterfaces ODataRef `json:"EthernetInterfaces"`
}

// RedfishNetworkInterface represents a network interface
type RedfishNetworkInterface struct {
	ODataContext string `json:"@odata.context"`
	ODataID      string `json:"@odata.id"`
	ODataType    string `json:"@odata.type"`
	ID           string `json:"Id"`
	Name         string `json:"Name"`
	MACAddress   string `json:"MACAddress"`
	Status       Status `json:"Status"`
}

// RedfishStorage represents storage configuration
type RedfishStorage struct {
	ODataContext string     `json:"@odata.context"`
	ODataID      string     `json:"@odata.id"`
	ODataType    string     `json:"@odata.type"`
	ID           string     `json:"Id"`
	Name         string     `json:"Name"`
	Drives       []ODataRef `json:"Drives"`
	Volumes      ODataRef   `json:"Volumes"`
}

// RedfishDrive represents a physical drive
type RedfishDrive struct {
	ODataContext  string `json:"@odata.context"`
	ODataID       string `json:"@odata.id"`
	ODataType     string `json:"@odata.type"`
	ID            string `json:"Id"`
	Name          string `json:"Name"`
	CapacityBytes int64  `json:"CapacityBytes"`
	MediaType     string `json:"MediaType"`
	Status        Status `json:"Status"`
}

// RedfishVolume represents a logical volume
type RedfishVolume struct {
	ODataContext  string `json:"@odata.context"`
	ODataID       string `json:"@odata.id"`
	ODataType     string `json:"@odata.type"`
	ID            string `json:"Id"`
	Name          string `json:"Name"`
	VolumeType    string `json:"VolumeType"`
	CapacityBytes int64  `json:"CapacityBytes"`
	Status        Status `json:"Status"`
}

// NewRedfishProvider creates a new Redfish provider
func NewRedfishProvider(config Config) *RedfishProvider {
	// Create HTTP client with TLS configuration
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // BMCs often use self-signed certs
		},
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: false,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	return &RedfishProvider{
		config: config,
		client: client,
	}
}

// connect establishes a connection to the Redfish API
func (p *RedfishProvider) connect(ctx context.Context, target *Target) error {
	if target.BMC.Type != "redfish" {
		return fmt.Errorf("unsupported BMC type: %s", target.BMC.Type)
	}

	// Build base URL
	port := target.BMC.Port
	if port == 0 {
		port = 443
	}
	p.baseURL = fmt.Sprintf("https://%s:%d/redfish/v1", target.BMC.Address, port)

	// Get service root to verify connection
	serviceRoot, err := p.getServiceRoot(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to Redfish API: %w", err)
	}

	// Verify Redfish version
	if !strings.HasPrefix(serviceRoot.RedfishVersion, "1.") {
		return fmt.Errorf("unsupported Redfish version: %s", serviceRoot.RedfishVersion)
	}

	return nil
}

// getServiceRoot retrieves the Redfish service root
func (p *RedfishProvider) getServiceRoot(ctx context.Context) (*RedfishServiceRoot, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", p.baseURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get service root: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var serviceRoot RedfishServiceRoot
	if err := json.Unmarshal(body, &serviceRoot); err != nil {
		return nil, err
	}

	return &serviceRoot, nil
}

// PowerOn powers on the target
func (p *RedfishProvider) PowerOn(ctx context.Context, target *Target) error {
	if err := p.connect(ctx, target); err != nil {
		return err
	}

	system, err := p.getSystem(ctx)
	if err != nil {
		return err
	}

	if system.PowerState == "On" {
		return nil // Already on
	}

	// Send power on request
	powerAction := map[string]string{
		"ResetType": "On",
	}

	return p.sendSystemAction(ctx, "ComputerSystem.Reset", powerAction)
}

// PowerOff powers off the target
func (p *RedfishProvider) PowerOff(ctx context.Context, target *Target) error {
	if err := p.connect(ctx, target); err != nil {
		return err
	}

	system, err := p.getSystem(ctx)
	if err != nil {
		return err
	}

	if system.PowerState == "Off" {
		return nil // Already off
	}

	// Send power off request
	powerAction := map[string]string{
		"ResetType": "GracefulShutdown",
	}

	return p.sendSystemAction(ctx, "ComputerSystem.Reset", powerAction)
}

// PowerStatus returns the power status of the target
func (p *RedfishProvider) PowerStatus(ctx context.Context, target *Target) (string, error) {
	if err := p.connect(ctx, target); err != nil {
		return "", err
	}

	system, err := p.getSystem(ctx)
	if err != nil {
		return "", err
	}

	return system.PowerState, nil
}

// SetBootOrder sets the boot order for the target
func (p *RedfishProvider) SetBootOrder(ctx context.Context, target *Target, order []string) error {
	if err := p.connect(ctx, target); err != nil {
		return err
	}

	// Map boot order to Redfish boot source
	var bootSource string
	if len(order) > 0 {
		switch order[0] {
		case "http":
			bootSource = "Http"
		case "pxe":
			bootSource = "Pxe"
		case "network":
			bootSource = "Network"
		case "disk":
			bootSource = "Hdd"
		default:
			bootSource = "None"
		}
	}

	// Update boot configuration
	bootConfig := map[string]interface{}{
		"Boot": map[string]string{
			"BootSourceOverrideEnabled": "Once",
			"BootSourceOverrideTarget":  bootSource,
		},
	}

	return p.updateSystem(ctx, bootConfig)
}

// GetBootOrder returns the current boot order
func (p *RedfishProvider) GetBootOrder(ctx context.Context, target *Target) ([]string, error) {
	if err := p.connect(ctx, target); err != nil {
		return nil, err
	}

	system, err := p.getSystem(ctx)
	if err != nil {
		return nil, err
	}

	var order []string
	switch system.Boot.BootSourceOverrideTarget {
	case "Http":
		order = []string{"http", "disk"}
	case "Pxe":
		order = []string{"pxe", "disk"}
	case "Network":
		order = []string{"network", "disk"}
	case "Hdd":
		order = []string{"disk"}
	default:
		order = []string{"disk"}
	}

	return order, nil
}

// GetFirmwareVersion returns the firmware version
func (p *RedfishProvider) GetFirmwareVersion(ctx context.Context, target *Target) (string, error) {
	if err := p.connect(ctx, target); err != nil {
		return "", err
	}

	serviceRoot, err := p.getServiceRoot(ctx)
	if err != nil {
		return "", err
	}

	// Get manager (BMC)
	managerURL := p.baseURL + serviceRoot.Managers.ODataID
	manager, err := p.getManager(ctx, managerURL)
	if err != nil {
		return "", err
	}

	return manager.FirmwareVersion, nil
}

// UpdateFirmware updates the firmware
func (p *RedfishProvider) UpdateFirmware(ctx context.Context, target *Target, firmwareURL string) error {
	// Firmware update is complex and vendor-specific
	// This is a simplified implementation
	return fmt.Errorf("firmware update not yet implemented for Redfish")
}

// GetRAIDConfiguration returns the RAID configuration
func (p *RedfishProvider) GetRAIDConfiguration(ctx context.Context, target *Target) (*RAIDConfig, error) {
	if err := p.connect(ctx, target); err != nil {
		return nil, err
	}

	// Get storage information
	serviceRoot, err := p.getServiceRoot(ctx)
	if err != nil {
		return nil, err
	}

	systemURL := p.baseURL + serviceRoot.Systems.ODataID
	system, err := p.getSystemFromURL(ctx, systemURL)
	if err != nil {
		return nil, err
	}

	// Get storage controllers
	storageURL := p.baseURL + system.ODataID + "/Storage"
	storage, err := p.getStorage(ctx, storageURL)
	if err != nil {
		return nil, err
	}

	// Get volumes
	if storage.Volumes.ODataID != "" {
		volumesURL := p.baseURL + storage.Volumes.ODataID
		volumes, err := p.getVolumes(ctx, volumesURL)
		if err != nil {
			return nil, err
		}

		// Return first volume as RAID config
		if len(volumes) > 0 {
			return &RAIDConfig{
				Level: volumes[0].VolumeType,
				Disks: []string{}, // Would need to parse drives
			}, nil
		}
	}

	return nil, fmt.Errorf("no RAID configuration found")
}

// SetRAIDConfiguration sets the RAID configuration
func (p *RedfishProvider) SetRAIDConfiguration(ctx context.Context, target *Target, config *RAIDConfig) error {
	// RAID configuration is complex and vendor-specific
	// This is a simplified implementation
	return fmt.Errorf("RAID configuration not yet implemented for Redfish")
}

// GetMACAddress returns the MAC address
func (p *RedfishProvider) GetMACAddress(ctx context.Context, target *Target) (string, error) {
	if err := p.connect(ctx, target); err != nil {
		return "", err
	}

	serviceRoot, err := p.getServiceRoot(ctx)
	if err != nil {
		return "", err
	}

	systemURL := p.baseURL + serviceRoot.Systems.ODataID
	system, err := p.getSystemFromURL(ctx, systemURL)
	if err != nil {
		return "", err
	}

	// Get network interfaces
	if system.NetworkInterfaces.ODataID != "" {
		nicURL := p.baseURL + system.NetworkInterfaces.ODataID
		nic, err := p.getNetworkInterface(ctx, nicURL)
		if err != nil {
			return "", err
		}
		return nic.MACAddress, nil
	}

	return "", fmt.Errorf("no MAC address found")
}

// GetBMCInfo returns BMC information
func (p *RedfishProvider) GetBMCInfo(ctx context.Context, target *Target) (*BMCInfo, error) {
	if err := p.connect(ctx, target); err != nil {
		return nil, err
	}

	serviceRoot, err := p.getServiceRoot(ctx)
	if err != nil {
		return nil, err
	}

	// Get manager (BMC)
	managerURL := p.baseURL + serviceRoot.Managers.ODataID
	manager, err := p.getManager(ctx, managerURL)
	if err != nil {
		return nil, err
	}

	// Get power state
	systemURL := p.baseURL + serviceRoot.Systems.ODataID
	system, err := p.getSystemFromURL(ctx, systemURL)
	if err != nil {
		return nil, err
	}

	return &BMCInfo{
		Type:       "redfish",
		Address:    target.BMC.Address,
		Firmware:   manager.FirmwareVersion,
		Health:     manager.Status.Health,
		PowerState: system.PowerState,
	}, nil
}

// Helper methods

func (p *RedfishProvider) getSystem(ctx context.Context) (*RedfishSystem, error) {
	serviceRoot, err := p.getServiceRoot(ctx)
	if err != nil {
		return nil, err
	}

	systemURL := p.baseURL + serviceRoot.Systems.ODataID
	return p.getSystemFromURL(ctx, systemURL)
}

func (p *RedfishProvider) getSystemFromURL(ctx context.Context, url string) (*RedfishSystem, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var system RedfishSystem
	if err := json.Unmarshal(body, &system); err != nil {
		return nil, err
	}

	return &system, nil
}

func (p *RedfishProvider) getManager(ctx context.Context, url string) (*RedfishManager, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var manager RedfishManager
	if err := json.Unmarshal(body, &manager); err != nil {
		return nil, err
	}

	return &manager, nil
}

func (p *RedfishProvider) getNetworkInterface(ctx context.Context, url string) (*RedfishNetworkInterface, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var nic RedfishNetworkInterface
	if err := json.Unmarshal(body, &nic); err != nil {
		return nil, err
	}

	return &nic, nil
}

func (p *RedfishProvider) getStorage(ctx context.Context, url string) (*RedfishStorage, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var storage RedfishStorage
	if err := json.Unmarshal(body, &storage); err != nil {
		return nil, err
	}

	return &storage, nil
}

func (p *RedfishProvider) getVolumes(ctx context.Context, url string) ([]RedfishVolume, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var volumeList struct {
		Members []RedfishVolume `json:"Members"`
	}
	if err := json.Unmarshal(body, &volumeList); err != nil {
		return nil, err
	}

	return volumeList.Members, nil
}

func (p *RedfishProvider) sendSystemAction(ctx context.Context, action string, payload interface{}) error {
	systemURL := p.baseURL + "/Systems/1"

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", systemURL+"/Actions/"+action, strings.NewReader(string(jsonPayload)))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("action failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (p *RedfishProvider) updateSystem(ctx context.Context, payload interface{}) error {
	systemURL := p.baseURL + "/Systems/1"

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "PATCH", systemURL, strings.NewReader(string(jsonPayload)))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("update failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
