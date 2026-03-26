package incus

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lxc/incus/client"
	"github.com/lxc/incus/shared/api"
	"github.com/kennetholsenatm-gif/omnigraph/internal/resources"
)

// Provider implements the Incus provider for OmniGraph
type Provider struct {
	client incus.InstanceServer
	project string
}

// NewProvider creates a new Incus provider
func NewProvider(config Config) (*Provider, error) {
	// Connect to Incus
	args := &incus.ConnectionArgs{
		TLSClientCert:      config.TLSClientCert,
		TLSClientKey:       config.TLSClientKey,
		TLSServerCert:      config.TLSServerCert,
		InsecureSkipVerify: config.InsecureSkipVerify,
		Proxy:              config.Proxy,
		UserAgent:          "omnigraph/1.0",
	}
	
	var client incus.InstanceServer
	var err error
	
	if config.Remote != "" {
		// Connect to remote Incus server
		client, err = incus.ConnectIncus(config.Remote, args)
	} else {
		// Connect to local Incus socket
		client, err = incus.ConnectIncusUnix(config.SocketPath, args)
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Incus: %w", err)
	}
	
	// Set project if specified
	if config.Project != "" {
		client = client.UseProject(config.Project)
	}
	
	return &Provider{
		client:  client,
		project: config.Project,
	}, nil
}

// Config contains Incus provider configuration
type Config struct {
	Remote            string
	SocketPath        string
	TLSClientCert     string
	TLSClientKey      string
	TLSServerCert     string
	InsecureSkipVerify bool
	Proxy             string
	Project           string
}

// GetActualState returns the actual state of a resource
func (p *Provider) GetActualState(ctx context.Context, resource resources.Resource) (resources.Resource, error) {
	switch resource.Kind {
	case "ComputeInstance":
		return p.getInstance(ctx, resource)
	case "Network":
		return p.getNetwork(ctx, resource)
	case "StoragePool":
		return p.getStoragePool(ctx, resource)
	case "Profile":
		return p.getProfile(ctx, resource)
	default:
		return resource, fmt.Errorf("unsupported resource kind: %s", resource.Kind)
	}
}

// Apply applies the desired state
func (p *Provider) Apply(ctx context.Context, desired resources.Resource) error {
	switch desired.Kind {
	case "ComputeInstance":
		return p.applyInstance(ctx, desired)
	case "Network":
		return p.applyNetwork(ctx, desired)
	case "StoragePool":
		return p.applyStoragePool(ctx, desired)
	case "Profile":
		return p.applyProfile(ctx, desired)
	default:
		return fmt.Errorf("unsupported resource kind: %s", desired.Kind)
	}
}

// Delete deletes a resource
func (p *Provider) Delete(ctx context.Context, resource resources.Resource) error {
	switch resource.Kind {
	case "ComputeInstance":
		return p.deleteInstance(ctx, resource)
	case "Network":
		return p.deleteNetwork(ctx, resource)
	case "StoragePool":
		return p.deleteStoragePool(ctx, resource)
	case "Profile":
		return p.deleteProfile(ctx, resource)
	default:
		return fmt.Errorf("unsupported resource kind: %s", resource.Kind)
	}
}

// Exists checks if a resource exists
func (p *Provider) Exists(ctx context.Context, resource resources.Resource) (bool, error) {
	switch resource.Kind {
	case "ComputeInstance":
		return p.instanceExists(ctx, resource)
	case "Network":
		return p.networkExists(ctx, resource)
	case "StoragePool":
		return p.storagePoolExists(ctx, resource)
	case "Profile":
		return p.profileExists(ctx, resource)
	default:
		return false, fmt.Errorf("unsupported resource kind: %s", resource.Kind)
	}
}

