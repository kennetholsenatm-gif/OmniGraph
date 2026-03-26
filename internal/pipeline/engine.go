package pipeline

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Engine represents the pipeline execution engine
type Engine struct {
	mu          sync.RWMutex
	definitions map[string]*Definition
	executions  map[string]*Execution
	history     []*ExecutionRecord
}

// Definition represents a parsed pipeline definition
type Definition struct {
	APIVersion string       `json:"apiVersion"`
	Kind       string       `json:"kind"`
	Metadata   Metadata     `json:"metadata"`
	Spec       PipelineSpec `json:"spec"`
}

// Metadata contains pipeline metadata
type Metadata struct {
	Name        string            `json:"name"`
	Version     string            `json:"version,omitempty"`
	Description string            `json:"description,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

// PipelineSpec defines the pipeline specification
type PipelineSpec struct {
	Triggers      []Trigger           `json:"triggers,omitempty"`
	Variables     map[string]Variable `json:"variables,omitempty"`
	Stages        []Stage             `json:"stages"`
	Notifications *Notifications      `json:"notifications,omitempty"`
	Timeout       string              `json:"timeout,omitempty"`
	RetryPolicy   *RetryPolicy        `json:"retryPolicy,omitempty"`
}

// Trigger defines a pipeline trigger
type Trigger struct {
	Type   string                 `json:"type"`
	Config map[string]interface{} `json:"config,omitempty"`
}

// Variable defines a pipeline variable
type Variable struct {
	Type        string      `json:"type"`
	Default     interface{} `json:"default,omitempty"`
	Required    bool        `json:"required,omitempty"`
	Description string      `json:"description,omitempty"`
}

// Stage represents a pipeline stage
type Stage struct {
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	DependsOn   []string          `json:"dependsOn,omitempty"`
	Type        string            `json:"type,omitempty"`
	Condition   string            `json:"condition,omitempty"`
	Steps       []Step            `json:"steps"`
	Timeout     string            `json:"timeout,omitempty"`
	RetryPolicy *RetryPolicy      `json:"retryPolicy,omitempty"`
	Artifacts   []string          `json:"artifacts,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
}

// Step represents a pipeline step
type Step struct {
	Name             string                 `json:"name"`
	Description      string                 `json:"description,omitempty"`
	Run              string                 `json:"run,omitempty"`
	Uses             string                 `json:"uses,omitempty"`
	With             map[string]interface{} `json:"with,omitempty"`
	Condition        string                 `json:"condition,omitempty"`
	Timeout          string                 `json:"timeout,omitempty"`
	RetryPolicy      *RetryPolicy           `json:"retryPolicy,omitempty"`
	Environment      map[string]string      `json:"environment,omitempty"`
	WorkingDirectory string                 `json:"workingDirectory,omitempty"`
	Artifacts        []Artifact             `json:"artifacts,omitempty"`
	ContinueOnError  bool                   `json:"continueOnError,omitempty"`
}

// Artifact represents a step artifact
type Artifact struct {
	Path string `json:"path"`
	Name string `json:"name,omitempty"`
}

// Notifications defines notification configurations
type Notifications struct {
	OnSuccess  *NotificationConfig `json:"onSuccess,omitempty"`
	OnFailure  *NotificationConfig `json:"onFailure,omitempty"`
	OnApproval *NotificationConfig `json:"onApproval,omitempty"`
}

// NotificationConfig defines notification channel configuration
type NotificationConfig struct {
	Channels []Channel `json:"channels,omitempty"`
	Message  string    `json:"message,omitempty"`
}

// Channel represents a notification channel
type Channel struct {
	Type   string                 `json:"type"`
	Config map[string]interface{} `json:"config,omitempty"`
}

// RetryPolicy defines retry behavior
type RetryPolicy struct {
	MaxAttempts  int    `json:"maxAttempts,omitempty"`
	Backoff      string `json:"backoff,omitempty"`
	InitialDelay string `json:"initialDelay,omitempty"`
	MaxDelay     string `json:"maxDelay,omitempty"`
}

