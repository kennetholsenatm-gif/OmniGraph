package serve

import (
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"
)

// MetricsCollector collects and exposes metrics in Prometheus format
type MetricsCollector struct {
	mu sync.RWMutex

	// Pipeline metrics
	pipelinesTotal      int64
	pipelinesRunning    int64
	pipelinesSuccess    int64
	pipelinesFailed     int64
	pipelinesCancelled  int64

	// Execution metrics
	executionsTotal     int64
	executionsDuration  time.Duration

	// API metrics
	apiRequestsTotal    int64
	apiRequestsDuration time.Duration
	apiErrorsTotal      int64

	// Resource metrics
	goroutinesCount     int
	memoryAllocated     uint64
	memoryTotalAlloc    uint64
	memorySys           uint64

	// State metrics
	stateFilesProcessed int64
	stateDriftDetected  int64
	stateLocksActive    int64

	// Security metrics
	securityScansTotal  int64
	securityFindings    int64
	secretRedactions    int64

	// Integration metrics
	netboxSyncsTotal    int64
	vaultFetchesTotal   int64
	notificationsSent   int64

	// Start time for uptime calculation
	startTime time.Time
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		startTime: time.Now(),
	}
}

// RecordPipelineMetrics records pipeline execution metrics
func (m *MetricsCollector) RecordPipelineMetrics(status string, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.pipelinesTotal++
	m.executionsTotal++
	m.executionsDuration += duration

	switch status {
	case "success":
		m.pipelinesSuccess++
	case "failed":
		m.pipelinesFailed++
	case "cancelled":
		m.pipelinesCancelled++
	case "running":
		m.pipelinesRunning++
	}
}

// RecordAPIRequest records an API request
func (m *MetricsCollector) RecordAPIRequest(duration time.Duration, isError bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.apiRequestsTotal++
	m.apiRequestsDuration += duration

	if isError {
		m.apiErrorsTotal++
	}
}

// RecordStateMetrics records state management metrics
func (m *MetricsCollector) RecordStateMetrics(filesProcessed, driftDetected, locksActive int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.stateFilesProcessed += filesProcessed
	m.stateDriftDetected += driftDetected
	m.stateLocksActive = locksActive
}

// RecordSecurityMetrics records security scan metrics
func (m *MetricsCollector) RecordSecurityMetrics(scans, findings, redactions int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.securityScansTotal += scans
	m.securityFindings += findings
	m.secretRedactions += redactions
}

// RecordIntegrationMetrics records integration metrics
func (m *MetricsCollector) RecordIntegrationMetrics(netboxSyncs, vaultFetches, notifications int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.netboxSyncsTotal += netboxSyncs
	m.vaultFetchesTotal += vaultFetches
	m.notificationsSent += notifications
}

// UpdateResourceMetrics updates resource usage metrics
func (m *MetricsCollector) UpdateResourceMetrics() {
	m.mu.Lock()
	defer m.mu.Unlock()

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	m.goroutinesCount = runtime.NumGoroutine()
	m.memoryAllocated = memStats.Alloc
	m.memoryTotalAlloc = memStats.TotalAlloc
	m.memorySys = memStats.Sys
}