// Watch watches for changes
func (p *Provider) Watch(ctx context.Context, resource resources.Resource) (<-chan resources.ResourceEvent, error) {
	// For now, return a channel that will be populated by polling
	// In production, use Incus event listener
	ch := make(chan resources.ResourceEvent, 10)
	
	go func() {
		defer close(ch)
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		
		var lastState resources.Resource
		
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				current, err := p.GetActualState(ctx, resource)
				if err != nil {
					ch <- resources.ResourceEvent{
						Type:      resources.EventTypeError,
						Resource:  resource,
						Timestamp: time.Now(),
					}
					continue
				}
				
				if lastState.Spec == nil {
					// First poll
					ch <- resources.ResourceEvent{
						Type:      resources.EventTypeCreated,
						Resource:  current,
						Timestamp: time.Now(),
					}
				} else if !specsEqual(lastState.Spec, current.Spec) {
					// State changed
					ch <- resources.ResourceEvent{
						Type:      resources.EventTypeUpdated,
						Resource:  current,
						Timestamp: time.Now(),
					}
				}
				
				lastState = current
			}
		}
	}()
	
	return ch, nil
}

// Instance operations

func (p *Provider) getInstance(ctx context.Context, resource resources.Resource) (resources.Resource, error) {
	instanceName := resource.Metadata.Name
	
	instance, _, err := p.client.GetInstance(instanceName)
	if err != nil {
		return resource, err
	}
	
	// Convert Incus instance to OmniGraph resource
	state := map[string]interface{}{
		"provider": "incus",
		"type":     instance.Type,
		"state":    instance.Status,
		"config":   instance.Config,
		"profiles": instance.Profiles,
		"devices":  instance.Devices,
	}
	
	stateJSON, _ := json.Marshal(state)
	resource.Spec = stateJSON
	
	// Set status
	resource.Status = &resources.ResourceStatus{
		State: instance.Status,
		Provider: &resources.ProviderStatus{
			ID:        instance.Name,
			CreatedAt: instance.CreatedAt,
			UpdatedAt: instance.LastUsedAt,
		},
	}
	
	return resource, nil
}

func (p *Provider) applyInstance(ctx context.Context, desired resources.Resource) error {
	var spec resources.ComputeInstanceSpec
	if err := json.Unmarshal(desired.Spec, &spec); err != nil {
		return fmt.Errorf("failed to parse spec: %w", err)
	}
	
	instanceName := desired.Metadata.Name
	
	// Check if instance exists
	exists, _ := p.instanceExists(ctx, desired)
	
	if exists {
		// Update instance
		return p.updateInstance(ctx, instanceName, spec)
	}
	
	// Create instance
	return p.createInstance(ctx, instanceName, spec)
}

func (p *Provider) createInstance(ctx context.Context, name string, spec resources.ComputeInstanceSpec) error {
	// Build instance creation request
	req := api.InstancesPost{
		Name: name,
		Type: api.InstanceType(spec.Type),
		Source: api.InstanceSource{
			Type:     "image",
			Alias:    spec.Source.Alias,
			Server:   spec.Source.Server,
			Protocol: spec.Source.Protocol,
		},
		InstancePut: api.InstancePut{
			Config:   spec.Config,
			Profiles: spec.Profiles,
			Devices:  convertDevices(spec.Devices),
			Ephemeral: spec.Ephemeral,
		},
	}
	
	// Set default state
	if spec.State == "" {
		spec.State = "running"
	}
	
	// Create instance
	op, err := p.client.CreateInstance(req)
	if err != nil {
		return fmt.Errorf("failed to create instance: %w", err)
	}
	
	// Wait for operation to complete
	err = op.Wait()
	if err != nil {
		return fmt.Errorf("failed to wait for instance creation: %w", err)
	}
	
	// Start instance if requested
	if spec.State == "running" {
		return p.startInstance(ctx, name)
	}
	
	return nil
}

