/** Prefer JSON `{ code, message }` bodies from serve when present (workspace summary, ingest). */
function errorMessageFromResponseBody(text: string, fallback: string): string {
  const t = text.trim()
  if (!t) {
    return fallback
  }
  try {
    const j = JSON.parse(t) as { code?: string; message?: string }
    if (typeof j.message === 'string' && j.message.trim()) {
      return j.code ? `${j.message} (${j.code})` : j.message
    }
  } catch {
    // plain text or HTML
  }
  return t
}

/** Base URL for the local OmniGraph workspace server API. Empty = same origin (UI served with `--web-dist`). */
export function omnigraphApiBase(): string {
  const v = import.meta.env.VITE_OMNIGRAPH_API
  if (typeof v === 'string' && v.trim()) {
    return v.replace(/\/$/, '')
  }
  return ''
}

export type ApiDiscoveredFile = { path: string; kind: string }

export type ApiDiscoverResult = {
  root: string
  files: ApiDiscoveredFile[]
}

export type ApiStateHostRow = {
  name: string
  ansibleHost: string
  origin: string
}

export type WorkspaceSummary = {
  root: string
  discover: ApiDiscoverResult
  stateInventory: ApiStateHostRow[]
  stateErrors?: string[]
  omnigraphIni: string
}

/** Requires `--enable-security-scan`, `Authorization: Bearer`, and same-origin or `VITE_OMNIGRAPH_API`. */
export async function fetchLocalSecurityScan(
  bearerToken: string,
  body: { mode?: string; profile?: string; tactic?: string; technique?: string; module?: string },
): Promise<unknown> {
  const base = omnigraphApiBase()
  const r = await fetch(`${base}/api/v1/security/scan`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${bearerToken.trim()}`,
    },
    body: JSON.stringify({ mode: 'local', ...body }),
  })
  if (!r.ok) {
    const t = await r.text()
    throw new Error(t || r.statusText)
  }
  return r.json() as Promise<unknown>
}

export async function fetchWorkspaceSummary(path: string): Promise<WorkspaceSummary> {
  const base = omnigraphApiBase()
  const r = await fetch(`${base}/api/v1/workspace/summary`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ path: path.trim() || '.' }),
  })
  if (!r.ok) {
    const t = await r.text()
    throw new Error(errorMessageFromResponseBody(t, r.statusText))
  }
  return r.json() as Promise<WorkspaceSummary>
}

export type IngestFileItem = {
  name: string
  contentType: string
  encoding: string
  data: string
  clientPathHint?: string
  lastModified?: string
}

import type { IngestOmniState } from './inventorySources'

export type IngestLocalError = { path?: string; code: string; message: string }

export type IngestLocalResponse = {
  state: IngestOmniState
  errors?: IngestLocalError[]
}

export type BomEntityClass = 'software_component' | 'hardware_asset' | 'service_endpoint'
export type BomRelationType = 'depends_on' | 'runs_on' | 'hosts' | 'connects_to'
export type BomConfidence = 'authoritative' | 'high' | 'medium' | 'low' | 'unknown'
export type BomRelationDriftCategory = 'missing_dependency' | 'stale_dependency' | 'confidence_drop'

export type BomEntity = {
  id: string
  class: BomEntityClass
  name: string
  version?: string
  provenance?: string
  confidence?: BomConfidence
  observedAt?: string
  attributes?: Record<string, unknown>
}

export type BomRelation = {
  from: string
  to: string
  type: BomRelationType
  confidence?: BomConfidence
  observedAt?: string
  attributes?: Record<string, unknown>
}

export type BomDocument = {
  apiVersion: 'omnigraph/bom/v1'
  kind: 'BOM'
  metadata: { generatedAt: string; source: string; correlationId?: string }
  spec: { entities: BomEntity[]; relations: BomRelation[]; errors?: { code: string; message: string; path?: string }[] }
}

export type ReconciliationSnapshot = {
  apiVersion: 'omnigraph/reconciliation-snapshot/v1'
  kind: 'ReconciliationSnapshot'
  metadata: { generatedAt: string; source: string; revision?: number }
  spec: {
    bom: BomDocument
    degradedNodes: { nodeId: string; reasons: string[] }[]
    fracturedEdges: { from: string; to: string; kind?: string; reason: string }[]
    relationDrifts: {
      from: string
      to: string
      relationType: BomRelationType
      category: BomRelationDriftCategory
      message: string
    }[]
    nextActions: string[]
    errors?: { code: string; message: string }[]
  }
}

/** Requires `--enable-ingest-local-api` and `Authorization: Bearer` (same token as other privileged APIs). */
export async function postLocalIngest(
  bearerToken: string,
  body: { clientSessionId?: string; files: IngestFileItem[] },
): Promise<IngestLocalResponse> {
  const base = omnigraphApiBase()
  const r = await fetch(`${base}/api/v1/ingest/local`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${bearerToken.trim()}`,
    },
    body: JSON.stringify(body),
  })
  if (!r.ok) {
    const t = await r.text()
    throw new Error(errorMessageFromResponseBody(t, r.statusText))
  }
  return r.json() as Promise<IngestLocalResponse>
}

/** Reconciliation projection for WUI triage (BOM + drift) from workspace states. */
export async function fetchReconciliationSnapshot(path: string, bearerToken?: string): Promise<ReconciliationSnapshot> {
  const base = omnigraphApiBase()
  const headers: Record<string, string> = { 'Content-Type': 'application/json' }
  if (typeof bearerToken === 'string' && bearerToken.trim()) {
    headers.Authorization = `Bearer ${bearerToken.trim()}`
  }
  const r = await fetch(`${base}/api/v1/reconciliation/snapshot`, {
    method: 'POST',
    headers,
    body: JSON.stringify({ path: path.trim() || '.' }),
  })
  if (!r.ok) {
    const t = await r.text()
    throw new Error(errorMessageFromResponseBody(t, r.statusText))
  }
  return r.json() as Promise<ReconciliationSnapshot>
}
