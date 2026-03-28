import { Check, Cloud, Copy, Download, FileJson, FileText, FolderSync, LayoutGrid, Server, Upload } from 'lucide-react'
import { useCallback, useMemo, useRef, useState, type ChangeEvent, type DragEvent } from 'react'

import { joinDisplayPath } from './gitWorkspace'
import { shellQuote } from './shellQuote'
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
import { omnigraphApiBase, type WorkspaceSummary } from './omnigraphApi'
import { isRepoFolderPickerSupported, type RepoScanSession } from './repoFolderScan'

export type InventoryTabProps = {
  tfStateText: string
  onTfStateTextChange: (v: string) => void
  planJsonText: string
  onPlanJsonTextChange: (v: string) => void
  ansibleIniText: string
  onAnsibleIniTextChange: (v: string) => void
  gitRepoRoot: string
  pipelineWorkdir: string
  pipelineAnsibleRoot: string
  pipelinePlanFile: string
  pipelineStateFile: string
  repoSession: RepoScanSession | null
  onOpenRepository: () => void
  onClearRepository: () => void
  serverSummary: WorkspaceSummary | null
  onClearServer: () => void
  onLoadServer: () => Promise<void>
  serverLoading: boolean
  serverError: string | null
  /** SSE workspace stream (GET /api/v1/workspace/stream); summary updates only from events. */
  workspaceStreamConnected: boolean
  workspaceStreamError: string | null
}

type SourceTab = 'state' | 'plan' | 'ini'

const segBtn =
  'flex-1 rounded-md px-3 py-2 text-xs font-medium transition-colors focus:outline-none focus-visible:ring-2 focus-visible:ring-blue-500/60'

function withOrigin(rows: InventoryRow[], origin: string): InventoryRow[] {
  return rows.map((r) => ({
    ...r,
    id: `${r.id}@origin:${origin}`,
    originPath: origin,
  }))
}

function FileOpenButton({ accept, onText }: { accept: string; onText: (text: string) => void }) {
  const ref = useRef<HTMLInputElement>(null)
  const onChange = (e: ChangeEvent<HTMLInputElement>) => {
    const f = e.target.files?.[0]
    e.target.value = ''
    if (!f) {
      return
    }
    const reader = new FileReader()
    reader.onload = () => onText(typeof reader.result === 'string' ? reader.result : '')
    reader.readAsText(f)
  }
  return (
    <>
      <input ref={ref} type="file" accept={accept} className="hidden" aria-hidden onChange={onChange} />
      <button
        type="button"
        onClick={() => ref.current?.click()}
        className="flex items-center gap-2 rounded-lg border border-gray-700 bg-gray-900 px-3 py-1.5 text-xs text-gray-200 hover:bg-gray-800"
      >
        <Upload size={14} aria-hidden />
        Open file
      </button>
    </>
  )
}

const fieldClass =
  'min-h-[140px] w-full flex-1 resize-none rounded-xl border-0 bg-black/35 p-3 font-mono text-[11px] leading-relaxed text-gray-300 ring-1 ring-gray-800/90 focus:ring-2 focus:ring-blue-500/30 lg:min-h-0'