func (p *Provider) updateInstance(ctx context.Context, name string, spec resources.ComputeInstanceSpec) error {
	// Get current instance
	instance, etag, err := p.client.GetInstance(name)
	if err != nil {
		return err
	}
	
	// Update config
	instancePut := api.InstancePut{
		Config:   spec.Config,
		Profiles: spec.Profiles,
		Devices:  convertDevices(spec.Devices),
		Ephemeral: spec.Ephemeral,
	}
	
	// Update instance
	op, err := p.client.UpdateInstance(name, instancePut, etag)
	if err != nil {
		return fmt.Errorf("failed to update instance: %w", err)
	}
	
	err = op.Wait()
	if err != nil {
		return fmt.Errorf("failed to wait for instance update: %w", err)
	}
	
	// Handle state changes
	if spec.State == "running" && instance.Status != "Running" {
		return p.startInstance(ctx, name)
	} else if spec.State == "stopped" && instance.Status == "Running" {
		return p.stopInstance(ctx, name)
	}
	
	return nil
}

func (p *Provider) deleteInstance(ctx context.Context, resource resources.Resource) error {
	instanceName := resource.Metadata.Name
	
	// Stop instance first
	_ = p.stopInstance(ctx, instanceName)
	
	// Delete instance
	op, err := p.client.DeleteInstance(instanceName)
	if err != nil {
		return fmt.Errorf("failed to delete instance: %w", err)
	}
	
	return op.Wait()
}

func (p *Provider) instanceExists(ctx context.Context, resource resources.Resource) (bool, error) {
	instanceName := resource.Metadata.Name
	
	_, _, err := p.client.GetInstance(instanceName)
	if err != nil {
		return false, nil
	}
	
	return true, nil
}

func (p *Provider) startInstance(ctx context.Context, name string) error {
	req := api.InstanceStatePut{
		Action:  "start",
		Timeout: 30,
	}
	
	op, err := p.client.UpdateInstanceState(name, req, "")
	if err != nil {
		return fmt.Errorf("failed to start instance: %w", err)
	}
	
	return op.Wait()
}

func (p *Provider) stopInstance(ctx context.Context, name string) error {
	req := api.InstanceStatePut{
		Action:  "stop",
		Timeout: 30,
	}
	
	op, err := p.client.UpdateInstanceState(name, req, "")
	if err != nil {
		return fmt.Errorf("failed to stop instance: %w", err)
	}
	
	return op.Wait()
}

// Network operations

func (p *Provider) getNetwork(ctx context.Context, resource resources.Resource) (resources.Resource, error) {
	networkName := resource.Metadata.Name
	
	network, _, err := p.client.GetNetwork(networkName)
	if err != nil {
		return resource, err
	}
	
	// Convert Incus network to OmniGraph resource
	state := map[string]interface{}{
		"provider":    "incus",
		"type":        network.Type,
		"config":      network.Config,
		"managed":     network.Managed,
		"description": network.Description,
	}
	
	stateJSON, _ := json.Marshal(state)
	resource.Spec = stateJSON
	
	resource.Status = &resources.ResourceStatus{
		State: "active",
		Provider: &resources.ProviderStatus{
			ID: network.Name,
		},
	}
	
	return resource, nil
}

func (p *Provider) applyNetwork(ctx context.Context, desired resources.Resource) error {
	var spec resources.NetworkSpec
	if err := json.Unmarshal(desired.Spec, &spec); err != nil {
		return fmt.Errorf("failed to parse spec: %w", err)
	}
	
	networkName := desired.Metadata.Name
	
	// Check if network exists
	exists, _ := p.networkExists(ctx, desired)
	
	if exists {
		// Update network
		return p.updateNetwork(ctx, networkName, spec)
	}
	
	// Create network
	return p.createNetwork(ctx, networkName, spec)
}

func (p *Provider) createNetwork(ctx context.Context, name string, spec resources.NetworkSpec) error {
	req := api.NetworksPost{
		NetworkPut: api.NetworkPut{
			Config:      spec.Config,
			Description: spec.Description,
			Managed:     spec.Managed,
		},
		Name: name,
		Type: spec.Type,
	}
	
	_, err := p.client.CreateNetwork(req)
	if err != nil {
		return fmt.Errorf("failed to create network: %w", err)
	}
	
	return nil
}

