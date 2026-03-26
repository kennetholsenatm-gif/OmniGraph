package resources

import (
	"encoding/json"
	"fmt"
	"time"
)

// Resource represents a declarative infrastructure resource
type Resource struct {
	APIVersion string            `json:"apiVersion" yaml:"apiVersion"`
	Kind       string            `json:"kind" yaml:"kind"`
	Metadata   ResourceMetadata  `json:"metadata" yaml:"metadata"`
	Spec       json.RawMessage   `json:"spec" yaml:"spec"`
	Status     *ResourceStatus   `json:"status,omitempty" yaml:"status,omitempty"`
}

// ResourceMetadata contains resource identification and labels
type ResourceMetadata struct {
	Name        string            `json:"name" yaml:"name"`
	Namespace   string            `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	Labels      map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`
	UID         string            `json:"uid,omitempty" yaml:"uid,omitempty"`
}

// ResourceStatus represents the current state of a resource
type ResourceStatus struct {
	ObservedGeneration int64              `json:"observedGeneration,omitempty" yaml:"observedGeneration,omitempty"`
	State              string             `json:"state,omitempty" yaml:"state,omitempty"`
	Conditions         []Condition        `json:"conditions,omitempty" yaml:"conditions,omitempty"`
	Resources          *ResourceUsage     `json:"resources,omitempty" yaml:"resources,omitempty"`
	Reconciliation     *ReconciliationStatus `json:"reconciliation,omitempty" yaml:"reconciliation,omitempty"`
	Provider           *ProviderStatus    `json:"provider,omitempty" yaml:"provider,omitempty"`
}

// Condition represents a status condition (Kubernetes-style)
type Condition struct {
	Type               string    `json:"type" yaml:"type"`
	Status             string    `json:"status" yaml:"status"` // True, False, Unknown
	LastTransitionTime time.Time `json:"lastTransitionTime" yaml:"lastTransitionTime"`
	Reason             string    `json:"reason,omitempty" yaml:"reason,omitempty"`
	Message            string    `json:"message,omitempty" yaml:"message,omitempty"`
}

// ResourceUsage tracks resource consumption
type ResourceUsage struct {
	CPU     *ResourceMetric `json:"cpu,omitempty" yaml:"cpu,omitempty"`
	Memory  *ResourceMetric `json:"memory,omitempty" yaml:"memory,omitempty"`
	Storage map[string]*ResourceMetric `json:"storage,omitempty" yaml:"storage,omitempty"`
}

// ResourceMetric represents a used/limit metric
type ResourceMetric struct {
	Used  string `json:"used" yaml:"used"`
	Limit string `json:"limit" yaml:"limit"`
}

// ReconciliationStatus tracks reconciliation progress
type ReconciliationStatus struct {
	LastAttempt          time.Time `json:"lastAttempt" yaml:"lastAttempt"`
	LastSuccess          time.Time `json:"lastSuccess,omitempty" yaml:"lastSuccess,omitempty"`
	LastFailure          time.Time `json:"lastFailure,omitempty" yaml:"lastFailure,omitempty"`
	ConsecutiveSuccesses int       `json:"consecutiveSuccesses" yaml:"consecutiveSuccesses"`
	ConsecutiveFailures  int       `json:"consecutiveFailures" yaml:"consecutiveFailures"`
	LastError            string    `json:"lastError,omitempty" yaml:"lastError,omitempty"`
}

// ProviderStatus contains provider-specific status
type ProviderStatus struct {
	ID           string            `json:"id,omitempty" yaml:"id,omitempty"`
	Addresses    []string          `json:"addresses,omitempty" yaml:"addresses,omitempty"`
	Architecture string            `json:"architecture,omitempty" yaml:"architecture,omitempty"`
	CreatedAt    time.Time         `json:"createdAt,omitempty" yaml:"createdAt,omitempty"`
	UpdatedAt    time.Time         `json:"updatedAt,omitempty" yaml:"updatedAt,omitempty"`
	Extra        map[string]interface{} `json:"extra,omitempty" yaml:"extra,omitempty"`
}

