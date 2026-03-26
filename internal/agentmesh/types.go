package agentmesh

import (
	"context"
	"time"
)

// Event represents a message in the event mesh
type Event struct {
	ID            string                 `json:"id"`
	Topic         string                 `json:"topic"`
	CorrelationID string                 `json:"correlationId,omitempty"`
	Timestamp     time.Time              `json:"timestamp"`
	Payload       map[string]interface{} `json:"payload"`
	Metadata      map[string]string      `json:"metadata,omitempty"`
}

// EventHandler processes events from the broker
type EventHandler func(ctx context.Context, event Event) error

// Subscription represents an active event subscription
type Subscription struct {
	ID      string
	Topic   string
	Handler EventHandler
	Cancel  context.CancelFunc
}

// AgentConfig defines configuration for a Wasm agent
type AgentConfig struct {
	ID          string            `json:"id"`
	Type        string            `json:"type"`
	Module      string            `json:"module"`
	Description string            `json:"description,omitempty"`
	Resources   ResourceLimits    `json:"resources"`
	Triggers    []Trigger         `json:"triggers,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
	ZTEE        *ZTEEConfig       `json:"ztee,omitempty"`
}

// ResourceLimits defines resource constraints for an agent
type ResourceLimits struct {
	Memory  string `json:"memory,omitempty"`
	CPU     string `json:"cpu,omitempty"`
	Storage string `json:"storage,omitempty"`
}

// Trigger defines an event trigger for agent execution
type Trigger struct {
	Topic         string `json:"topic"`
	Filter        string `json:"filter,omitempty"`
	CorrelationID string `json:"correlationId,omitempty"`
}

// ZTEEConfig defines Zero-Trust Enclave Enrollment settings
type ZTEEConfig struct {
	Enabled   bool   `json:"enabled"`
	EnrollURL string `json:"enrollUrl,omitempty"`
}

// Agent represents a running Wasm agent instance
type Agent struct {
	ID        string
	Config    AgentConfig
	Status    AgentStatus
	CreatedAt time.Time
	StartedAt *time.Time
}

// AgentStatus represents the current state of an agent
type AgentStatus string

const (
	AgentStatusPending   AgentStatus = "pending"
	AgentStatusRunning   AgentStatus = "running"
	AgentStatusStopped   AgentStatus = "stopped"
	AgentStatusFailed    AgentStatus = "failed"
	AgentStatusEnrolling AgentStatus = "enrolling"
)

// AgentState contains runtime state of an agent
type AgentState struct {
	AgentID       string
	Status        AgentStatus
	Health        HealthStatus
	ModuleLoaded  bool
	LastHeartbeat time.Time
	Metrics       AgentMetrics
}

// HealthStatus represents agent health
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
)

// AgentMetrics contains performance metrics
type AgentMetrics struct {
	EventsProcessed  int64
	AverageLatency   time.Duration
	ErrorCount       int64
	MemoryUsageBytes uint64
	CPUUsagePercent  float64
}

// InferenceInput represents input for ML inference
type InferenceInput struct {
	ModelName string                 `json:"modelName"`
	Data      map[string]interface{} `json:"data"`
	Options   map[string]interface{} `json:"options,omitempty"`
}

// InferenceResult represents output from ML inference
type InferenceResult struct {
	Prediction  interface{}            `json:"prediction"`
	Confidence  float64                `json:"confidence"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	ProcessedAt time.Time              `json:"processedAt"`
}

// StateChange represents an infrastructure state change
type StateChange struct {
	ResourceType string                 `json:"resourceType"`
	ResourceID   string                 `json:"resourceId"`
	Action       string                 `json:"action"` // created, updated, deleted
	OldState     map[string]interface{} `json:"oldState,omitempty"`
	NewState     map[string]interface{} `json:"newState,omitempty"`
	Timestamp    time.Time              `json:"timestamp"`
}

// StatePattern defines a pattern for matching state changes
type StatePattern struct {
	ResourceType string `json:"resourceType,omitempty"`
	Action       string `json:"action,omitempty"`
	FieldPath    string `json:"fieldPath,omitempty"`
}

// AgentResult represents the result of agent computation
type AgentResult struct {
	AgentID     string                 `json:"agentID"`
	EventID     string                 `json:"eventId"`
	Success     bool                   `json:"success"`
	Result      map[string]interface{} `json:"result,omitempty"`
	Error       string                 `json:"error,omitempty"`
	ProcessedAt time.Time              `json:"processedAt"`
	Latency     time.Duration          `json:"latency"`
}

// InventoryNode represents a node in the inventory for optimization
type InventoryNode struct {
	ID       string            `json:"id"`
	Type     string            `json:"type"`
	Labels   map[string]string `json:"labels,omitempty"`
	Host     string            `json:"host,omitempty"`
	Features []string          `json:"features,omitempty"`
}

// Playbook represents an Ansible playbook for optimization
type Playbook struct {
	Name    string   `json:"name"`
	Path    string   `json:"path"`
	Tags    []string `json:"tags,omitempty"`
	Depends []string `json:"depends,omitempty"`
}

// Host represents a target host for optimization
type Host struct {
	Name     string            `json:"name"`
	Address  string            `json:"address"`
	Labels   map[string]string `json:"labels,omitempty"`
	Features []string          `json:"features,omitempty"`
}

// OptimizedPath represents an optimized inventory traversal
type OptimizedPath struct {
	Nodes     []InventoryNode `json:"nodes"`
	Order     []string        `json:"order"`
	TotalCost float64         `json:"totalCost"`
	Generated time.Time       `json:"generated"`
}

// ExecutionPlan represents an optimized execution plan
type ExecutionPlan struct {
	Playbooks []Playbook    `json:"playbooks"`
	Order     []string      `json:"order"`
	Parallel  [][]string    `json:"parallel,omitempty"`
	Estimated time.Duration `json:"estimated"`
	Generated time.Time     `json:"generated"`
}

// TelemetryData represents telemetry data for traffic optimization
type TelemetryData struct {
	Source    string    `json:"source"`
	Dest      string    `json:"dest"`
	Bytes     int64     `json:"bytes"`
	Latency   float64   `json:"latency"`
	Timestamp time.Time `json:"timestamp"`
}

// TrafficPolicy represents optimized traffic routing
type TrafficPolicy struct {
	Rules     []TrafficRule `json:"rules"`
	Generated time.Time     `json:"generated"`
}

// TrafficRule represents a single traffic routing rule
type TrafficRule struct {
	Source   string `json:"source"`
	Dest     string `json:"dest"`
	Action   string `json:"action"`
	Priority int    `json:"priority"`
}
