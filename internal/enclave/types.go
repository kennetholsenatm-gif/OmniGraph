package enclave

import (
	"time"
)

// Enclave represents a Wasm enclave configuration
type Enclave struct {
	APIVersion string          `json:"apiVersion" yaml:"apiVersion"`
	Kind       string          `json:"kind" yaml:"kind"`
	Metadata   EnclaveMetadata `json:"metadata" yaml:"metadata"`
	Spec       EnclaveSpec     `json:"spec" yaml:"spec"`
	Status     *EnclaveStatus  `json:"status,omitempty" yaml:"status,omitempty"`
}

// EnclaveMetadata contains enclave identification
type EnclaveMetadata struct {
	Name        string            `json:"name" yaml:"name"`
	Namespace   string            `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	Labels      map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`
}

// EnclaveSpec defines the desired state
type EnclaveSpec struct {
	DependsOn          []string          `json:"dependsOn,omitempty" yaml:"dependsOn,omitempty"`
	DeploymentStrategy string            `json:"deploymentStrategy" yaml:"deploymentStrategy"`
	Replicas           int               `json:"replicas" yaml:"replicas"`
	Runtime            RuntimeConfig     `json:"runtime" yaml:"runtime"`
	TrustBoundary      TrustBoundary     `json:"trustBoundary" yaml:"trustBoundary"`
	CognitivePayload   CognitivePayload  `json:"cognitivePayload" yaml:"cognitivePayload"`
	Routing            *RoutingConfig    `json:"routing,omitempty" yaml:"routing,omitempty"`
	Resources          *ResourceLimits   `json:"resources,omitempty" yaml:"resources,omitempty"`
	Environment        map[string]string `json:"environment,omitempty" yaml:"environment,omitempty"`
	Volumes            []Volume          `json:"volumes,omitempty" yaml:"volumes,omitempty"`
	HealthCheck        *HealthCheck      `json:"healthCheck,omitempty" yaml:"healthCheck,omitempty"`
	Scaling            *ScalingConfig    `json:"scaling,omitempty" yaml:"scaling,omitempty"`
}

// RuntimeConfig defines Wasm runtime settings
type RuntimeConfig struct {
	Engine                 string `json:"engine" yaml:"engine"`
	MemoryLimitMb          int    `json:"memoryLimitMb" yaml:"memoryLimitMb"`
	CpuLimitMs             int    `json:"cpuLimitMs,omitempty" yaml:"cpuLimitMs,omitempty"`
	DeterministicExecution bool   `json:"deterministicExecution" yaml:"deterministicExecution"`
	NetworkAccess          bool   `json:"networkAccess" yaml:"networkAccess"`
	FilesystemAccess       string `json:"filesystemAccess" yaml:"filesystemAccess"`
	MaxInstances           int    `json:"maxInstances" yaml:"maxInstances"`
}

// TrustBoundary defines ZTEE enrollment settings
type TrustBoundary struct {
	Enrollment          string   `json:"enrollment" yaml:"enrollment"`
	AttestationProvider string   `json:"attestationProvider,omitempty" yaml:"attestationProvider,omitempty"`
	AllowedPeers        []string `json:"allowedPeers,omitempty" yaml:"allowedPeers,omitempty"`
	CertificateRotation string   `json:"certificateRotation" yaml:"certificateRotation"`
	AuditLog            bool     `json:"auditLog" yaml:"auditLog"`
}

// CognitivePayload defines ML model configuration
type CognitivePayload struct {
	SourceURI     string               `json:"sourceUri" yaml:"sourceUri"`
	WeightFormat  string               `json:"weightFormat" yaml:"weightFormat"`
	Checksum      string               `json:"checksum,omitempty" yaml:"checksum,omitempty"`
	Signature     string               `json:"signature,omitempty" yaml:"signature,omitempty"`
	InferenceMode string               `json:"inferenceMode" yaml:"inferenceMode"`
	Preprocessing *PreprocessingConfig `json:"preprocessing,omitempty" yaml:"preprocessing,omitempty"`
}

// PreprocessingConfig defines model preprocessing
type PreprocessingConfig struct {
	Normalize bool   `json:"normalize,omitempty" yaml:"normalize,omitempty"`
	Resize    string `json:"resize,omitempty" yaml:"resize,omitempty"`
	Quantize  bool   `json:"quantize,omitempty" yaml:"quantize,omitempty"`
}

// RoutingConfig defines routing strategy
type RoutingConfig struct {
	Strategy           string           `json:"strategy" yaml:"strategy"`
	ClassicalHeuristic string           `json:"classicalHeuristic" yaml:"classicalHeuristic"`
	QuantumFallback    *QuantumFallback `json:"quantumFallback,omitempty" yaml:"quantumFallback,omitempty"`
	Triggers           []RoutingTrigger `json:"triggers,omitempty" yaml:"triggers,omitempty"`
}