// Handler returns an HTTP handler for the /metrics endpoint
func (m *MetricsCollector) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.UpdateResourceMetrics()

		m.mu.RLock()
		defer m.mu.RUnlock()

		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")

		// Write Prometheus format metrics
		fmt.Fprintf(w, "# HELP omnigraph_info Information about the OmniGraph instance\n")
		fmt.Fprintf(w, "# TYPE omnigraph_info gauge\n")
		fmt.Fprintf(w, "omnigraph_info{version=\"%s\"} 1\n", "1.0.0")

		fmt.Fprintf(w, "# HELP omnigraph_uptime_seconds Time since the instance started\n")
		fmt.Fprintf(w, "# TYPE omnigraph_uptime_seconds gauge\n")
		fmt.Fprintf(w, "omnigraph_uptime_seconds %f\n", time.Since(m.startTime).Seconds())

		// Pipeline metrics
		fmt.Fprintf(w, "# HELP omnigraph_pipelines_total Total number of pipeline executions\n")
		fmt.Fprintf(w, "# TYPE omnigraph_pipelines_total counter\n")
		fmt.Fprintf(w, "omnigraph_pipelines_total %d\n", m.pipelinesTotal)

		fmt.Fprintf(w, "# HELP omnigraph_pipelines_running Currently running pipelines\n")
		fmt.Fprintf(w, "# TYPE omnigraph_pipelines_running gauge\n")
		fmt.Fprintf(w, "omnigraph_pipelines_running %d\n", m.pipelinesRunning)

		fmt.Fprintf(w, "# HELP omnigraph_pipelines_success Total successful pipeline executions\n")
		fmt.Fprintf(w, "# TYPE omnigraph_pipelines_success counter\n")
		fmt.Fprintf(w, "omnigraph_pipelines_success %d\n", m.pipelinesSuccess)

		fmt.Fprintf(w, "# HELP omnigraph_pipelines_failed Total failed pipeline executions\n")
		fmt.Fprintf(w, "# TYPE omnigraph_pipelines_failed counter\n")
		fmt.Fprintf(w, "omnigraph_pipelines_failed %d\n", m.pipelinesFailed)

		fmt.Fprintf(w, "# HELP omnigraph_pipelines_cancelled Total cancelled pipeline executions\n")
		fmt.Fprintf(w, "# TYPE omnigraph_pipelines_cancelled counter\n")
		fmt.Fprintf(w, "omnigraph_pipelines_cancelled %d\n", m.pipelinesCancelled)

		// Execution metrics
		fmt.Fprintf(w, "# HELP omnigraph_executions_total Total number of executions\n")
		fmt.Fprintf(w, "# TYPE omnigraph_executions_total counter\n")
		fmt.Fprintf(w, "omnigraph_executions_total %d\n", m.executionsTotal)

		fmt.Fprintf(w, "# HELP omnigraph_executions_duration_seconds_total Total execution duration\n")
		fmt.Fprintf(w, "# TYPE omnigraph_executions_duration_seconds_total counter\n")
		fmt.Fprintf(w, "omnigraph_executions_duration_seconds_total %f\n", m.executionsDuration.Seconds())

		// API metrics
		fmt.Fprintf(w, "# HELP omnigraph_api_requests_total Total API requests\n")
		fmt.Fprintf(w, "# TYPE omnigraph_api_requests_total counter\n")
		fmt.Fprintf(w, "omnigraph_api_requests_total %d\n", m.apiRequestsTotal)

		fmt.Fprintf(w, "# HELP omnigraph_api_requests_duration_seconds_total Total API request duration\n")
		fmt.Fprintf(w, "# TYPE omnigraph_api_requests_duration_seconds_total counter\n")
		fmt.Fprintf(w, "omnigraph_api_requests_duration_seconds_total %f\n", m.apiRequestsDuration.Seconds())

		fmt.Fprintf(w, "# HELP omnigraph_api_errors_total Total API errors\n")
		fmt.Fprintf(w, "# TYPE omnigraph_api_errors_total counter\n")
		fmt.Fprintf(w, "omnigraph_api_errors_total %d\n", m.apiErrorsTotal)

		// Resource metrics
		fmt.Fprintf(w, "# HELP omnigraph_goroutines Current number of goroutines\n")
		fmt.Fprintf(w, "# TYPE omnigraph_goroutines gauge\n")
		fmt.Fprintf(w, "omnigraph_goroutines %d\n", m.goroutinesCount)

		fmt.Fprintf(w, "# HELP omnigraph_memory_allocated_bytes Currently allocated memory\n")
		fmt.Fprintf(w, "# TYPE omnigraph_memory_allocated_bytes gauge\n")
		fmt.Fprintf(w, "omnigraph_memory_allocated_bytes %d\n", m.memoryAllocated)

		fmt.Fprintf(w, "# HELP omnigraph_memory_total_allocated_bytes Total allocated memory\n")
		fmt.Fprintf(w, "# TYPE omnigraph_memory_total_allocated_bytes counter\n")
		fmt.Fprintf(w, "omnigraph_memory_total_allocated_bytes %d\n", m.memoryTotalAlloc)

		fmt.Fprintf(w, "# HELP omnigraph_memory_system_bytes System memory\n")
		fmt.Fprintf(w, "# TYPE omnigraph_memory_system_bytes gauge\n")
		fmt.Fprintf(w, "omnigraph_memory_system_bytes %d\n", m.memorySys)

		// State metrics
		fmt.Fprintf(w, "# HELP omnigraph_state_files_processed_total Total state files processed\n")
		fmt.Fprintf(w, "# TYPE omnigraph_state_files_processed_total counter\n")
		fmt.Fprintf(w, "omnigraph_state_files_processed_total %d\n", m.stateFilesProcessed)

		fmt.Fprintf(w, "# HELP omnigraph_state_drift_detected_total Total drift detections\n")
		fmt.Fprintf(w, "# TYPE omnigraph_state_drift_detected_total counter\n")
		fmt.Fprintf(w, "omnigraph_state_drift_detected_total %d\n", m.stateDriftDetected)

		fmt.Fprintf(w, "# HELP omnigraph_state_locks_active Currently active state locks\n")
		fmt.Fprintf(w, "# TYPE omnigraph_state_locks_active gauge\n")
		fmt.Fprintf(w, "omnigraph_state_locks_active %d\n", m.stateLocksActive)

		// Security metrics
		fmt.Fprintf(w, "# HELP omnigraph_security_scans_total Total security scans\n")
		fmt.Fprintf(w, "# TYPE omnigraph_security_scans_total counter\n")
		fmt.Fprintf(w, "omnigraph_security_scans_total %d\n", m.securityScansTotal)

		fmt.Fprintf(w, "# HELP omnigraph_security_findings_total Total security findings\n")
		fmt.Fprintf(w, "# TYPE omnigraph_security_findings_total counter\n")
		fmt.Fprintf(w, "omnigraph_security_findings_total %d\n", m.securityFindings)

		fmt.Fprintf(w, "# HELP omnigraph_secret_redactions_total Total secret redactions\n")
		fmt.Fprintf(w, "# TYPE omnigraph_secret_redactions_total counter\n")
		fmt.Fprintf(w, "omnigraph_secret_redactions_total %d\n", m.secretRedactions)

		// Integration metrics
		fmt.Fprintf(w, "# HELP omnigraph_netbox_syncs_total Total NetBox syncs\n")
		fmt.Fprintf(w, "# TYPE omnigraph_netbox_syncs_total counter\n")
		fmt.Fprintf(w, "omnigraph_netbox_syncs_total %d\n", m.netboxSyncsTotal)

		fmt.Fprintf(w, "# HELP omnigraph_vault_fetches_total Total Vault secret fetches\n")
		fmt.Fprintf(w, "# TYPE omnigraph_vault_fetches_total counter\n")
		fmt.Fprintf(w, "omnigraph_vault_fetches_total %d\n", m.vaultFetchesTotal)

		fmt.Fprintf(w, "# HELP omnigraph_notifications_sent_total Total notifications sent\n")
		fmt.Fprintf(w, "# TYPE omnigraph_notifications_sent_total counter\n")
		fmt.Fprintf(w, "omnigraph_notifications_sent_total %d\n", m.notificationsSent)
	})
}

// Global metrics collector
var globalMetricsCollector *MetricsCollector

// GetMetricsCollector returns the global metrics collector
func GetMetricsCollector() *MetricsCollector {
	if globalMetricsCollector == nil {
		globalMetricsCollector = NewMetricsCollector()
	}
	return globalMetricsCollector
}