// Manifest represents a collection of resources
type Manifest struct {
	APIVersion string     `json:"apiVersion" yaml:"apiVersion"`
	Kind       string     `json:"kind" yaml:"kind"`
	Metadata   ManifestMetadata `json:"metadata" yaml:"metadata"`
	Spec       ManifestSpec `json:"spec" yaml:"spec"`
}

// ManifestMetadata contains manifest identification
type ManifestMetadata struct {
	Name      string `json:"name" yaml:"name"`
	Namespace string `json:"namespace,omitempty" yaml:"namespace,omitempty"`
}

// ManifestSpec contains the desired state
type ManifestSpec struct {
	Resources       []Resource           `json:"resources" yaml:"resources"`
	Reconciliation  *ReconciliationPolicy `json:"reconciliation,omitempty" yaml:"reconciliation,omitempty"`
}

// ReconciliationPolicy defines how reconciliation should behave
type ReconciliationPolicy struct {
	Interval   string          `json:"interval,omitempty" yaml:"interval,omitempty"` // e.g., "5m"
	OnDrift    string          `json:"onDrift,omitempty" yaml:"onDrift,omitempty"` // auto, manual, alert
	RetryPolicy *RetryPolicy    `json:"retryPolicy,omitempty" yaml:"retryPolicy,omitempty"`
}

// RetryPolicy defines retry behavior
type RetryPolicy struct {
	MaxAttempts int    `json:"maxAttempts,omitempty" yaml:"maxAttempts,omitempty"`
	Backoff     string `json:"backoff,omitempty" yaml:"backoff,omitempty"` // fixed, exponential
}

// ComputeInstance represents an Incus container or VM
type ComputeInstance struct {
	APIVersion string                   `json:"apiVersion" yaml:"apiVersion"`
	Kind       string                   `json:"kind" yaml:"kind"`
	Metadata   ResourceMetadata         `json:"metadata" yaml:"metadata"`
	Spec       ComputeInstanceSpec      `json:"spec" yaml:"spec"`
	Status     *ComputeInstanceStatus   `json:"status,omitempty" yaml:"status,omitempty"`
}

// ComputeInstanceSpec defines the desired state of a compute instance
type ComputeInstanceSpec struct {
	Provider   string                 `json:"provider" yaml:"provider"`
	Type       string                 `json:"type" yaml:"type"` // container, virtual-machine
	Source     *InstanceSource        `json:"source,omitempty" yaml:"source,omitempty"`
	Config     map[string]string      `json:"config,omitempty" yaml:"config,omitempty"`
	Devices    map[string]Device      `json:"devices,omitempty" yaml:"devices,omitempty"`
	Profiles   []string               `json:"profiles,omitempty" yaml:"profiles,omitempty"`
	State      string                 `json:"state,omitempty" yaml:"state,omitempty"` // running, stopped, frozen
	Ephemeral  bool                   `json:"ephemeral,omitempty" yaml:"ephemeral,omitempty"`
}

// InstanceSource defines where to get the instance image
type InstanceSource struct {
	Server   string `json:"server,omitempty" yaml:"server,omitempty"`
	Protocol string `json:"protocol,omitempty" yaml:"protocol,omitempty"` // simplestreams, lxd
	Alias    string `json:"alias,omitempty" yaml:"alias,omitempty"`
	Image    string `json:"image,omitempty" yaml:"image,omitempty"`
	Type     string `json:"type,omitempty" yaml:"type,omitempty"` // image, migration
}

