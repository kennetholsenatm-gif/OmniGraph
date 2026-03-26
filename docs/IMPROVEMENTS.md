# OmniGraph CI/OT Control Plane Improvements

This document describes the improvements made to transform OmniGraph into a complete CI/OT (Continuous Integration/Operations Technology) Control/Management plane.

## Overview

OmniGraph has been enhanced with enterprise-grade features for managing infrastructure lifecycle, pipeline orchestration, observability, and state management. These improvements make it suitable for production use as a comprehensive DevSecOps control plane.

## New Features

### 1. Pipeline Management System

A complete pipeline orchestration engine with support for multi-stage workflows, approval gates, and conditional execution.

#### Pipeline Definition Schema

**Location:** `schemas/pipeline.v1.schema.json`

The pipeline schema defines:
- **Stages**: Logical groupings of steps with dependencies
- **Steps**: Individual tasks to execute (commands, actions)
- **Triggers**: Webhook, schedule, manual, or event-based
- **Variables**: Configuration parameters with types and defaults
- **Notifications**: Success/failure alerts via Slack, Email, PagerDuty
- **Retry Policies**: Exponential backoff with configurable attempts
- **Timeouts**: Per-stage and per-step timeout configuration

#### Pipeline Engine

**Location:** `internal/pipeline/engine.go`

The pipeline engine provides:
- **Definition Management**: Load, list, and version pipeline definitions
- **Execution**: Start, monitor, and cancel pipeline runs
- **Stage Dependencies**: Automatic dependency resolution
- **Approval Gates**: Manual intervention points for production deployments
- **Conditional Execution**: Run stages based on variable values
- **History Tracking**: Complete audit trail of all executions

#### CLI Commands

**Location:** `internal/cli/pipeline.go`

New pipeline commands:
```bash
# Load a pipeline definition
omnigraph pipeline define my-pipeline.yaml

# Execute a pipeline
omnigraph pipeline run deploy-infrastructure --var environment=production

# List available pipelines
omnigraph pipeline list

# Check execution status
omnigraph pipeline status exec-1234567890

# Approve a stage waiting for approval
omnigraph pipeline approve exec-1234567890 approve-stage --approver john.doe

# Cancel a running execution
omnigraph pipeline cancel exec-1234567890

# View execution history
omnigraph pipeline history --limit 10
```

#### Sample Pipeline

**Location:** `testdata/sample.pipeline.json`

A complete example pipeline with:
- 6 stages: validate → plan → approve → apply → configure → notify
- Webhook and schedule triggers
- Production approval gate
- Artifact collection
- Slack and email notifications
- Retry policies

### 2. Metrics and Observability

Prometheus-compatible metrics endpoint for monitoring and alerting.

#### Metrics Endpoint

**Location:** `internal/serve/metrics.go`

Available at: `GET /metrics` (when `--enable-metrics` flag is used)

#### Available Metrics

**Pipeline Metrics:**
- `omnigraph_pipelines_total` - Total pipeline executions
- `omnigraph_pipelines_running` - Currently running pipelines
- `omnigraph_pipelines_success` - Successful executions
- `omnigraph_pipelines_failed` - Failed executions
- `omnigraph_pipelines_cancelled` - Cancelled executions

**Execution Metrics:**
- `omnigraph_executions_total` - Total step executions
- `omnigraph_executions_duration_seconds_total` - Total execution time

**API Metrics:**
- `omnigraph_api_requests_total` - Total API requests
- `omnigraph_api_requests_duration_seconds_total` - Total API request time
- `omnigraph_api_errors_total` - Total API errors

**Resource Metrics:**
- `omnigraph_goroutines` - Current goroutine count
- `omnigraph_memory_allocated_bytes` - Currently allocated memory
- `omnigraph_memory_total_allocated_bytes` - Total allocated memory
- `omnigraph_memory_system_bytes` - System memory usage

