/**
 * Pure assembly of inventory merge + path hints + discovery counts.
 * Single place for merge precedence: repo scan → paste overrides → server summary → ingest rows (via mergeInventoryRows).
 */

import { joinDisplayPath } from './gitWorkspace'
import {
  buildOmnigraphIni,
  extractHostsFromPlanJson,
  extractHostsFromTfStateJson,
  mergeInventoryRows,
  parseAnsibleIni,
  sourceLabel,
  type InventoryRow,
  type InventorySourceKind,
} from './inventorySources'
import type { WorkspaceSummary } from './omnigraphApi'
import { type RepoScanSession } from './repoFolderScan'

export type BuildInventoryViewModelInput = {
  tfStateText: string
  planJsonText: string
  ansibleIniText: string
  repoSession: RepoScanSession | null
  serverSummary: WorkspaceSummary | null
  ingestRows: InventoryRow[]
  pipelineWorkdir: string
  pipelineStateFile: string
  pipelinePlanFile: string
  pipelineAnsibleRoot: string
}

export type InventoryViewModel = {
  mergedRows: InventoryRow[]
  /** Parse/merge issues for state (Terraform/OpenTofu JSON)—operator-facing strings. */
  stateErrors: string[]
  /** Parse issues for plan JSON—operator-facing strings. */
  planErrors: string[]
  omnigraphIni: string
  /** Row counts by inventory source kind after merge. */
  countsBySource: Record<InventorySourceKind, number>
  /** Discovered artifact kinds (browser scan + server discovery) for badges. */
  discoveredByKind: Map<string, number>
  statePathHint: string
  planPathHint: string
  ansiblePathHint: string
}

function withOrigin(rows: InventoryRow[], origin: string): InventoryRow[] {
  return rows.map((r) => ({
    ...r,
    id: `${r.id}@origin:${origin}`,
    originPath: origin,
  }))
}

function countBySource(rows: InventoryRow[]): Record<InventorySourceKind, number> {
  const out: Record<InventorySourceKind, number> = {
    'terraform-state': 0,
    'plan-json': 0,
    'ansible-ini': 0,
  }
  for (const r of rows) {
    out[r.source] = (out[r.source] ?? 0) + 1
  }
  return out
}

function buildDiscoveredByKind(
  repoSession: RepoScanSession | null,
  serverSummary: WorkspaceSummary | null,
): Map<string, number> {
  const m = new Map<string, number>()
  for (const d of repoSession?.discovered ?? []) {
    m.set(d.kind, (m.get(d.kind) ?? 0) + 1)
  }
  for (const d of serverSummary?.discover.files ?? []) {
    m.set(d.kind, (m.get(d.kind) ?? 0) + 1)
  }
  return m
}