func (p *Provider) updateNetwork(ctx context.Context, name string, spec resources.NetworkSpec) error {
	network, etag, err := p.client.GetNetwork(name)
	if err != nil {
		return err
	}
	
	networkPut := api.NetworkPut{
		Config:      spec.Config,
		Description: spec.Description,
		Managed:     spec.Managed,
	}
	
	err = p.client.UpdateNetwork(name, networkPut, etag)
	if err != nil {
		return fmt.Errorf("failed to update network: %w", err)
	}
	
	return nil
}

func (p *Provider) deleteNetwork(ctx context.Context, resource resources.Resource) error {
	networkName := resource.Metadata.Name
	
	err := p.client.DeleteNetwork(networkName)
	if err != nil {
		return fmt.Errorf("failed to delete network: %w", err)
	}
	
	return nil
}

func (p *Provider) networkExists(ctx context.Context, resource resources.Resource) (bool, error) {
	networkName := resource.Metadata.Name
	
	_, _, err := p.client.GetNetwork(networkName)
	if err != nil {
		return false, nil
	}
	
	return true, nil
}

// StoragePool operations

func (p *Provider) getStoragePool(ctx context.Context, resource resources.Resource) (resources.Resource, error) {
	poolName := resource.Metadata.Name
	
	pool, _, err := p.client.GetStoragePool(poolName)
	if err != nil {
		return resource, err
	}
	
	state := map[string]interface{}{
		"provider":    "incus",
		"driver":      pool.Driver,
		"config":      pool.Config,
		"description": pool.Description,
	}
	
	stateJSON, _ := json.Marshal(state)
	resource.Spec = stateJSON
	
	resource.Status = &resources.ResourceStatus{
		State: "active",
		Provider: &resources.ProviderStatus{
			ID: pool.Name,
		},
	}
	
	return resource, nil
}

func (p *Provider) applyStoragePool(ctx context.Context, desired resources.Resource) error {
	var spec resources.StoragePoolSpec
	if err := json.Unmarshal(desired.Spec, &spec); err != nil {
		return fmt.Errorf("failed to parse spec: %w", err)
	}
	
	poolName := desired.Metadata.Name
	
	exists, _ := p.storagePoolExists(ctx, desired)
	
	if exists {
		return p.updateStoragePool(ctx, poolName, spec)
	}
	
	return p.createStoragePool(ctx, poolName, spec)
}

func (p *Provider) createStoragePool(ctx context.Context, name string, spec resources.StoragePoolSpec) error {
	req := api.StoragePoolsPost{
		StoragePoolPut: api.StoragePoolPut{
			Config:      spec.Config,
			Description: spec.Description,
		},
		Name:   name,
		Driver: spec.Driver,
	}
	
	_, err := p.client.CreateStoragePool(req)
	if err != nil {
		return fmt.Errorf("failed to create storage pool: %w", err)
	}
	
	return nil
}

func (p *Provider) updateStoragePool(ctx context.Context, name string, spec resources.StoragePoolSpec) error {
	pool, etag, err := p.client.GetStoragePool(name)
	if err != nil {
		return err
	}
	
	poolPut := api.StoragePoolPut{
		Config:      spec.Config,
		Description: spec.Description,
	}
	
	err = p.client.UpdateStoragePool(name, poolPut, etag)
	if err != nil {
		return fmt.Errorf("failed to update storage pool: %w", err)
	}
	
	return nil
}

func (p *Provider) deleteStoragePool(ctx context.Context, resource resources.Resource) error {
	poolName := resource.Metadata.Name
	
	err := p.client.DeleteStoragePool(poolName)
	if err != nil {
		return fmt.Errorf("failed to delete storage pool: %w", err)
	}
	
	return nil
}