**State Metrics:**
- `omnigraph_state_files_processed_total` - Total state files processed
- `omnigraph_state_drift_detected_total` - Total drift detections
- `omnigraph_state_locks_active` - Currently active state locks

**Security Metrics:**
- `omnigraph_security_scans_total` - Total security scans
- `omnigraph_security_findings_total` - Total security findings
- `omnigraph_secret_redactions_total` - Total secret redactions

**Integration Metrics:**
- `omnigraph_netbox_syncs_total` - Total NetBox syncs
- `omnigraph_vault_fetches_total` - Total Vault secret fetches
- `omnigraph_notifications_sent_total` - Total notifications sent

**System Metrics:**
- `omnigraph_info` - Instance information
- `omnigraph_uptime_seconds` - Instance uptime

#### Usage

```bash
# Start server with metrics enabled
omnigraph serve --enable-metrics

# Query metrics
curl http://127.0.0.1:38671/metrics
```

#### Integration with Monitoring

Metrics can be scraped by:
- **Prometheus**: Add to `prometheus.yml`:
  ```yaml
  scrape_configs:
    - job_name: 'omnigraph'
      static_configs:
        - targets: ['localhost:38671']
  ```

- **Grafana**: Create dashboards using Prometheus datasource

### 3. Enhanced State Management

Advanced state management with drift detection, locking, and versioning.

**Location:** `internal/state/manager.go`

#### Features

**State Locking:**
- Prevent concurrent modifications to state files
- Configurable TTL (time-to-live)
- Automatic cleanup of expired locks
- Lock file metadata

```go
// Acquire a lock
lock, err := manager.AcquireLock(ctx, "terraform.tfstate", "user@example.com", "apply", 10*time.Minute)

// Release the lock
err := manager.ReleaseLock("terraform.tfstate", lock.ID)

// List all active locks
locks := manager.ListLocks()
```

**Drift Detection:**
- Compare current state with expected state
- Hash-based change detection
- Detailed change analysis
- Cached results for performance

```go
// Detect drift
result, err := manager.DetectDrift(ctx, "terraform.tfstate", expectedHash)

if result.HasDrift {
    fmt.Printf("Drift detected! Changes: %v\n", result.Changes)
}
```

**State Versioning:**
- Create versioned snapshots
- Track author and commit message
- Tag versions (e.g., "production", "backup")
- Version history queries

```go
// Create a version
version, err := manager.CreateVersion(ctx, "terraform.tfstate", "user@example.com", "Pre-deployment backup", []string{"backup"})

// Get version history
versions, err := manager.GetVersionHistory("terraform.tfstate", 10)
```

**State Watching:**
- Monitor state files for changes
- Configurable polling interval
- Callback on changes
- Automatic hash calculation

```go
// Start watching
err := manager.Watch("terraform.tfstate", 30*time.Second, func(path string, oldHash, newHash string) {
    fmt.Printf("State changed: %s -> %s\n", oldHash, newHash)
})

// Stop watching
err := manager.Unwatch("terraform.tfstate")
```

### 4. Enhanced Serve Command

The serve command now supports metrics endpoint.

**Location:** `internal/cli/serve.go`

#### New Flags

```bash
--enable-metrics    # Enable Prometheus metrics endpoint at /metrics
```

#### Example

```bash
# Start server with all features
omnigraph serve \
  --web-dist web/dist \
  --enable-security-scan \
  --enable-host-ops \
  --enable-inventory-api \
  --enable-metrics \
  --auth-token my-secret-token
```

## Architecture Improvements

### Pipeline Engine Architecture