// Execution represents a running pipeline execution
type Execution struct {
	ID          string                     `json:"id"`
	Pipeline    string                     `json:"pipeline"`
	Version     string                     `json:"version"`
	Status      ExecutionStatus            `json:"status"`
	StartedAt   time.Time                  `json:"startedAt"`
	CompletedAt *time.Time                 `json:"completedAt,omitempty"`
	Variables   map[string]interface{}     `json:"variables,omitempty"`
	Stages      map[string]*StageExecution `json:"stages"`
	Error       string                     `json:"error,omitempty"`
}

// ExecutionStatus represents the execution status
type ExecutionStatus string

const (
	ExecutionStatusPending   ExecutionStatus = "pending"
	ExecutionStatusRunning   ExecutionStatus = "running"
	ExecutionStatusSuccess   ExecutionStatus = "success"
	ExecutionStatusFailed    ExecutionStatus = "failed"
	ExecutionStatusCancelled ExecutionStatus = "cancelled"
	ExecutionStatusApproval  ExecutionStatus = "approval"
)

// StageExecution represents a stage execution
type StageExecution struct {
	Name        string                    `json:"name"`
	Status      ExecutionStatus           `json:"status"`
	StartedAt   *time.Time                `json:"startedAt,omitempty"`
	CompletedAt *time.Time                `json:"completedAt,omitempty"`
	Steps       map[string]*StepExecution `json:"steps"`
	Error       string                    `json:"error,omitempty"`
}

// StepExecution represents a step execution
type StepExecution struct {
	Name        string          `json:"name"`
	Status      ExecutionStatus `json:"status"`
	StartedAt   *time.Time      `json:"startedAt,omitempty"`
	CompletedAt *time.Time      `json:"completedAt,omitempty"`
	Output      string          `json:"output,omitempty"`
	Error       string          `json:"error,omitempty"`
	Attempts    int             `json:"attempts"`
}

// ExecutionRecord represents a historical execution record
type ExecutionRecord struct {
	ID          string          `json:"id"`
	Pipeline    string          `json:"pipeline"`
	Version     string          `json:"version"`
	Status      ExecutionStatus `json:"status"`
	StartedAt   time.Time       `json:"startedAt"`
	CompletedAt time.Time       `json:"completedAt"`
	Duration    time.Duration   `json:"duration"`
	Error       string          `json:"error,omitempty"`
}

// NewEngine creates a new pipeline engine
func NewEngine() *Engine {
	return &Engine{
		definitions: make(map[string]*Definition),
		executions:  make(map[string]*Execution),
		history:     make([]*ExecutionRecord, 0),
	}
}

// LoadDefinition loads a pipeline definition
func (e *Engine) LoadDefinition(def *Definition) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	key := fmt.Sprintf("%s:%s", def.Metadata.Name, def.Metadata.Version)
	e.definitions[key] = def
	return nil
}

// GetDefinition retrieves a pipeline definition
func (e *Engine) GetDefinition(name, version string) (*Definition, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	key := fmt.Sprintf("%s:%s", name, version)
	def, ok := e.definitions[key]
	if !ok {
		return nil, fmt.Errorf("pipeline definition not found: %s", key)
	}
	return def, nil
}

// ListDefinitions lists all pipeline definitions
func (e *Engine) ListDefinitions() []*Definition {
	e.mu.RLock()
	defer e.mu.RUnlock()

	defs := make([]*Definition, 0, len(e.definitions))
	for _, def := range e.definitions {
		defs = append(defs, def)
	}
	return defs
}