// QuantumFallback defines quantum routing configuration
type QuantumFallback struct {
	Enabled   bool   `json:"enabled" yaml:"enabled"`
	Provider  string `json:"provider,omitempty" yaml:"provider,omitempty"`
	Algorithm string `json:"algorithm" yaml:"algorithm"`
	Threshold string `json:"threshold,omitempty" yaml:"threshold,omitempty"`
	MaxQubits int    `json:"maxQubits" yaml:"maxQubits"`
}

// RoutingTrigger defines when to switch routing strategy
type RoutingTrigger struct {
	Metric    string `json:"metric" yaml:"metric"`
	Condition string `json:"condition" yaml:"condition"`
	Action    string `json:"action" yaml:"action"`
}

// ResourceLimits defines resource constraints
type ResourceLimits struct {
	Requests ResourceList `json:"requests,omitempty" yaml:"requests,omitempty"`
	Limits   ResourceList `json:"limits,omitempty" yaml:"limits,omitempty"`
}

// ResourceList defines resource quantities
type ResourceList struct {
	CPU     string `json:"cpu,omitempty" yaml:"cpu,omitempty"`
	Memory  string `json:"memory,omitempty" yaml:"memory,omitempty"`
	Storage string `json:"storage,omitempty" yaml:"storage,omitempty"`
}

// Volume defines a mounted volume
type Volume struct {
	Name      string `json:"name" yaml:"name"`
	MountPath string `json:"mountPath" yaml:"mountPath"`
	ReadOnly  bool   `json:"readOnly" yaml:"readOnly"`
	Source    string `json:"source,omitempty" yaml:"source,omitempty"`
}

// HealthCheck defines health check configuration
type HealthCheck struct {
	Enabled          bool   `json:"enabled" yaml:"enabled"`
	Endpoint         string `json:"endpoint" yaml:"endpoint"`
	IntervalSeconds  int    `json:"intervalSeconds" yaml:"intervalSeconds"`
	TimeoutSeconds   int    `json:"timeoutSeconds" yaml:"timeoutSeconds"`
	FailureThreshold int    `json:"failureThreshold" yaml:"failureThreshold"`
}

// ScalingConfig defines auto-scaling settings
type ScalingConfig struct {
	MinReplicas       int    `json:"minReplicas" yaml:"minReplicas"`
	MaxReplicas       int    `json:"maxReplicas" yaml:"maxReplicas"`
	TargetCPUPercent  int    `json:"targetCpuPercent" yaml:"targetCpuPercent"`
	ScaleUpCooldown   string `json:"scaleUpCooldown" yaml:"scaleUpCooldown"`
	ScaleDownCooldown string `json:"scaleDownCooldown" yaml:"scaleDownCooldown"`
}

// EnclaveStatus represents the observed state
type EnclaveStatus struct {
	Phase            string            `json:"phase" yaml:"phase"`
	Conditions       []Condition       `json:"conditions,omitempty" yaml:"conditions,omitempty"`
	EnrollmentStatus *EnrollmentStatus `json:"enrollmentStatus,omitempty" yaml:"enrollmentStatus,omitempty"`
	RuntimeMetrics   *RuntimeMetrics   `json:"runtimeMetrics,omitempty" yaml:"runtimeMetrics,omitempty"`
}

// Condition represents a status condition
type Condition struct {
	Type               string    `json:"type" yaml:"type"`
	Status             string    `json:"status" yaml:"status"`
	LastTransitionTime time.Time `json:"lastTransitionTime" yaml:"lastTransitionTime"`
	Reason             string    `json:"reason,omitempty" yaml:"reason,omitempty"`
	Message            string    `json:"message,omitempty" yaml:"message,omitempty"`
}

// EnrollmentStatus represents ZTEE enrollment state
type EnrollmentStatus struct {
	Enrolled          bool      `json:"enrolled" yaml:"enrolled"`
	AttestedAt        time.Time `json:"attestedAt,omitempty" yaml:"attestedAt,omitempty"`
	CertificateExpiry time.Time `json:"certificateExpiry,omitempty" yaml:"certificateExpiry,omitempty"`
}

// RuntimeMetrics represents runtime performance data
type RuntimeMetrics struct {
	MemoryUsageMb  float64 `json:"memoryUsageMb" yaml:"memoryUsageMb"`
	CPUPercent     float64 `json:"cpuUsagePercent" yaml:"cpuUsagePercent"`
	InferenceCount int64   `json:"inferenceCount" yaml:"inferenceCount"`
	AvgLatencyMs   float64 `json:"avgLatencyMs" yaml:"avgLatencyMs"`
}