```
┌─────────────────────────────────────────────┐
│           Pipeline Definition               │
│  (JSON/YAML with stages, steps, triggers)   │
└─────────────┬───────────────────────────────┘
              │
              ▼
┌─────────────────────────────────────────────┐
│         Pipeline Engine                      │
│  - Definition management                     │
│  - Execution orchestration                   │
│  - Stage dependency resolution               │
│  - Approval gate handling                    │
│  - Variable substitution                     │
└─────────────┬───────────────────────────────┘
              │
              ▼
┌─────────────────────────────────────────────┐
│         Execution Layer                      │
│  - ExecRunner / ContainerRunner              │
│  - Secret injection                          │
│  - Log capture & redaction                  │
│  - Timeout enforcement                       │
│  - Retry with backoff                        │
└─────────────────────────────────────────────┘
```

### State Management Architecture

```
┌─────────────────────────────────────────────┐
│           State Store                        │
│  - Local filesystem                          │
│  - Version control                           │
│  - Distributed locking                       │
│  - Hash-based change detection               │
└─────────────┬───────────────────────────────┘
              │
              ▼
┌─────────────────────────────────────────────┐
│         Drift Detector                       │
│  - Continuous monitoring                     │
│  - Hash comparison                           │
│  - Change analysis                           │
│  - Alert generation                          │
└─────────────┬───────────────────────────────┘
              │
              ▼
┌─────────────────────────────────────────────┐
│         State API                            │
│  - REST API for state operations             │
│  - Lock management                           │
│  - Version history                           │
│  - Statistics and monitoring                 │
└─────────────────────────────────────────────┘
```

### Observability Architecture

```
┌─────────────────────────────────────────────┐
│           Metrics Collector                  │
│  - Prometheus format                         │
│  - Custom metrics                            │
│  - Resource usage                            │
│  - Pipeline performance                      │
└─────────────┬───────────────────────────────┘
              │
              ▼
┌─────────────────────────────────────────────┐
│         HTTP Handler                         │
│  - GET /metrics endpoint                     │
│  - Content negotiation                       │
│  - Caching headers                           │
└─────────────┬───────────────────────────────┘
              │
              ▼
┌─────────────────────────────────────────────┐
│         Monitoring Integration               │
│  - Prometheus scraping                       │
│  - Grafana dashboards                        │
│  - Alertmanager rules                        │
│  - PagerDuty integration                     │
└─────────────────────────────────────────────┘
```

## Usage Examples

### Pipeline Execution

```bash
# Define a pipeline
omnigraph pipeline define testdata/sample.pipeline.json

# Run the pipeline
omnigraph pipeline run deploy-infrastructure \
  --var environment=production \
  --var region=us-west-2 \
  --version 1.0.0

# Monitor execution
omnigraph pipeline status exec-1234567890

# Approve production deployment
omnigraph pipeline approve exec-1234567890 approve-stage --approver john.doe
```

### State Management

```bash
# Start server with state management
omnigraph serve --enable-metrics

# The state manager will automatically:
# - Lock state files during operations
# - Detect drift between expected and actual state
# - Create versions before changes
# - Monitor for external modifications
```

### Metrics Monitoring

```bash
# Start server with metrics
omnigraph serve --enable-metrics

# Query metrics in another terminal
curl http://127.0.0.1:38671/metrics

# Example output:
# omnigraph_pipelines_total 42
# omnigraph_pipelines_success 38
# omnigraph_pipelines_failed 4
# omnigraph_api_requests_total 1234
# omnigraph_goroutines 15
# omnigraph_memory_allocated_bytes 12345678
```

## Configuration

### Pipeline Configuration

Pipeline definitions support:
- **Triggers**: webhook, schedule, manual, event
- **Variables**: string, number, boolean, secret types
- **Stages**: standard, approval, parallel types
- **Steps**: run commands or use reusable actions
- **Notifications**: slack, teams, email, webhook, pagerduty
- **Retry**: fixed or exponential backoff

### Metrics Configuration

Enable metrics with:
```bash
omnigraph serve --enable-metrics
```

Metrics are available at `/metrics` endpoint.

### State Management Configuration

State management is automatic when using OmniGraph commands. The state manager:
- Creates locks before state modifications
- Detects drift continuously
- Versions state before changes
- Monitors for external changes