func (p *Provider) storagePoolExists(ctx context.Context, resource resources.Resource) (bool, error) {
	poolName := resource.Metadata.Name
	
	_, _, err := p.client.GetStoragePool(poolName)
	if err != nil {
		return false, nil
	}
	
	return true, nil
}

// Profile operations

func (p *Provider) getProfile(ctx context.Context, resource resources.Resource) (resources.Resource, error) {
	profileName := resource.Metadata.Name
	
	profile, _, err := p.client.GetProfile(profileName)
	if err != nil {
		return resource, err
	}
	
	state := map[string]interface{}{
		"provider":    "incus",
		"config":      profile.Config,
		"devices":     profile.Devices,
		"description": profile.Description,
	}
	
	stateJSON, _ := json.Marshal(state)
	resource.Spec = stateJSON
	
	resource.Status = &resources.ResourceStatus{
		State: "active",
		Provider: &resources.ProviderStatus{
			ID: profile.Name,
		},
	}
	
	return resource, nil
}

func (p *Provider) applyProfile(ctx context.Context, desired resources.Resource) error {
	var spec resources.ProfileSpec
	if err := json.Unmarshal(desired.Spec, &spec); err != nil {
		return fmt.Errorf("failed to parse spec: %w", err)
	}
	
	profileName := desired.Metadata.Name
	
	exists, _ := p.profileExists(ctx, desired)
	
	if exists {
		return p.updateProfile(ctx, profileName, spec)
	}
	
	return p.createProfile(ctx, profileName, spec)
}

func (p *Provider) createProfile(ctx context.Context, name string, spec resources.ProfileSpec) error {
	req := api.ProfilesPost{
		ProfilePut: api.ProfilePut{
			Config:      spec.Config,
			Devices:     convertDevices(spec.Devices),
			Description: spec.Description,
		},
		Name: name,
	}
	
	_, err := p.client.CreateProfile(req)
	if err != nil {
		return fmt.Errorf("failed to create profile: %w", err)
	}
	
	return nil
}

func (p *Provider) updateProfile(ctx context.Context, name string, spec resources.ProfileSpec) error {
	profile, etag, err := p.client.GetProfile(name)
	if err != nil {
		return err
	}
	
	profilePut := api.ProfilePut{
		Config:      spec.Config,
		Devices:     convertDevices(spec.Devices),
		Description: spec.Description,
	}
	
	err = p.client.UpdateProfile(name, profilePut, etag)
	if err != nil {
		return fmt.Errorf("failed to update profile: %w", err)
	}
	
	return nil
}

func (p *Provider) deleteProfile(ctx context.Context, resource resources.Resource) error {
	profileName := resource.Metadata.Name
	
	err := p.client.DeleteProfile(profileName)
	if err != nil {
		return fmt.Errorf("failed to delete profile: %w", err)
	}
	
	return nil
}

func (p *Provider) profileExists(ctx context.Context, resource resources.Resource) (bool, error) {
	profileName := resource.Metadata.Name
	
	_, _, err := p.client.GetProfile(profileName)
	if err != nil {
		return false, nil
	}
	
	return true, nil
}

// Helper functions

func convertDevices(devices map[string]resources.Device) map[string]map[string]string {
	result := make(map[string]map[string]string)
	
	for name, device := range devices {
		devMap := make(map[string]string)
		devMap["type"] = device.Type
		
		if device.Name != "" {
			devMap["name"] = device.Name
		}
		if device.Parent != "" {
			devMap["parent"] = device.Parent
		}
		if device.NICType != "" {
			devMap["nictype"] = device.NICType
		}
		if device.Path != "" {
			devMap["path"] = device.Path
		}
		if device.Pool != "" {
			devMap["pool"] = device.Pool
		}
		if device.Size != "" {
			devMap["size"] = device.Size
		}
		
		// Add extra properties
		for k, v := range device.Properties {
			devMap[k] = v
		}
		
		result[name] = devMap
	}
	
	return result
}

func specsEqual(a, b interface{}) bool {
	// Simple comparison - in production, use deep equal
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}