export function buildInventoryViewModel(input: BuildInventoryViewModelInput): InventoryViewModel {
  const stateRows: InventoryRow[] = []
  const stateErrors: string[] = []
  for (const sf of input.repoSession?.stateFiles ?? []) {
    const r = extractHostsFromTfStateJson(sf.text)
    if (r.error) {
      stateErrors.push(`State artifact ${sf.path}: ${r.error}`)
    }
    stateRows.push(...withOrigin(r.rows, sf.path))
  }
  if (input.tfStateText.trim()) {
    const r = extractHostsFromTfStateJson(input.tfStateText)
    if (r.error) {
      stateErrors.push(`Overrides (paste): ${r.error}`)
    }
    stateRows.push(...withOrigin(r.rows, 'Overrides (paste)'))
  }

  const planRows: InventoryRow[] = []
  const planErrors: string[] = []
  if (input.planJsonText.trim()) {
    const r = extractHostsFromPlanJson(input.planJsonText)
    if (r.error) {
      planErrors.push(r.error)
    }
    planRows.push(...withOrigin(r.rows, 'Overrides (paste)'))
  }

  const iniRows: InventoryRow[] = []
  for (const inf of input.repoSession?.iniFiles ?? []) {
    const r = parseAnsibleIni(inf.text)
    iniRows.push(...withOrigin(r.rows, inf.path))
  }
  if (input.ansibleIniText.trim()) {
    iniRows.push(...withOrigin(parseAnsibleIni(input.ansibleIniText).rows, 'Overrides (paste)'))
  }

  if (input.serverSummary?.stateInventory?.length) {
    let i = 0
    for (const r of input.serverSummary.stateInventory) {
      stateRows.push({
        id: `server:${r.origin}:${r.name}:${i}`,
        name: r.name,
        ansibleHost: r.ansibleHost,
        source: 'terraform-state',
        originPath: `${r.origin} (server summary)`,
      })
      i++
    }
  }
  if (input.serverSummary?.stateErrors?.length) {
    for (const e of input.serverSummary.stateErrors) {
      stateErrors.push(`Server workspace: ${e}`)
    }
  }

  const ingestMerged = input.ingestRows.length ? withOrigin(input.ingestRows, 'POST /ingest/local') : []
  const mergedRows = mergeInventoryRows(stateRows, planRows, [...iniRows, ...ingestMerged])

  const wd = input.pipelineWorkdir.trim()
  const sf = input.pipelineStateFile.trim() || 'terraform.tfstate'
  const statePathHint = wd ? joinDisplayPath(wd, sf) : sf
  const pf = input.pipelinePlanFile.trim() || 'tfplan'
  const planPathHint = wd ? `${joinDisplayPath(wd, pf)} → tofu show -json` : `tfplan → tofu show -json`
  const ar = input.pipelineAnsibleRoot.trim()
  const ansiblePathHint = ar ? joinDisplayPath(ar, 'inventory') : 'inventory/ under Ansible root'

  return {
    mergedRows,
    stateErrors,
    planErrors,
    omnigraphIni: buildOmnigraphIni(mergedRows),
    countsBySource: countBySource(mergedRows),
    discoveredByKind: buildDiscoveredByKind(input.repoSession, input.serverSummary),
    statePathHint,
    planPathHint,
    ansiblePathHint,
  }
}

export function filterInventoryRows(
  merged: InventoryRow[],
  filterSource: InventorySourceKind | 'all',
  search: string,
): InventoryRow[] {
  const q = search.trim().toLowerCase()
  return merged.filter((r) => {
    if (filterSource !== 'all' && r.source !== filterSource) {
      return false
    }
    if (!q) {
      return true
    }
    const g = (r.group ?? '').toLowerCase()
    const o = (r.originPath ?? '').toLowerCase()
    return (
      r.name.toLowerCase().includes(q) ||
      r.ansibleHost.toLowerCase().includes(q) ||
      g.includes(q) ||
      o.includes(q) ||
      sourceLabel(r.source).toLowerCase().includes(q)
    )
  })
}

type IngestPartialErr = { path?: string; code?: string; message?: string }

/** Operator-facing summary line after successful POST /api/v1/ingest/local (HTTP 200). */
export function formatLocalIngestSummary(
  fileCount: number,
  nodeCount: number,
  topLevelErrors: { path?: string; code: string; message: string }[] | undefined,
  partialErrors: IngestPartialErr[] | undefined,
): { note: string; detailLines: string[] } {
  const detailLines: string[] = []
  for (const e of topLevelErrors ?? []) {
    const p = e.path ? `${e.path}: ` : ''
    detailLines.push(`ingest/local [${e.code}] ${p}${e.message}`)
  }
  for (const e of partialErrors ?? []) {
    const p = e.path ? `${e.path}: ` : ''
    const c = e.code ? `[${e.code}] ` : ''
    const m = typeof e.message === 'string' ? e.message : String(e)
    detailLines.push(`ingest/local (partial) ${p}${c}${m}`)
  }
  const errCount = (topLevelErrors?.length ?? 0) + (partialErrors?.length ?? 0)
  const note =
    `Normalized ${nodeCount} node(s) from ${fileCount} file(s)` +
    (errCount ? ` (${errCount} partial issue(s); see details below if shown).` : '.')
  return { note, detailLines }
}