export function InventoryTab(p: InventoryTabProps) {
  const [activeSource, setActiveSource] = useState<SourceTab>('state')
  const [filterSource, setFilterSource] = useState<InventorySourceKind | 'all'>('all')
  const [search, setSearch] = useState('')
  const [iniCopied, setIniCopied] = useState(false)
  const [cliCopied, setCliCopied] = useState(false)
  const [overridesOpen, setOverridesOpen] = useState(false)

  const parsed = useMemo(() => {
    const stateRows: InventoryRow[] = []
    const stateErrs: string[] = []
    for (const sf of p.repoSession?.stateFiles ?? []) {
      const r = extractHostsFromTfStateJson(sf.text)
      if (r.error) {
        stateErrs.push(`${sf.path}: ${r.error}`)
      }
      stateRows.push(...withOrigin(r.rows, sf.path))
    }
    if (p.tfStateText.trim()) {
      const r = extractHostsFromTfStateJson(p.tfStateText)
      if (r.error) {
        stateErrs.push(`Overrides: ${r.error}`)
      }
      stateRows.push(...withOrigin(r.rows, 'Overrides (paste)'))
    }

    const planRows: InventoryRow[] = []
    const planErrs: string[] = []
    if (p.planJsonText.trim()) {
      const r = extractHostsFromPlanJson(p.planJsonText)
      if (r.error) {
        planErrs.push(r.error)
      }
      planRows.push(...withOrigin(r.rows, 'Overrides (paste)'))
    }

    const iniRows: InventoryRow[] = []
    for (const inf of p.repoSession?.iniFiles ?? []) {
      const r = parseAnsibleIni(inf.text)
      iniRows.push(...withOrigin(r.rows, inf.path))
    }
    if (p.ansibleIniText.trim()) {
      iniRows.push(...withOrigin(parseAnsibleIni(p.ansibleIniText).rows, 'Overrides (paste)'))
    }

    if (p.serverSummary?.stateInventory?.length) {
      let i = 0
      for (const r of p.serverSummary.stateInventory) {
        stateRows.push({
          id: `server:${r.origin}:${r.name}:${i}`,
          name: r.name,
          ansibleHost: r.ansibleHost,
          source: 'terraform-state',
          originPath: `${r.origin} (server)`,
        })
        i++
      }
    }
    if (p.serverSummary?.stateErrors?.length) {
      for (const e of p.serverSummary.stateErrors) {
        stateErrs.push(`Server: ${e}`)
      }
    }

    const merged = mergeInventoryRows(stateRows, planRows, iniRows)
    return {
      stateErr: stateErrs.join(' · ') || undefined,
      planErr: planErrs.join(' · ') || undefined,
      merged,
    }
  }, [p.repoSession, p.tfStateText, p.planJsonText, p.ansibleIniText, p.serverSummary])

  const filtered = useMemo(() => {
    const q = search.trim().toLowerCase()
    return parsed.merged.filter((r) => {
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
  }, [parsed.merged, filterSource, search])

  const omnigraphIni = useMemo(() => buildOmnigraphIni(parsed.merged), [parsed.merged])

  const statePathHint = useMemo(() => {
    const wd = p.pipelineWorkdir.trim()
    const sf = p.pipelineStateFile.trim() || 'terraform.tfstate'
    return wd ? joinDisplayPath(wd, sf) : sf
  }, [p.pipelineWorkdir, p.pipelineStateFile])

  const planPathHint = useMemo(() => {
    const wd = p.pipelineWorkdir.trim()
    const pf = p.pipelinePlanFile.trim() || 'tfplan'
    return wd ? `${joinDisplayPath(wd, pf)} → tofu show -json` : `tfplan → tofu show -json`
  }, [p.pipelineWorkdir, p.pipelinePlanFile])

  const ansiblePathHint = useMemo(() => {
    const ar = p.pipelineAnsibleRoot.trim()
    return ar ? joinDisplayPath(ar, 'inventory') : 'inventory/ under Ansible root'
  }, [p.pipelineAnsibleRoot])

  const inventoryFromStateCmd = useMemo(() => {
    const target = statePathHint.includes(' ') ? shellQuote(statePathHint) : statePathHint
    return `omnigraph inventory from-state ${target}`
  }, [statePathHint])

  const repoScanCmd = 'omnigraph repo scan --path .'
  const serveCmd = 'omnigraph serve --listen 127.0.0.1:38671 --web-dist packages/web/dist --root .'

  const copyIni = async () => {
    try {
      await navigator.clipboard.writeText(omnigraphIni)
      setIniCopied(true)
      window.setTimeout(() => setIniCopied(false), 2000)
    } catch {
      setIniCopied(false)
    }
  }

  const copyCli = async () => {
    try {
      await navigator.clipboard.writeText(inventoryFromStateCmd)
      setCliCopied(true)
      window.setTimeout(() => setCliCopied(false), 2000)
    } catch {
      setCliCopied(false)
    }
  }

  const copyRepoScan = async () => {
    try {
      await navigator.clipboard.writeText(repoScanCmd)
    } catch {
      /* ignore */
    }
  }

  const downloadIni = () => {
    const blob = new Blob([omnigraphIni], { type: 'text/plain;charset=utf-8' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = 'omnigraph-inventory.ini'
    a.click()
    URL.revokeObjectURL(url)
  }

  const sourceStyles: Record<InventorySourceKind, string> = {
    'terraform-state': 'bg-cyan-950/50 text-cyan-200 ring-1 ring-cyan-500/25',
    'plan-json': 'bg-violet-950/40 text-violet-200 ring-1 ring-violet-500/25',
    'ansible-ini': 'bg-amber-950/40 text-amber-200 ring-1 ring-amber-500/25',
  }

  const byKind = useMemo(() => {
    const m = new Map<string, number>()
    for (const d of p.repoSession?.discovered ?? []) {
      m.set(d.kind, (m.get(d.kind) ?? 0) + 1)
    }
    for (const d of p.serverSummary?.discover.files ?? []) {
      m.set(d.kind, (m.get(d.kind) ?? 0) + 1)
    }
    return m
  }, [p.repoSession, p.serverSummary])

  const pickerOk = isRepoFolderPickerSupported()
  const apiHint = omnigraphApiBase() || '(same origin as this page)'

  return (
    <div className="flex h-full min-h-0 flex-col lg:flex-row">
      <aside className="flex w-full shrink-0 flex-col border-b border-gray-800/90 bg-gray-950 lg:w-[min(100%,400px)] lg:border-b-0 lg:border-r xl:w-[430px]">
        <div className="border-b border-gray-800/90 p-5">
          <div className="mb-2 flex items-center gap-2 text-gray-200">
            <Server size={18} className="text-blue-400" aria-hidden />
            <h2 className="text-base font-semibold tracking-tight">Infrastructure inventory</h2>
          </div>
          <p className="text-[13px] leading-relaxed text-gray-500">
            The <span className="text-gray-400">repository</span> is the unit of management. Prefer{' '}
            <span className="text-gray-400">omnigraph serve</span> so the control plane reads the whole tree from disk;
            the browser picker is a fallback. Overrides are for edge cases only.
          </p>
        </div>

        <div className="border-b border-gray-800/90 p-5">
          <div className="mb-3 flex flex-col gap-2">
            <button
              type="button"
              onClick={() => void p.onLoadServer()}
              disabled={p.serverLoading}
              className="flex w-full items-center justify-center gap-2 rounded-lg bg-indigo-600 px-4 py-2.5 text-sm font-medium text-white hover:bg-indigo-500 disabled:bg-gray-800 disabled:text-gray-500"
            >
              <Cloud size={18} aria-hidden />
              {p.serverLoading ? 'Loading from server…' : 'Load from OmniGraph server'}
            </button>
            <p className="text-[10px] text-gray-600">
              API base: <span className="font-mono text-gray-500">{apiHint}</span>
              {!omnigraphApiBase() ? (
                <span className="text-gray-600"> — set </span>
              ) : null}
              {!omnigraphApiBase() ? (
                <code className="font-mono text-gray-500">VITE_OMNIGRAPH_API</code>
              ) : null}
              {!omnigraphApiBase() ? <span className="text-gray-600"> for Vite dev against a local server.</span> : null}
            </p>
            <p className="text-[10px] text-gray-600">
              Workspace SSE{' '}
              <span className="font-mono text-gray-500">/api/v1/workspace/stream</span>:{' '}
              {p.workspaceStreamConnected ? (
                <span className="text-emerald-500/90">connected</span>
              ) : (
                <span className="text-gray-600">not connected</span>
              )}
              {p.workspaceStreamError ? (
                <span className="mt-0.5 block text-rose-400/90">{p.workspaceStreamError}</span>
              ) : null}
            </p>
            {p.serverError ? <p className="text-[11px] text-rose-400/90">{p.serverError}</p> : null}
            {p.serverSummary ? (
              <div className="flex flex-wrap items-center gap-2">
                <span className="text-[11px] text-gray-500">
                  Server root: <span className="font-mono text-gray-400">{p.serverSummary.root}</span>
                </span>
                <button
                  type="button"
                  onClick={p.onClearServer}
                  className="rounded border border-gray-700 px-2 py-1 text-[10px] text-gray-400 hover:bg-gray-900"
                >
                  Clear server
                </button>
              </div>
            ) : null}
          </div>

          <div className="mb-3 flex flex-wrap items-center gap-2">
            <button
              type="button"
              onClick={p.onOpenRepository}
              disabled={!pickerOk}
              className="flex flex-1 items-center justify-center gap-2 rounded-lg border border-gray-700 bg-gray-900 px-4 py-2.5 text-sm font-medium text-gray-200 hover:bg-gray-800 disabled:cursor-not-allowed disabled:text-gray-500 min-[400px]:flex-none"
            >
              <FolderSync size={18} aria-hidden />
              Browser folder scan…
            </button>
            {p.repoSession ? (
              <button
                type="button"
                onClick={p.onClearRepository}
                className="rounded-lg border border-gray-700 px-3 py-2 text-xs font-medium text-gray-400 hover:bg-gray-900 hover:text-gray-200"
              >
                Clear browser scan
              </button>
            ) : null}
          </div>
          {!pickerOk ? (
            <p className="text-[11px] text-amber-600/90">
              Folder picker needs Chromium, or use the server and{' '}
              <code className="font-mono text-gray-500">{repoScanCmd}</code>.
            </p>
          ) : (
            <p className="text-[11px] text-gray-600">Browser scan never uploads files; everything stays local.</p>
          )}
          {p.repoSession || p.serverSummary ? (
            <div className="mt-4 rounded-xl bg-gray-900/50 p-3 ring-1 ring-gray-800/80">
              <div className="mb-2 text-[11px] font-medium uppercase tracking-wide text-gray-500">Discovered artifacts</div>
              <div className="flex max-h-28 flex-wrap gap-1.5 overflow-y-auto">
                {Array.from(byKind.entries())
                  .sort((a, b) => a[0].localeCompare(b[0]))
                  .map(([kind, n]) => (
                    <span
                      key={kind}
                      className="rounded-md bg-gray-800/80 px-2 py-0.5 font-mono text-[10px] text-gray-400"
                    >
                      {kind.replace(/-/g, ' ')} ×{n}
                    </span>
                  ))}
              </div>
              {p.repoSession ? (
                <p className="mt-2 text-[10px] text-gray-600">
                  Browser: {p.repoSession.stateFiles.length} state file(s), {p.repoSession.iniFiles.length} inventory file(s)
                  read.
                </p>
              ) : null}
            </div>
          ) : null}
        </div>

        <div className="border-b border-gray-800/90 px-5 py-4">
          <h3 className="mb-2 text-[11px] font-semibold uppercase tracking-[0.12em] text-gray-500">Path hints (Pipeline)</h3>
          <ul className="space-y-2 text-[12px] text-gray-400">
            <li className="flex gap-2">
              <span className="shrink-0 font-mono text-[10px] text-gray-600">STATE</span>
              <span className="min-w-0 break-all font-mono text-gray-300">{statePathHint}</span>
            </li>
            <li className="flex gap-2">
              <span className="shrink-0 font-mono text-[10px] text-gray-600">PLAN</span>
              <span className="min-w-0 break-all font-mono text-gray-300">{planPathHint}</span>
            </li>
            <li className="flex gap-2">
              <span className="shrink-0 font-mono text-[10px] text-gray-600">INI</span>
              <span className="min-w-0 break-all font-mono text-gray-300">{ansiblePathHint}</span>
            </li>
          </ul>
        </div>

        <div className="flex flex-1 flex-col gap-0 p-4 lg:min-h-0">
          <details
            className="mb-2 rounded-lg ring-1 ring-gray-800/80 open:bg-gray-900/20"
            open={overridesOpen}
            onToggle={(e) => setOverridesOpen(e.currentTarget.open)}
          >
            <summary className="cursor-pointer list-none px-2 py-2 text-[12px] font-medium text-gray-400 [&::-webkit-details-marker]:hidden">
              <span className="text-gray-500">Overrides</span>
              <span className="ml-2 text-[11px] font-normal text-gray-600">— paste or single file (optional)</span>
            </summary>
            <div className="border-t border-gray-800/80 px-2 pb-3 pt-2">
              <div className="mb-3 flex rounded-lg bg-gray-900/80 p-1 ring-1 ring-gray-800">
                {(
                  [
                    ['state', 'State', FileJson],
                    ['plan', 'Plan', LayoutGrid],
                    ['ini', 'Ansible', FileText],
                  ] as const
                ).map(([id, label, Icon]) => (
                  <button
                    key={id}
                    type="button"
                    onClick={() => setActiveSource(id)}
                    className={`${segBtn} flex items-center justify-center gap-1.5 ${
                      activeSource === id ? 'bg-gray-800 text-gray-100 shadow-sm' : 'text-gray-500 hover:text-gray-300'
                    }`}
                  >
                    <Icon size={14} aria-hidden />
                    {label}
                  </button>
                ))}
              </div>
              <div className="flex min-h-[200px] flex-col gap-2 lg:min-h-[180px]">
                {activeSource === 'state' ? (
                  <SourcePanel
                    error={parsed.stateErr}
                    value={p.tfStateText}
                    onChange={p.onTfStateTextChange}
                    accept=".json,application/json,.tfstate,text/plain"
                    dropLabel="Drop a state JSON file"
                  />
                ) : null}
                {activeSource === 'plan' ? (
                  <SourcePanel
                    error={parsed.planErr}
                    value={p.planJsonText}
                    onChange={p.onPlanJsonTextChange}
                    accept=".json,application/json,text/plain"
                    dropLabel="Drop plan JSON"
                  />
                ) : null}
                {activeSource === 'ini' ? (
                  <SourcePanel
                    value={p.ansibleIniText}
                    onChange={p.onAnsibleIniTextChange}
                    accept=".ini,.txt,text/plain"
                    dropLabel="Drop Ansible inventory"
                  />
                ) : null}
              </div>
            </div>
          </details>
        </div>

        <div className="mt-auto space-y-3 border-t border-gray-800/90 p-4">
          <div className="rounded-lg bg-gray-900/50 p-3 ring-1 ring-gray-800/80">
            <div className="mb-1 text-[11px] font-medium uppercase tracking-wide text-gray-500">Local server (recommended)</div>
            <code className="block whitespace-pre-wrap break-all font-mono text-[10px] leading-snug text-gray-300">
              {serveCmd}
            </code>
          </div>
          <div className="rounded-lg bg-gray-900/50 p-3 ring-1 ring-gray-800/80">
            <div className="mb-1 text-[11px] font-medium uppercase tracking-wide text-gray-500">Repo scan (CLI)</div>
            <code className="block font-mono text-[11px] text-gray-300">{repoScanCmd}</code>
            <button
              type="button"
              onClick={() => void copyRepoScan()}
              className="mt-2 text-[11px] font-medium text-blue-400 hover:text-blue-300"
            >
              Copy
            </button>
          </div>
          <div className="rounded-lg bg-gray-900/50 p-3 ring-1 ring-gray-800/80">
            <div className="mb-1 text-[11px] font-medium uppercase tracking-wide text-gray-500">Inventory from state file</div>
            <code className="block break-all font-mono text-[11px] text-gray-300">{inventoryFromStateCmd}</code>
            <button
              type="button"
              onClick={() => void copyCli()}
              className="mt-2 flex items-center gap-1.5 text-[11px] font-medium text-blue-400 hover:text-blue-300"
            >
              {cliCopied ? <Check size={12} className="text-emerald-400" aria-hidden /> : <Copy size={12} aria-hidden />}
              {cliCopied ? 'Copied' : 'Copy'}
            </button>
          </div>
        </div>
      </aside>

      <main className="flex min-h-0 min-w-0 flex-1 flex-col bg-[#0c0d10]">
        <div className="flex shrink-0 flex-col gap-3 border-b border-gray-800/90 px-5 py-4 sm:flex-row sm:items-center sm:justify-between">
          <div className="flex flex-wrap gap-1.5">
            {(['all', 'terraform-state', 'plan-json', 'ansible-ini'] as const).map((k) => (
              <button
                key={k}
                type="button"
                onClick={() => setFilterSource(k)}
                className={`rounded-full px-3 py-1 text-[11px] font-medium transition-colors ${
                  filterSource === k
                    ? 'bg-gray-100 text-gray-950'
                    : 'bg-gray-800/60 text-gray-400 hover:bg-gray-800 hover:text-gray-200'
                }`}
              >
                {k === 'all' ? 'All' : k === 'terraform-state' ? 'State' : k === 'plan-json' ? 'Plan' : 'Ansible'}
              </button>
            ))}
          </div>
          <input
            type="search"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Search host, IP, path…"
            className="w-full rounded-lg border-0 bg-gray-900/90 px-3 py-2 text-sm text-gray-200 placeholder:text-gray-600 ring-1 ring-gray-800 focus:ring-blue-500/40 sm:max-w-xs"
          />
        </div>

        <div className="min-h-0 flex-1 overflow-auto px-5 py-4">
          <div className="overflow-hidden rounded-xl ring-1 ring-gray-800/90">
            <table className="w-full border-collapse text-left text-[13px]">
              <thead>
                <tr className="border-b border-gray-800 bg-gray-950/95 text-[11px] font-medium uppercase tracking-wide text-gray-500">
                  <th className="px-4 py-3">Source</th>
                  <th className="px-4 py-3">Host</th>
                  <th className="px-4 py-3">Address</th>
                  <th className="px-4 py-3">Group</th>
                  <th className="px-4 py-3">Origin</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-800/80 bg-gray-950/40">
                {filtered.length === 0 ? (
                  <tr>
                    <td colSpan={5} className="px-4 py-16 text-center text-sm text-gray-600">
                      Scan a repository folder, or open Overrides to paste artifacts.
                    </td>
                  </tr>
                ) : (
                  filtered.map((r) => <InventoryTableRow key={r.id} row={r} sourceStyles={sourceStyles} />)
                )}
              </tbody>
            </table>
          </div>
          <p className="mt-3 text-center text-[11px] text-gray-600">
            {filtered.length} of {parsed.merged.length} rows
            {parsed.merged.length !== filtered.length ? ' (filtered)' : ''}
          </p>
        </div>

        <div className="shrink-0 border-t border-gray-800/90 bg-gray-950 px-5 py-4">
          <div className="flex flex-wrap items-center justify-between gap-3">
            <div>
              <div className="text-[11px] font-medium uppercase tracking-wide text-gray-500">Merged output</div>
              <div className="text-[12px] text-gray-500">
                INI <span className="font-mono text-gray-400">[omnigraph]</span> (deduped by host name)
              </div>
            </div>
            <div className="flex gap-2">
              <button
                type="button"
                onClick={() => void copyIni()}
                className="flex items-center gap-1.5 rounded-lg bg-gray-100 px-3 py-2 text-xs font-medium text-gray-900 hover:bg-white"
              >
                {iniCopied ? <Check size={14} className="text-emerald-600" aria-hidden /> : <Copy size={14} aria-hidden />}
                {iniCopied ? 'Copied' : 'Copy'}
              </button>
              <button
                type="button"
                onClick={downloadIni}
                className="flex items-center gap-1.5 rounded-lg bg-gray-800 px-3 py-2 text-xs font-medium text-gray-200 ring-1 ring-gray-700 hover:bg-gray-700"
              >
                <Download size={14} aria-hidden />
                Download
              </button>
            </div>
          </div>
          <pre className="mt-3 max-h-36 overflow-auto rounded-lg bg-black/40 p-3 font-mono text-[11px] leading-relaxed text-gray-400 ring-1 ring-gray-800/80">
            {omnigraphIni}
          </pre>
        </div>
      </main>
    </div>
  )
}

function SourcePanel({
  value,
  onChange,
  accept,
  dropLabel,
  error,
}: {
  value: string
  onChange: (v: string) => void
  accept: string
  dropLabel: string
  error?: string
}) {
  const onDrop = useCallback(
    (e: DragEvent) => {
      e.preventDefault()
      const f = e.dataTransfer.files?.[0]
      if (!f) {
        return
      }
      const reader = new FileReader()
      reader.onload = () => onChange(typeof reader.result === 'string' ? reader.result : '')
      reader.readAsText(f)
    },
    [onChange],
  )

  return (
    <div className="flex min-h-0 flex-1 flex-col gap-2">
      <div
        onDragOver={(e) => e.preventDefault()}
        onDrop={onDrop}
        className="flex shrink-0 flex-col items-center justify-center rounded-xl border border-dashed border-gray-700/90 bg-gray-900/30 px-4 py-5 text-center hover:border-gray-600 hover:bg-gray-900/50"
      >
        <Upload size={20} className="mb-2 text-gray-600" aria-hidden />
        <p className="text-[12px] text-gray-500">{dropLabel}</p>
        <FileOpenButton accept={accept} onText={onChange} />
      </div>
      <textarea
        value={value}
        onChange={(e) => onChange(e.target.value)}
        spellCheck={false}
        className={fieldClass}
        placeholder="Paste here if needed"
      />
      {error ? <p className="text-[11px] text-rose-400/90">{error}</p> : null}
    </div>
  )
}

function InventoryTableRow({
  row,
  sourceStyles,
}: {
  row: InventoryRow
  sourceStyles: Record<InventorySourceKind, string>
}) {
  return (
    <tr className="transition-colors hover:bg-gray-900/60">
      <td className="px-4 py-2.5 align-middle">
        <span
          className={`inline-flex rounded-md px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wide ${sourceStyles[row.source]}`}
        >
          {row.source === 'terraform-state' ? 'State' : row.source === 'plan-json' ? 'Plan' : 'Ansible'}
        </span>
      </td>
      <td className="px-4 py-2.5 font-mono text-[12px] text-gray-200">{row.name}</td>
      <td className="px-4 py-2.5 font-mono text-[12px] text-gray-400">{row.ansibleHost || '—'}</td>
      <td className="px-4 py-2.5 text-[12px] text-gray-500">{row.group ?? '—'}</td>
      <td className="max-w-[200px] truncate px-4 py-2.5 font-mono text-[11px] text-gray-600" title={row.originPath}>
        {row.originPath ?? '—'}
      </td>
    </tr>
  )
}
