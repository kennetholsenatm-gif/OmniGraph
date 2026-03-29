/** Base URL for `omnigraph serve` API. Empty = same origin (UI served by the binary). */
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

/** Requires `omnigraph serve --enable-security-scan --auth-token …` and same-origin or VITE_OMNIGRAPH_API. */
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
    throw new Error(t || r.statusText)
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

/** Requires `serve --enable-ingest-local-api` and `Authorization: Bearer` (same token as other privileged APIs). */
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
    throw new Error(t || r.statusText)
  }
  return r.json() as Promise<IngestLocalResponse>
}