## Best Practices

### Pipeline Design

1. **Stage Granularity**: Keep stages focused and atomic
2. **Dependencies**: Use `dependsOn` for stage ordering
3. **Timeouts**: Set appropriate timeouts for long-running operations
4. **Retry Policies**: Configure retries for transient failures
5. **Notifications**: Set up success/failure alerts
6. **Approval Gates**: Use for production deployments
7. **Variables**: Use for environment-specific configuration
8. **Artifacts**: Collect important outputs for later stages

### State Management

1. **Locking**: Always acquire locks before state modifications
2. **Versioning**: Create versions before major changes
3. **Drift Detection**: Enable continuous monitoring
4. **Backup**: Regular state backups with tags
5. **Cleanup**: Regular cleanup of expired locks and old versions

### Observability

1. **Metrics**: Enable metrics in production
2. **Dashboards**: Create Grafana dashboards for key metrics
3. **Alerts**: Set up alerts for failures and anomalies
4. **Logging**: Use structured logging for correlation
5. **Tracing**: Enable distributed tracing for complex operations

## Future Enhancements

### Planned Features

1. **Workflow Engine**:
   - Conditional branching (if/else)
   - Loops and iteration
   - Error handling and compensation
   - Sub-workflow composition
   - Event-driven triggers

2. **Multi-cloud Support**:
   - AWS, Azure, GCP native integrations
   - Cloud resource inventory
   - Cost optimization recommendations
   - Cross-cloud orchestration

3. **Advanced Observability**:
   - Distributed tracing (OpenTelemetry)
   - Centralized logging
   - Performance baselines
   - Anomaly detection

4. **Access Control**:
   - Fine-grained resource permissions
   - Approval workflow engine
   - OPA/Rego policy integration
   - Audit log search and export

5. **Integration Ecosystem**:
   - CI/CD platform integration (Jenkins, GitLab CI, GitHub Actions)
   - Artifact repository integration
   - Ticketing system integration
   - Cloud provider native integrations

6. **Web UI Enhancements**:
   - Real-time pipeline execution dashboard
   - Collaborative editing
   - Visual pipeline builder
   - Mobile-responsive design

## Troubleshooting

### Pipeline Issues

**Problem**: Pipeline fails to start
- **Solution**: Check pipeline definition syntax with `omnigraph validate`

**Problem**: Stage dependencies not resolving
- **Solution**: Verify `dependsOn` references match stage names

**Problem**: Approval stage not progressing
- **Solution**: Use `omnigraph pipeline approve` to approve the stage

### Metrics Issues

**Problem**: Metrics endpoint not available
- **Solution**: Ensure `--enable-metrics` flag is used when starting server

**Problem**: Metrics not updating
- **Solution**: Check that operations are being performed and metrics are being recorded

### State Management Issues

**Problem**: Lock acquisition fails
- **Solution**: Check for expired locks or use `ListLocks()` to see active locks

**Problem**: Drift detection not working
- **Solution**: Ensure state file exists and is readable

**Problem**: Version history missing
- **Solution**: Versions are created automatically before state changes

## Contributing

To contribute to these improvements:

1. **Pipeline Engine**: Extend `internal/pipeline/engine.go`
2. **Metrics**: Add new metrics in `internal/serve/metrics.go`
3. **State Management**: Enhance `internal/state/manager.go`
4. **CLI Commands**: Add new commands in `internal/cli/pipeline.go`
5. **Documentation**: Update this file and related docs

## References

- [Pipeline Schema](../schemas/pipeline.v1.schema.json)
- [Pipeline Engine](../internal/pipeline/engine.go)
- [Metrics Handler](../internal/serve/metrics.go)
- [State Manager](../internal/state/manager.go)
- [Sample Pipeline](../testdata/sample.pipeline.json)
- [Architecture](./architecture.md)
- [Execution Matrix](./execution-matrix.md)
- [Integrations](./integrations.md)