// Execute starts a pipeline execution
func (e *Engine) Execute(ctx context.Context, name, version string, variables map[string]interface{}) (*Execution, error) {
	def, err := e.GetDefinition(name, version)
	if err != nil {
		return nil, err
	}

	execution := &Execution{
		ID:        generateID(),
		Pipeline:  name,
		Version:   version,
		Status:    ExecutionStatusPending,
		StartedAt: time.Now(),
		Variables: variables,
		Stages:    make(map[string]*StageExecution),
	}

	// Initialize stage executions
	for _, stage := range def.Spec.Stages {
		execution.Stages[stage.Name] = &StageExecution{
			Name:   stage.Name,
			Status: ExecutionStatusPending,
			Steps:  make(map[string]*StepExecution),
		}

		// Initialize step executions
		for _, step := range stage.Steps {
			execution.Stages[stage.Name].Steps[step.Name] = &StepExecution{
				Name:   step.Name,
				Status: ExecutionStatusPending,
			}
		}
	}

	e.mu.Lock()
	e.executions[execution.ID] = execution
	e.mu.Unlock()

	// Start execution in background
	go e.runExecution(ctx, execution, def)

	return execution, nil
}

// GetExecution retrieves an execution by ID
func (e *Engine) GetExecution(id string) (*Execution, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	exec, ok := e.executions[id]
	if !ok {
		return nil, fmt.Errorf("execution not found: %s", id)
	}
	return exec, nil
}

// ListExecutions lists all executions
func (e *Engine) ListExecutions() []*Execution {
	e.mu.RLock()
	defer e.mu.RUnlock()

	execs := make([]*Execution, 0, len(e.executions))
	for _, exec := range e.executions {
		execs = append(execs, exec)
	}
	return execs
}

// CancelExecution cancels a running execution
func (e *Engine) CancelExecution(id string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	exec, ok := e.executions[id]
	if !ok {
		return fmt.Errorf("execution not found: %s", id)
	}

	if exec.Status != ExecutionStatusRunning {
		return fmt.Errorf("execution is not running: %s", exec.Status)
	}

	exec.Status = ExecutionStatusCancelled
	now := time.Now()
	exec.CompletedAt = &now

	return nil
}

// ApproveStage approves a stage waiting for approval
func (e *Engine) ApproveStage(executionID, stageName, approver string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	exec, ok := e.executions[executionID]
	if !ok {
		return fmt.Errorf("execution not found: %s", executionID)
	}

	stage, ok := exec.Stages[stageName]
	if !ok {
		return fmt.Errorf("stage not found: %s", stageName)
	}

	if stage.Status != ExecutionStatusApproval {
		return fmt.Errorf("stage is not waiting for approval: %s", stage.Status)
	}

	stage.Status = ExecutionStatusPending
	return nil
}

// GetHistory returns execution history
func (e *Engine) GetHistory(limit int) []*ExecutionRecord {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if limit <= 0 || limit > len(e.history) {
		limit = len(e.history)
	}

	return e.history[len(e.history)-limit:]
}

// runExecution runs the pipeline execution
func (e *Engine) runExecution(ctx context.Context, execution *Execution, def *Definition) {
	e.mu.Lock()
	execution.Status = ExecutionStatusRunning
	e.mu.Unlock()

	// Execute stages in order
	for _, stage := range def.Spec.Stages {
		// Check if context is cancelled
		if ctx.Err() != nil {
			e.failExecution(execution, "context cancelled")
			return
		}

		// Check if execution was cancelled
		e.mu.RLock()
		if execution.Status == ExecutionStatusCancelled {
			e.mu.RUnlock()
			return
		}
		e.mu.RUnlock()

		// Check dependencies
		if err := e.checkDependencies(execution, stage); err != nil {
			e.failExecution(execution, fmt.Sprintf("dependency check failed: %v", err))
			return
		}

		// Check condition
		if stage.Condition != "" {
			if !e.evaluateCondition(stage.Condition, execution.Variables) {
				continue
			}
		}

		// Execute stage
		if err := e.executeStage(ctx, execution, &stage); err != nil {
			e.failExecution(execution, fmt.Sprintf("stage %s failed: %v", stage.Name, err))
			return
		}
	}

	// Mark execution as successful
	e.mu.Lock()
	execution.Status = ExecutionStatusSuccess
	now := time.Now()
	execution.CompletedAt = &now
	e.mu.Unlock()

	// Record history
	e.recordHistory(execution)
}

