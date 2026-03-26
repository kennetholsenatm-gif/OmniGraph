package reconcile

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/kennetholsenatm-gif/omnigraph/internal/resources"
)

// Controller implements a Kubernetes-style reconciliation loop
type Controller struct {
	mu           sync.RWMutex
	providers    map[string]Provider
	resources    map[string]*ResourceEntry
	watcher      StateWatcher
	reconciler   Reconciler
	statusMgr    StatusManager
	eventBus     EventBus
	interval     time.Duration
	stopCh       chan struct{}
}

// ResourceEntry tracks a managed resource
type ResourceEntry struct {
	Desired    *resources.Resource
	Actual     resources.Resource
	LastSync   time.Time
	Reconciled bool
}

// Provider interface for resource providers
type Provider interface {
	// GetActualState returns the actual state from the provider
	GetActualState(ctx context.Context, resource resources.Resource) (resources.Resource, error)
	
	// Apply applies the desired state to the provider
	Apply(ctx context.Context, desired resources.Resource) error
	
	// Delete deletes the resource from the provider
	Delete(ctx context.Context, resource resources.Resource) error
	
	// Exists checks if the resource exists
	Exists(ctx context.Context, resource resources.Resource) (bool, error)
	
	// Watch watches for changes
	Watch(ctx context.Context, resource resources.Resource) (<-chan resources.ResourceEvent, error)
}

// StateWatcher watches for state changes
type StateWatcher interface {
	// Watch watches a resource for changes
	Watch(ctx context.Context, resource resources.Resource) (<-chan StateChange, error)
	
	// Stop stops watching
	Stop(resourceKey string)
}

// Reconciler reconciles desired state with actual state
type Reconciler interface {
	// Reconcile performs reconciliation
	Reconcile(ctx context.Context, desired, actual resources.Resource) (ReconcileResult, error)
	
	// Diff computes the difference between desired and actual
	Diff(desired, actual resources.Resource) (DiffResult, error)
}

// StatusManager manages resource status
type StatusManager interface {
	// UpdateStatus updates the resource status
	UpdateStatus(ctx context.Context, resourceKey string, status resources.ResourceStatus) error
	
	// GetStatus gets the resource status
	GetStatus(ctx context.Context, resourceKey string) (*resources.ResourceStatus, error)
	
	// SetCondition sets a condition on the resource
	SetCondition(ctx context.Context, resourceKey string, condition resources.Condition) error
}

// EventBus publishes events
type EventBus interface {
	// Publish publishes an event
	Publish(event Event)
	
	// Subscribe subscribes to events
	Subscribe(eventType string, handler EventHandler) Subscription
	
	// Unsubscribe unsubscribes
	Unsubscribe(sub Subscription)
}

// ResourceEvent represents a resource change event
type ResourceEvent struct {
	Type      EventType
	Resource  resources.Resource
	Timestamp time.Time
}

// EventType represents the type of event
type EventType string

const (
	EventTypeCreated  EventType = "Created"
	EventTypeUpdated  EventType = "Updated"
	EventTypeDeleted  EventType = "Deleted"
	EventTypeError    EventType = "Error"
)

// StateChange represents a state change
type StateChange struct {
	ResourceKey string
	OldState    resources.Resource
	NewState    resources.Resource
	Timestamp   time.Time
}

// ReconcileResult represents the result of reconciliation
type ReconcileResult struct {
	Requeue      bool
	RequeueAfter time.Duration
	Error        error
}

// DiffResult represents the difference between desired and actual state
type DiffResult struct {
	NeedsCreate  bool
	NeedsUpdate  bool
	NeedsDelete  bool
	Changes      []Change
}

// Change represents a single change
type Change struct {
	Path     string
	OldValue interface{}
	NewValue interface{}
}

// Event represents an event
type Event struct {
	Type      string
	Resource  string
	Timestamp time.Time
	Data      interface{}
}

// EventHandler handles events
type EventHandler func(event Event)

// Subscription represents an event subscription
type Subscription struct {
	ID      string
	Handler EventHandler
}

