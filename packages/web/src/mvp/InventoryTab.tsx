import { Check, Cloud, Copy, Download, FileJson, FileText, FolderSync, HardDrive, LayoutGrid, Server, Upload } from 'lucide-react'
import { useCallback, useMemo, useRef, useState, type ChangeEvent, type DragEvent } from 'react'

import { buildInventoryViewModel, filterInventoryRows, formatLocalIngestSummary } from './buildInventoryViewModel'
import { buildReconciliationViewModel } from './buildReconciliationViewModel'
import { rowsFromIngestOmniState, type InventoryRow, type InventorySourceKind } from './inventorySources'
import { isFileSystemAccessSupported, pickInfrastructureFiles, readFilesForIngest } from './localFilePick'
import { omnigraphApiBase, postLocalIngest, type ReconciliationSnapshot, type WorkspaceSummary } from './omnigraphApi'
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
  reconciliationSnapshot: ReconciliationSnapshot | null
  /** SSE workspace stream (GET /api/v1/workspace/stream); summary updates only from events. */
  workspaceStreamConnected: boolean
  workspaceStreamError: string | null
  serveApiToken: string
  onServeApiTokenChange: (s: string) => void
}

type SourceTab = 'state' | 'plan' | 'ini'

const segBtn =
  'flex-1 rounded-md px-3 py-2 text-xs font-medium transition-colors focus:outline-none focus-visible:ring-2 focus-visible:ring-blue-500/60'

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
  const [overridesOpen, setOverridesOpen] = useState(false)
  const [ingestBusy, setIngestBusy] = useState(false)
  const [ingestErr, setIngestErr] = useState<string | null>(null)
  const [ingestNote, setIngestNote] = useState<string | null>(null)
  const [ingestDetailLines, setIngestDetailLines] = useState<string[]>([])
  const [ingestRows, setIngestRows] = useState<InventoryRow[]>([])

  const vm = useMemo(
    () =>
      buildInventoryViewModel({
        tfStateText: p.tfStateText,
        planJsonText: p.planJsonText,
        ansibleIniText: p.ansibleIniText,
        repoSession: p.repoSession,
        serverSummary: p.serverSummary,
        ingestRows,
        pipelineWorkdir: p.pipelineWorkdir,
        pipelineStateFile: p.pipelineStateFile,
        pipelinePlanFile: p.pipelinePlanFile,
        pipelineAnsibleRoot: p.pipelineAnsibleRoot,
      }),
    [
      p.repoSession,
      p.tfStateText,
      p.planJsonText,
      p.ansibleIniText,
      p.serverSummary,
      ingestRows,
      p.pipelineWorkdir,
      p.pipelineStateFile,
      p.pipelinePlanFile,
      p.pipelineAnsibleRoot,
    ],
  )

  const stateErrDisplay = vm.stateErrors.length ? vm.stateErrors.join(' · ') : undefined
  const planErrDisplay = vm.planErrors.length ? vm.planErrors.join(' · ') : undefined

  const filtered = useMemo(
    () => filterInventoryRows(vm.mergedRows, filterSource, search),
    [vm.mergedRows, filterSource, search],
  )
  const reconVm = useMemo(() => buildReconciliationViewModel(p.reconciliationSnapshot), [p.reconciliationSnapshot])

  const onPickLocalForIngest = useCallback(async () => {
    setIngestErr(null)
    setIngestNote(null)
    setIngestDetailLines([])
    const tok = p.serveApiToken.trim()
    if (!tok) {
      setIngestErr('Set the Bearer token below (same value as serve --auth-token) before ingesting.')
      return
    }
    let files: File[]
    try {
      files = await pickInfrastructureFiles()
    } catch (e) {
      setIngestErr(e instanceof Error ? e.message : String(e))
      return
    }
    if (files.length === 0) {
      return
    }
    setIngestBusy(true)
    try {
      const payloads = await readFilesForIngest(files)
      const res = await postLocalIngest(tok, { files: payloads })
      const rows = rowsFromIngestOmniState(res.state)
      setIngestRows(rows)
      const pe = res.state.partialErrors as { path?: string; code?: string; message?: string }[] | undefined
      const { note, detailLines } = formatLocalIngestSummary(
        files.length,
        res.state.nodes?.length ?? 0,
        res.errors,
        pe,
      )
      setIngestNote(note)
      setIngestDetailLines(detailLines)
    } catch (e) {
      setIngestErr(e instanceof Error ? e.message : String(e))
    } finally {
      setIngestBusy(false)
    }
  }, [p.serveApiToken])

  const copyIni = async () => {
    try {
      await navigator.clipboard.writeText(vm.omnigraphIni)
      setIniCopied(true)
      window.setTimeout(() => setIniCopied(false), 2000)
    } catch {
      setIniCopied(false)
    }
  }

  const downloadIni = () => {
    const blob = new Blob([vm.omnigraphIni], { type: 'text/plain;charset=utf-8' })
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

  const byKind = vm.discoveredByKind

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
            Bring <span className="text-gray-400">Terraform/OpenTofu state</span>, <span className="text-gray-400">plan JSON</span>, and{' '}
            <span className="text-gray-400">Ansible inventory</span> into one table. Use <strong className="text-gray-400">Load from server</strong> when
            the control plane runs same-origin, <strong className="text-gray-400">Pick files for server ingest</strong> to read local files in the browser
            and normalize on the API, or <strong className="text-gray-400">Browser folder scan</strong> for an offline pass. Contributor automation
            and <span className="text-gray-400">go test</span> paths are in <span className="text-gray-400">docs/ci-and-contributor-automation.md</span>.
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
            {p.reconciliationSnapshot ? (
              <p className="text-[10px] text-gray-600">
                BOM entities <span className="font-mono text-gray-500">{reconVm.bom.totalEntities}</span> · relations{' '}
                <span className="font-mono text-gray-500">{reconVm.bom.totalRelations}</span> · drift cues{' '}
                <span className="font-mono text-gray-500">
                  {reconVm.degradedNodeCount + reconVm.fracturedEdgeCount + reconVm.relationDriftCount}
                </span>
              </p>
            ) : null}
          </div>

          <div className="mb-4 rounded-xl border border-gray-800 bg-gray-900/40 p-3">
            <div className="mb-2 flex items-center gap-2 text-[11px] font-semibold uppercase tracking-wide text-gray-500">
              <HardDrive size={14} className="text-gray-500" aria-hidden />
              Local files → server normalize
            </div>
            <p className="mb-2 text-[10px] leading-relaxed text-gray-600">
              Pick state or inventory files with the{' '}
              <span className="text-gray-500">File System Access API</span> when supported; otherwise a multi-file dialog.
              Requires <span className="font-mono text-gray-500">POST /api/v1/ingest/local</span> enabled on serve.
            </p>
            <label className="mb-2 flex flex-col gap-1 text-[10px] text-gray-500">
              Bearer token (shared with Posture tab)
              <input
                type="password"
                autoComplete="off"
                value={p.serveApiToken}
                onChange={(e) => p.onServeApiTokenChange(e.target.value)}
                placeholder="OMNIGRAPH_SERVE_TOKEN"
                className="rounded-lg border border-gray-800 bg-gray-950 px-2 py-1.5 font-mono text-[11px] text-gray-200 placeholder:text-gray-600"
              />
            </label>
            <div className="flex flex-wrap gap-2">
              <button
                type="button"
                disabled={ingestBusy}
                onClick={() => void onPickLocalForIngest()}
                className="flex items-center gap-2 rounded-lg bg-teal-700/80 px-3 py-2 text-xs font-medium text-white hover:bg-teal-600 disabled:opacity-50"
              >
                <Upload size={14} aria-hidden />
                {ingestBusy ? 'Ingesting…' : 'Pick files for server ingest'}
              </button>
              {ingestRows.length > 0 ? (
                <button
                  type="button"
                  onClick={() => {
                    setIngestRows([])
                    setIngestNote(null)
                    setIngestDetailLines([])
                  }}
                  className="rounded-lg border border-gray-700 px-3 py-2 text-xs text-gray-400 hover:bg-gray-900"
                >
                  Clear ingest rows
                </button>
              ) : null}
            </div>
            {isFileSystemAccessSupported() ? (
              <p className="mt-2 text-[10px] text-gray-600">Native multi-file picker available in this browser.</p>
            ) : null}
            {ingestErr ? <p className="mt-2 text-[11px] text-rose-400">{ingestErr}</p> : null}
            {ingestNote ? <p className="mt-2 text-[11px] text-emerald-500/90">{ingestNote}</p> : null}
            {ingestDetailLines.length ? (
              <ul className="mt-2 max-h-24 list-inside list-disc overflow-auto text-[10px] text-amber-200/90">
                {ingestDetailLines.map((line, i) => (
                  <li key={i} className="font-mono">
                    {line}
                  </li>
                ))}
              </ul>
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
              Folder scan needs a Chromium-family browser, or use Load from server / Pick files for server ingest.
            </p>
          ) : (
            <p className="text-[11px] text-gray-600">Folder scan reads files only in your browser unless you use server ingest.</p>
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
              <span className="min-w-0 break-all font-mono text-gray-300">{vm.statePathHint}</span>
            </li>
            <li className="flex gap-2">
              <span className="shrink-0 font-mono text-[10px] text-gray-600">PLAN</span>
              <span className="min-w-0 break-all font-mono text-gray-300">{vm.planPathHint}</span>
            </li>
            <li className="flex gap-2">
              <span className="shrink-0 font-mono text-[10px] text-gray-600">INI</span>
              <span className="min-w-0 break-all font-mono text-gray-300">{vm.ansiblePathHint}</span>
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
                    error={stateErrDisplay}
                    value={p.tfStateText}
                    onChange={p.onTfStateTextChange}
                    accept=".json,application/json,.tfstate,text/plain"
                    dropLabel="Drop a state JSON file"
                  />
                ) : null}
                {activeSource === 'plan' ? (
                  <SourcePanel
                    error={planErrDisplay}
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
          <p className="text-[11px] leading-relaxed text-gray-600">
            Validation, graph emit smoke tests, and policy checks for CI are documented in{' '}
            <span className="text-gray-400">docs/ci-and-contributor-automation.md</span> (repository root).
          </p>
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
            {filtered.length} of {vm.mergedRows.length} rows
            {vm.mergedRows.length !== filtered.length ? ' (filtered)' : ''}
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
            {vm.omnigraphIni}
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