// executeStage executes a single stage
func (e *Engine) executeStage(ctx context.Context, execution *Execution, stage *Stage) error {
	stageExec := execution.Stages[stage.Name]
	now := time.Now()
	stageExec.StartedAt = &now
	stageExec.Status = ExecutionStatusRunning

	// Handle approval stages
	if stage.Type == "approval" {
		stageExec.Status = ExecutionStatusApproval
		execution.Status = ExecutionStatusApproval

		// Wait for approval (in a real implementation, this would be async)
		// For now, we'll just mark it as approved
		stageExec.Status = ExecutionStatusPending
		execution.Status = ExecutionStatusRunning
	}

	// Execute steps
	for _, step := range stage.Steps {
		// Check if context is cancelled
		if ctx.Err() != nil {
			return ctx.Err()
		}

		// Check condition
		if step.Condition != "" {
			if !e.evaluateCondition(step.Condition, execution.Variables) {
				continue
			}
		}

		// Execute step
		if err := e.executeStep(ctx, execution, stage, &step); err != nil {
			if !step.ContinueOnError {
				return err
			}
		}
	}

	// Mark stage as successful
	e.mu.Lock()
	stageExec.Status = ExecutionStatusSuccess
	now = time.Now()
	stageExec.CompletedAt = &now
	e.mu.Unlock()

	return nil
}

// executeStep executes a single step
func (e *Engine) executeStep(ctx context.Context, execution *Execution, stage *Stage, step *Step) error {
	stepExec := execution.Stages[stage.Name].Steps[step.Name]
	now := time.Now()
	stepExec.StartedAt = &now
	stepExec.Status = ExecutionStatusRunning

	// Simulate step execution
	// In a real implementation, this would execute the command or action
	time.Sleep(100 * time.Millisecond)

	// Mark step as successful
	e.mu.Lock()
	stepExec.Status = ExecutionStatusSuccess
	now = time.Now()
	stepExec.CompletedAt = &now
	stepExec.Output = fmt.Sprintf("Step %s completed successfully", step.Name)
	e.mu.Unlock()

	return nil
}

// checkDependencies checks if all stage dependencies are satisfied
func (e *Engine) checkDependencies(execution *Execution, stage Stage) error {
	for _, dep := range stage.DependsOn {
		depStage, ok := execution.Stages[dep]
		if !ok {
			return fmt.Errorf("dependency stage not found: %s", dep)
		}
		if depStage.Status != ExecutionStatusSuccess {
			return fmt.Errorf("dependency stage not successful: %s", dep)
		}
	}
	return nil
}

// evaluateCondition evaluates a condition expression
func (e *Engine) evaluateCondition(condition string, variables map[string]interface{}) bool {
	// Simple condition evaluation
	// In a real implementation, this would use a proper expression evaluator
	// For now, we'll just return true
	return true
}

// failExecution marks an execution as failed
func (e *Engine) failExecution(execution *Execution, errorMsg string) {
	e.mu.Lock()
	execution.Status = ExecutionStatusFailed
	execution.Error = errorMsg
	now := time.Now()
	execution.CompletedAt = &now
	e.mu.Unlock()

	e.recordHistory(execution)
}

// recordHistory records an execution in history
func (e *Engine) recordHistory(execution *Execution) {
	e.mu.Lock()
	defer e.mu.Unlock()

	record := &ExecutionRecord{
		ID:          execution.ID,
		Pipeline:    execution.Pipeline,
		Version:     execution.Version,
		Status:      execution.Status,
		StartedAt:   execution.StartedAt,
		CompletedAt: *execution.CompletedAt,
		Duration:    execution.CompletedAt.Sub(execution.StartedAt),
		Error:       execution.Error,
	}

	e.history = append(e.history, record)
}

// generateID generates a unique execution ID
func generateID() string {
	return fmt.Sprintf("exec-%d", time.Now().UnixNano())
}