// Device represents an instance device
type Device struct {
	Type       string            `json:"type" yaml:"type"`
	Name       string            `json:"name,omitempty" yaml:"name,omitempty"`
	Parent     string            `json:"parent,omitempty" yaml:"parent,omitempty"`
	NICType    string            `json:"nictype,omitempty" yaml:"nictype,omitempty"`
	Path       string            `json:"path,omitempty" yaml:"path,omitempty"`
	Pool       string            `json:"pool,omitempty" yaml:"pool,omitempty"`
	Size       string            `json:"size,omitempty" yaml:"size,omitempty"`
	Properties map[string]string `json:"properties,omitempty" yaml:"properties,omitempty"`
}

// ComputeInstanceStatus represents the actual state
type ComputeInstanceStatus struct {
	ObservedGeneration int64              `json:"observedGeneration,omitempty" yaml:"observedGeneration,omitempty"`
	State              string             `json:"state,omitempty" yaml:"state,omitempty"`
	Conditions         []Condition        `json:"conditions,omitempty" yaml:"conditions,omitempty"`
	Resources          *ResourceUsage     `json:"resources,omitempty" yaml:"resources,omitempty"`
	Reconciliation     *ReconciliationStatus `json:"reconciliation,omitempty" yaml:"reconciliation,omitempty"`
	Provider           *ProviderStatus    `json:"provider,omitempty" yaml:"provider,omitempty"`
	Instance           *InstanceDetails   `json:"instance,omitempty" yaml:"instance,omitempty"`
}

// InstanceDetails contains Incus-specific instance details
type InstanceDetails struct {
	ID           string            `json:"id,omitempty" yaml:"id,omitempty"`
	Name         string            `json:"name,omitempty" yaml:"name,omitempty"`
	Type         string            `json:"type,omitempty" yaml:"type,omitempty"`
	State        string            `json:"state,omitempty" yaml:"state,omitempty"`
	Status       string            `json:"status,omitempty" yaml:"status,omitempty"`
	IPv4         []string          `json:"ipv4,omitempty" yaml:"ipv4,omitempty"`
	IPv6         []string          `json:"ipv6,omitempty" yaml:"ipv6,omitempty"`
	Architecture string            `json:"architecture,omitempty" yaml:"architecture,omitempty"`
	CreatedAt    time.Time         `json:"createdAt,omitempty" yaml:"createdAt,omitempty"`
	UpdatedAt    time.Time         `json:"updatedAt,omitempty" yaml:"updatedAt,omitempty"`
	LastUsedAt   time.Time         `json:"lastUsedAt,omitempty" yaml:"lastUsedAt,omitempty"`
	Location     string            `json:"location,omitempty" yaml:"location,omitempty"`
	Project      string            `json:"project,omitempty" yaml:"project,omitempty"`
	Ephemeral    bool              `json:"ephemeral,omitempty" yaml:"ephemeral,omitempty"`
	Profiles     []string          `json:"profiles,omitempty" yaml:"profiles,omitempty"`
	Config       map[string]string `json:"config,omitempty" yaml:"config,omitempty"`
	Devices      map[string]Device `json:"devices,omitempty" yaml:"devices,omitempty"`
}

// Network represents an Incus network
type Network struct {
	APIVersion string           `json:"apiVersion" yaml:"apiVersion"`
	Kind       string           `json:"kind" yaml:"kind"`
	Metadata   ResourceMetadata `json:"metadata" yaml:"metadata"`
	Spec       NetworkSpec      `json:"spec" yaml:"spec"`
	Status     *NetworkStatus   `json:"status,omitempty" yaml:"status,omitempty"`
}

// NetworkSpec defines the desired state of a network
type NetworkSpec struct {
	Provider    string            `json:"provider" yaml:"provider"`
	Type        string            `json:"type,omitempty" yaml:"type,omitempty"` // bridge, ovn, macvlan, sriov
	Config      map[string]string `json:"config,omitempty" yaml:"config,omitempty"`
	Description string            `json:"description,omitempty" yaml:"description,omitempty"`
	Managed     *bool             `json:"managed,omitempty" yaml:"managed,omitempty"`
}