// NewController creates a new reconciliation controller
func NewController(interval time.Duration) *Controller {
	return &Controller{
		providers: make(map[string]Provider),
		resources: make(map[string]*ResourceEntry),
		interval:  interval,
		stopCh:    make(chan struct{}),
	}
}

// RegisterProvider registers a provider
func (c *Controller) RegisterProvider(name string, provider Provider) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.providers[name] = provider
}

// GetProvider returns a provider by name
func (c *Controller) GetProvider(name string) Provider {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.providers[name]
}

// AddResource adds a resource to be managed
func (c *Controller) AddResource(ctx context.Context, resource resources.Resource) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	key := resourceKey(resource)
	c.resources[key] = &ResourceEntry{
		Desired:  &resource,
		LastSync: time.Now(),
	}
	
	return nil
}

// RemoveResource removes a resource from management
func (c *Controller) RemoveResource(ctx context.Context, resource resources.Resource) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	key := resourceKey(resource)
	delete(c.resources, key)
	
	return nil
}

// Run starts the reconciliation loop
func (c *Controller) Run(ctx context.Context) error {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-c.stopCh:
			return nil
		case <-ticker.C:
			if err := c.reconcileAll(ctx); err != nil {
				// Log error but continue
				fmt.Printf("Reconciliation error: %v\n", err)
			}
		}
	}
}

// Stop stops the controller
func (c *Controller) Stop() {
	close(c.stopCh)
}

// reconcileAll reconciles all managed resources
func (c *Controller) reconcileAll(ctx context.Context) error {
	c.mu.RLock()
	resources := make([]*ResourceEntry, 0, len(c.resources))
	for _, entry := range c.resources {
		resources = append(resources, entry)
	}
	c.mu.RUnlock()
	
	for _, entry := range resources {
		if err := c.reconcileResource(ctx, entry); err != nil {
			fmt.Printf("Failed to reconcile %s: %v\n", entry.Desired.Metadata.Name, err)
		}
	}
	
	return nil
}

// reconcileResource reconciles a single resource
func (c *Controller) reconcileResource(ctx context.Context, entry *ResourceEntry) error {
	// Unmarshal spec to get provider
	var spec struct {
		Provider string `json:"provider"`
	}
	if err := json.Unmarshal(entry.Desired.Spec, &spec); err != nil {
		return fmt.Errorf("failed to unmarshal spec: %w", err)
	}
	
	provider, ok := c.providers[spec.Provider]
	if !ok {
		return fmt.Errorf("provider %s not found", spec.Provider)
	}
	
	// Get actual state
	actual, err := provider.GetActualState(ctx, *entry.Desired)
	if err != nil {
		// Resource doesn't exist, create it
		if err := provider.Apply(ctx, *entry.Desired); err != nil {
			return fmt.Errorf("failed to create resource: %w", err)
		}
		entry.Reconciled = true
		entry.LastSync = time.Now()
		return nil
	}
	
	entry.Actual = actual
	
	// Compute diff
	diff, err := c.computeDiff(*entry.Desired, actual)
	if err != nil {
		return fmt.Errorf("failed to compute diff: %w", err)
	}
	
	// Apply changes if needed
	if diff.NeedsUpdate {
		if err := provider.Apply(ctx, *entry.Desired); err != nil {
			return fmt.Errorf("failed to update resource: %w", err)
		}
		entry.Reconciled = true
		entry.LastSync = time.Now()
	}
	
	return nil
}

// computeDiff computes the difference between desired and actual state
func (c *Controller) computeDiff(desired, actual resources.Resource) (DiffResult, error) {
	result := DiffResult{}
	
	// Compare specs
	if !specsEqual(desired.Spec, actual.Spec) {
		result.NeedsUpdate = true
		result.Changes = append(result.Changes, Change{
			Path:     "spec",
			OldValue: actual.Spec,
			NewValue: desired.Spec,
		})
	}
	
	return result, nil
}

// specsEqual compares two specs
func specsEqual(a, b interface{}) bool {
	// Simple comparison - in production, use deep equal
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

// resourceKey generates a unique key for a resource
func resourceKey(r resources.Resource) string {
	return fmt.Sprintf("%s/%s/%s", r.APIVersion, r.Kind, r.Metadata.Name)
}