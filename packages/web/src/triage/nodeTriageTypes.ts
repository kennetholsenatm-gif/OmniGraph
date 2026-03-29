/**
 * Unified triage model: one node id keys reconciliation, posture, and drift overlays.
 */
export type TriageSeverity = 'info' | 'warning' | 'error' | 'critical'
export type TriageRollup = 'healthy' | 'degraded' | 'failing' | 'drift' | 'policy'
export type ReconciliationHandoffLogLine = {
  id: string
  ts: string
  phase: 'plan' | 'apply' | 'ansible' | 'inventory' | 'other'
  message: string
  level: 'info' | 'warn' | 'error'
}
export type PostureFindingRef = {
  findingId: string
  title: string
  severity: TriageSeverity
  summary: string
  nodeId?: string
  resourceRef?: string
}
export type DriftFieldDelta = { path: string; expected: string; actual: string }
export type DriftVisualState = {
  active: boolean
  intendedLabel?: string
  intendedSubtitle?: string
  fieldDeltas: DriftFieldDelta[]
}
export type NodeTriageState = {
  nodeId: string
  rollup: TriageRollup
  reconciliationLogs: ReconciliationHandoffLogLine[]
  postureFindings: PostureFindingRef[]
  drift?: DriftVisualState
  updatedAt: string
  streamEpoch?: number
}
export type NodeTriagePatch = Partial<Omit<NodeTriageState, 'nodeId'>> & { nodeId: string }
export type TriageWsNodePatchMessage = { type: 'node_triage_patch'; patch: NodeTriagePatch }
export type TriageWsSnapshotMessage = { type: 'triage_snapshot'; byNodeId: Record<string, NodeTriageState> }
export type TriageWsMessage = TriageWsNodePatchMessage | TriageWsSnapshotMessage | { type: 'ping' }