// NetworkStatus represents the actual state
type NetworkStatus struct {
	ObservedGeneration int64              `json:"observedGeneration,omitempty" yaml:"observedGeneration,omitempty"`
	State              string             `json:"state,omitempty" yaml:"state,omitempty"`
	Conditions         []Condition        `json:"conditions,omitempty" yaml:"conditions,omitempty"`
	Reconciliation     *ReconciliationStatus `json:"reconciliation,omitempty" yaml:"reconciliation,omitempty"`
	Provider           *ProviderStatus    `json:"provider,omitempty" yaml:"provider,omitempty"`
	Network            *NetworkDetails    `json:"network,omitempty" yaml:"network,omitempty"`
}

// NetworkDetails contains Incus-specific network details
type NetworkDetails struct {
	Name        string            `json:"name,omitempty" yaml:"name,omitempty"`
	Type        string            `json:"type,omitempty" yaml:"type,omitempty"`
	Managed     bool              `json:"managed,omitempty" yaml:"managed,omitempty"`
	IPv4        string            `json:"ipv4,omitempty" yaml:"ipv4,omitempty"`
	IPv6        string            `json:"ipv6,omitempty" yaml:"ipv6,omitempty"`
	UsedBy      []string          `json:"usedBy,omitempty" yaml:"usedBy,omitempty"`
	Config      map[string]string `json:"config,omitempty" yaml:"config,omitempty"`
	Description string            `json:"description,omitempty" yaml:"description,omitempty"`
}

// StoragePool represents an Incus storage pool
type StoragePool struct {
	APIVersion string             `json:"apiVersion" yaml:"apiVersion"`
	Kind       string             `json:"kind" yaml:"kind"`
	Metadata   ResourceMetadata   `json:"metadata" yaml:"metadata"`
	Spec       StoragePoolSpec    `json:"spec" yaml:"spec"`
	Status     *StoragePoolStatus `json:"status,omitempty" yaml:"status,omitempty"`
}

// StoragePoolSpec defines the desired state of a storage pool
type StoragePoolSpec struct {
	Provider    string            `json:"provider" yaml:"provider"`
	Driver      string            `json:"driver" yaml:"driver"` // dir, btrfs, lvm, zfs, ceph, cephfs
	Config      map[string]string `json:"config,omitempty" yaml:"config,omitempty"`
	Description string            `json:"description,omitempty" yaml:"description,omitempty"`
}

// StoragePoolStatus represents the actual state
type StoragePoolStatus struct {
	ObservedGeneration int64              `json:"observedGeneration,omitempty" yaml:"observedGeneration,omitempty"`
	State              string             `json:"state,omitempty" yaml:"state,omitempty"`
	Conditions         []Condition        `json:"conditions,omitempty" yaml:"conditions,omitempty"`
	Reconciliation     *ReconciliationStatus `json:"reconciliation,omitempty" yaml:"reconciliation,omitempty"`
	Provider           *ProviderStatus    `json:"provider,omitempty" yaml:"provider,omitempty"`
	Pool               *StoragePoolDetails `json:"pool,omitempty" yaml:"pool,omitempty"`
}

// StoragePoolDetails contains Incus-specific pool details
type StoragePoolDetails struct {
	Name        string            `json:"name,omitempty" yaml:"name,omitempty"`
	Driver      string            `json:"driver,omitempty" yaml:"driver,omitempty"`
	Source      string            `json:"source,omitempty" yaml:"source,omitempty"`
	Config      map[string]string `json:"config,omitempty" yaml:"config,omitempty"`
	UsedBy      []string          `json:"usedBy,omitempty" yaml:"usedBy,omitempty"`
	Description string            `json:"description,omitempty" yaml:"description,omitempty"`
}

