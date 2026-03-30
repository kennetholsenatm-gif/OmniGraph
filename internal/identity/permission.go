package identity

// Permission strings are stable API identifiers for RBAC. Serve, IR, and lock paths
// should reuse these constants as they gain IdP-backed authorization.
const (
	PermServeHealth         = "serve:health"
	PermServeInventoryRead  = "serve:inventory:read"
	PermServeSecurityScan   = "serve:security:scan"
	PermServeHostOpsRead    = "serve:host-ops:read"
	PermServeHostOpsWrite   = "serve:host-ops:write"
	PermServeAuditRead      = "serve:audit:read"
	PermServeIngestLocal    = "serve:ingest:local"
	PermServeSyncWS         = "serve:sync:ws"
	PermServeWorkspaceDrift = "serve:workspace:drift"
	PermServeIntegrationRun = "serve:integration:run"
	PermIRValidate          = "ir:validate"
	PermIREmit              = "ir:emit"
	PermLockAcquire         = "lock:acquire"
	PermLockRelease         = "lock:release"
	PermCIReport            = "ci:report"
)