// Profile represents an Incus profile
type Profile struct {
	APIVersion string           `json:"apiVersion" yaml:"apiVersion"`
	Kind       string           `json:"kind" yaml:"kind"`
	Metadata   ResourceMetadata `json:"metadata" yaml:"metadata"`
	Spec       ProfileSpec      `json:"spec" yaml:"spec"`
	Status     *ProfileStatus   `json:"status,omitempty" yaml:"status,omitempty"`
}

// ProfileSpec defines the desired state of a profile
type ProfileSpec struct {
	Provider    string            `json:"provider" yaml:"provider"`
	Config      map[string]string `json:"config,omitempty" yaml:"config,omitempty"`
	Devices     map[string]Device `json:"devices,omitempty" yaml:"devices,omitempty"`
	Description string            `json:"description,omitempty" yaml:"description,omitempty"`
}

// ProfileStatus represents the actual state
type ProfileStatus struct {
	ObservedGeneration int64              `json:"observedGeneration,omitempty" yaml:"observedGeneration,omitempty"`
	Conditions         []Condition        `json:"conditions,omitempty" yaml:"conditions,omitempty"`
	Reconciliation     *ReconciliationStatus `json:"reconciliation,omitempty" yaml:"reconciliation,omitempty"`
	Provider           *ProviderStatus    `json:"provider,omitempty" yaml:"provider,omitempty"`
	Profile            *ProfileDetails    `json:"profile,omitempty" yaml:"profile,omitempty"`
}

// ProfileDetails contains Incus-specific profile details
type ProfileDetails struct {
	Name        string            `json:"name,omitempty" yaml:"name,omitempty"`
	Config      map[string]string `json:"config,omitempty" yaml:"config,omitempty"`
	Devices     map[string]Device `json:"devices,omitempty" yaml:"devices,omitempty"`
	UsedBy      []string          `json:"usedBy,omitempty" yaml:"usedBy,omitempty"`
	Description string            `json:"description,omitempty" yaml:"description,omitempty"`
}

// ResourceEvent represents a resource change event
type ResourceEvent struct {
	Type      string    `json:"type" yaml:"type"`
	Resource  Resource  `json:"resource" yaml:"resource"`
	Timestamp time.Time `json:"timestamp" yaml:"timestamp"`
}

// EventType constants
const (
	EventTypeCreated = "Created"
	EventTypeUpdated = "Updated"
	EventTypeDeleted = "Deleted"
	EventTypeError   = "Error"
)

// Helper functions

// NewCondition creates a new condition
func NewCondition(condType, status, reason, message string) Condition {
	return Condition{
		Type:               condType,
		Status:             status,
		LastTransitionTime: time.Now(),
		Reason:             reason,
		Message:            message,
	}
}

// IsConditionTrue checks if a condition is true
func IsConditionTrue(conditions []Condition, condType string) bool {
	for _, c := range conditions {
		if c.Type == condType && c.Status == "True" {
			return true
		}
	}
	return false
}

// GetCondition returns a condition by type
func GetCondition(conditions []Condition, condType string) *Condition {
	for i, c := range conditions {
		if c.Type == condType {
			return &conditions[i]
		}
	}
	return nil
}

// SetCondition sets or updates a condition
func SetCondition(conditions *[]Condition, newCond Condition) {
	for i, c := range *conditions {
		if c.Type == newCond.Type {
			if c.Status != newCond.Status {
				(*conditions)[i] = newCond
			}
			return
		}
	}
	*conditions = append(*conditions, newCond)
}

// Validate validates a resource
func (r *Resource) Validate() error {
	if r.APIVersion == "" {
		return fmt.Errorf("apiVersion is required")
	}
	if r.Kind == "" {
		return fmt.Errorf("kind is required")
	}
	if r.Metadata.Name == "" {
		return fmt.Errorf("metadata.name is required")
	}
	return nil
}

// GetSpec unmarshals the spec into the provided interface
func (r *Resource) GetSpec(target interface{}) error {
	return json.Unmarshal(r.Spec, target)
}