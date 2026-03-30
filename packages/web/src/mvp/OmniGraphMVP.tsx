import { FolderGit2, Layers, Network, RotateCcw, Server, Settings, Shield, TerminalSquare, Workflow } from 'lucide-react'
import { useCallback, useEffect, useState, type ReactNode } from 'react'

import { initHclWasm } from '../hclWasm'
import { wasmSpikeAdd, wasmSpikeEnabled } from '../wasmSpike'
import type { GraphNodeSelection } from '../graph/GraphCanvas'
import { defaultWorkspaceSnapshot, loadWorkspace, persistWorkspace, clearWorkspaceStorage, type WorkspaceSnapshotV1 } from './workspaceStorage'
import { buildWorkspaceManifest, manifestToJson } from './gitWorkspace'
import { tryParseMetadataName } from './parseMetadataName'
import { defaultPostureSecurityJson, mvpTabDisplayName, type MvpTab } from './constants'
import { GraphVisualizerTab } from './GraphVisualizerTab'
import { InventoryTab } from './InventoryTab'
import { fetchReconciliationSnapshot, fetchWorkspaceSummary, type ReconciliationSnapshot, type WorkspaceSummary } from './omnigraphApi'
import { scanRepositoryFolder, type RepoScanSession } from './repoFolderScan'
import { NavItem } from './NavItem'
import { PipelineTab } from './PipelineTab'
import { SchemaTab } from './SchemaTab'
import { PostureTab } from './PostureTab'
import { WebIDETab } from './WebIDETab'
import { useWorkspaceSummaryStream } from './useWorkspaceSummaryStream'

const webVersion = __OMNIGRAPH_WEB_VERSION__

function NavSectionLabel({ children }: { children: ReactNode }) {
  return (
    <div className="hidden px-2 pb-0.5 pt-3 text-[10px] font-semibold uppercase tracking-wider text-gray-600 first:pt-1 md:block">
      {children}
    </div>
  )
}

function wasmStatusLabel(s: 'loading' | 'ok' | 'err'): string {
  switch (s) {
    case 'ok':
      return 'HCL Wasm ready'
    case 'err':
      return 'HCL Wasm unavailable'
    default:
      return 'Loading Wasm…'
  }
}

function wasmStatusDotClass(s: 'loading' | 'ok' | 'err'): string {
  switch (s) {
    case 'ok':
      return 'bg-emerald-500'
    case 'err':
      return 'bg-rose-500'
    default:
      return 'animate-pulse bg-amber-500'
  }
}

function snapshotFromState(args: {
  schemaText: string
  graphText: string
  hclText: string
  projectLabel: string
  gitRepoRoot: string
  schemaManifestRelativePath: string
  schemaFileNameHint?: string
  graphFileNameHint?: string
  pipelineWorkdir: string
  pipelineAnsibleRoot: string
  pipelinePlaybookRel: string
  pipelinePlaybookOverride: string
  pipelineUsePlaybookOverride: boolean
  pipelineSchema: string
  pipelineTfBinary: string
  pipelinePlanFile: string
  pipelineStateFile: string
  pipelineRunner: 'exec' | 'container'
  pipelineContainerRuntime: string
  pipelineAutoApprove: boolean
  pipelineSkipAnsible: boolean
  pipelineGraphOut: string
  pipelineTelemetryFile: string
  pipelineIacEngine: string
  pipelineTofuImage: string
  pipelineAnsibleImage: string
  pipelinePulumiImage: string
  inventoryTfStateText: string
  inventoryPlanJsonText: string
  inventoryAnsibleIniText: string
  postureSecurityJson: string
  serveApiToken: string
}): WorkspaceSnapshotV1 {
  return {
    v: 1,
    schemaText: args.schemaText,
    graphText: args.graphText,
    hclText: args.hclText,
    projectLabel: args.projectLabel,
    gitRepoRoot: args.gitRepoRoot,
    schemaManifestRelativePath: args.schemaManifestRelativePath,
    schemaFileNameHint: args.schemaFileNameHint,
    graphFileNameHint: args.graphFileNameHint,
    pipelineWorkdir: args.pipelineWorkdir,
    pipelineAnsibleRoot: args.pipelineAnsibleRoot,
    pipelinePlaybookRel: args.pipelinePlaybookRel,
    pipelinePlaybookOverride: args.pipelinePlaybookOverride,
    pipelineUsePlaybookOverride: args.pipelineUsePlaybookOverride,
    pipelineSchema: args.pipelineSchema,
    pipelineTfBinary: args.pipelineTfBinary,
    pipelinePlanFile: args.pipelinePlanFile,
    pipelineStateFile: args.pipelineStateFile,
    pipelineRunner: args.pipelineRunner,
    pipelineContainerRuntime: args.pipelineContainerRuntime,
    pipelineAutoApprove: args.pipelineAutoApprove,
    pipelineSkipAnsible: args.pipelineSkipAnsible,
    pipelineGraphOut: args.pipelineGraphOut,
    pipelineTelemetryFile: args.pipelineTelemetryFile,
    pipelineIacEngine: args.pipelineIacEngine,
    pipelineTofuImage: args.pipelineTofuImage,
    pipelineAnsibleImage: args.pipelineAnsibleImage,
    pipelinePulumiImage: args.pipelinePulumiImage,
    inventoryTfStateText: args.inventoryTfStateText,
    inventoryPlanJsonText: args.inventoryPlanJsonText,
    inventoryAnsibleIniText: args.inventoryAnsibleIniText,
    postureSecurityJson: args.postureSecurityJson,
    serveApiToken: args.serveApiToken.trim() ? args.serveApiToken : undefined,
  }
}

export default function OmniGraphMVP() {
  const initial = loadWorkspace() ?? defaultWorkspaceSnapshot()

  const [activeTab, setActiveTab] = useState<MvpTab>('visualizer')
  const [schemaText, setSchemaText] = useState(initial.schemaText)
  const [graphText, setGraphText] = useState(initial.graphText)
  const [hclText, setHclText] = useState(initial.hclText)
  const [projectLabel, setProjectLabel] = useState(initial.projectLabel)
  const [gitRepoRoot, setGitRepoRoot] = useState(initial.gitRepoRoot ?? '')
  const [schemaManifestRelativePath, setSchemaManifestRelativePath] = useState(
    initial.schemaManifestRelativePath,
  )
  const [schemaFileNameHint, setSchemaFileNameHint] = useState<string | undefined>(initial.schemaFileNameHint)
  const [graphFileNameHint, setGraphFileNameHint] = useState<string | undefined>(initial.graphFileNameHint)

  const [pipelineWorkdir, setPipelineWorkdir] = useState(initial.pipelineWorkdir ?? '')
  const [pipelineAnsibleRoot, setPipelineAnsibleRoot] = useState(initial.pipelineAnsibleRoot ?? '')
  const [pipelinePlaybookRel, setPipelinePlaybookRel] = useState(initial.pipelinePlaybookRel ?? 'site.yml')
  const [pipelinePlaybookOverride, setPipelinePlaybookOverride] = useState(initial.pipelinePlaybookOverride ?? '')
  const [pipelineUsePlaybookOverride, setPipelineUsePlaybookOverride] = useState(initial.pipelineUsePlaybookOverride ?? false)
  const [pipelineSchema, setPipelineSchema] = useState(initial.pipelineSchema ?? '.omnigraph.schema')
  const [pipelineTfBinary, setPipelineTfBinary] = useState(initial.pipelineTfBinary ?? 'tofu')
  const [pipelinePlanFile, setPipelinePlanFile] = useState(initial.pipelinePlanFile ?? 'tfplan')
  const [pipelineStateFile, setPipelineStateFile] = useState(initial.pipelineStateFile ?? 'terraform.tfstate')
  const [pipelineRunner, setPipelineRunner] = useState<'exec' | 'container'>(initial.pipelineRunner ?? 'exec')
  const [pipelineContainerRuntime, setPipelineContainerRuntime] = useState(initial.pipelineContainerRuntime ?? '')
  const [pipelineAutoApprove, setPipelineAutoApprove] = useState(initial.pipelineAutoApprove ?? false)
  const [pipelineSkipAnsible, setPipelineSkipAnsible] = useState(initial.pipelineSkipAnsible ?? false)
  const [pipelineGraphOut, setPipelineGraphOut] = useState(initial.pipelineGraphOut ?? '')
  const [pipelineTelemetryFile, setPipelineTelemetryFile] = useState(initial.pipelineTelemetryFile ?? '')
  const [pipelineIacEngine, setPipelineIacEngine] = useState(initial.pipelineIacEngine ?? '')
  const [pipelineTofuImage, setPipelineTofuImage] = useState(initial.pipelineTofuImage ?? '')
  const [pipelineAnsibleImage, setPipelineAnsibleImage] = useState(initial.pipelineAnsibleImage ?? '')
  const [pipelinePulumiImage, setPipelinePulumiImage] = useState(initial.pipelinePulumiImage ?? '')

  const [inventoryTfStateText, setInventoryTfStateText] = useState(initial.inventoryTfStateText ?? '')
  const [inventoryPlanJsonText, setInventoryPlanJsonText] = useState(initial.inventoryPlanJsonText ?? '')
  const [inventoryAnsibleIniText, setInventoryAnsibleIniText] = useState(initial.inventoryAnsibleIniText ?? '')
  const [postureSecurityJson, setPostureSecurityJson] = useState(initial.postureSecurityJson ?? defaultPostureSecurityJson)
  const [serveApiToken, setServeApiToken] = useState(initial.serveApiToken ?? '')

  const [repoScanSession, setRepoScanSession] = useState<RepoScanSession | null>(null)
  const [serverSummary, setServerSummary] = useState<WorkspaceSummary | null>(null)
  const [reconciliationSnapshot, setReconciliationSnapshot] = useState<ReconciliationSnapshot | null>(null)
  const [serverLoading, setServerLoading] = useState(false)
  const [serverError, setServerError] = useState<string | null>(null)

  const [selectedGraphNode, setSelectedGraphNode] = useState<GraphNodeSelection | null>(null)
  const [hclWasm, setHclWasm] = useState<'loading' | 'ok' | 'err'>('loading')
  const [wasmNote, setWasmNote] = useState<string | null>(null)

  const workspaceStreamPath = gitRepoRoot.trim() || '.'
  const { connected: workspaceStreamConnected, error: workspaceStreamError } = useWorkspaceSummaryStream(
    workspaceStreamPath,
    useCallback((s: WorkspaceSummary) => {
      setServerSummary(s)
      setServerError(null)
    }, []),
  )

  useEffect(() => {
    initHclWasm()
      .then(() => setHclWasm('ok'))
      .catch(() => setHclWasm('err'))
  }, [])

  const switchTab = useCallback((tab: MvpTab) => {
    setActiveTab(tab)
    // Keep triage context deterministic: node selection belongs to Topology only.
    if (tab !== 'visualizer') {
      setSelectedGraphNode(null)
    }
  }, [])

  useEffect(() => {
    if (!wasmSpikeEnabled()) {
      return
    }
    let cancelled = false
    wasmSpikeAdd(2, 3)
      .then((n) => {
        if (!cancelled) {
          setWasmNote(`Wasm spike: add(2,3) = ${n} (VITE_ENABLE_WASM_SPIKE)`)
        }
      })
      .catch((e: unknown) => {
        if (!cancelled) {
          const m = e instanceof Error ? e.message : String(e)
          setWasmNote(`Wasm spike failed: ${m}`)
        }
      })
    return () => {
      cancelled = true
    }
  }, [])

  useEffect(() => {
    const t = window.setTimeout(() => {
      persistWorkspace(
        snapshotFromState({
          schemaText,
          graphText,
          hclText,
          projectLabel,
          gitRepoRoot,
          schemaManifestRelativePath,
          schemaFileNameHint,
          graphFileNameHint,
          pipelineWorkdir,
          pipelineAnsibleRoot,
          pipelinePlaybookRel,
          pipelinePlaybookOverride,
          pipelineUsePlaybookOverride,
          pipelineSchema,
          pipelineTfBinary,
          pipelinePlanFile,
          pipelineStateFile,
          pipelineRunner,
          pipelineContainerRuntime,
          pipelineAutoApprove,
          pipelineSkipAnsible,
          pipelineGraphOut,
          pipelineTelemetryFile,
          pipelineIacEngine,
          pipelineTofuImage,
          pipelineAnsibleImage,
          pipelinePulumiImage,
          inventoryTfStateText,
          inventoryPlanJsonText,
          inventoryAnsibleIniText,
          postureSecurityJson,
          serveApiToken,
        }),
      )
    }, 400)
    return () => window.clearTimeout(t)
  }, [
    schemaText,
    graphText,
    hclText,
    projectLabel,
    gitRepoRoot,
    schemaManifestRelativePath,
    schemaFileNameHint,
    graphFileNameHint,
    pipelineWorkdir,
    pipelineAnsibleRoot,
    pipelinePlaybookRel,
    pipelinePlaybookOverride,
    pipelineUsePlaybookOverride,
    pipelineSchema,
    pipelineTfBinary,
    pipelinePlanFile,
    pipelineStateFile,
    pipelineRunner,
    pipelineContainerRuntime,
    pipelineAutoApprove,
    pipelineSkipAnsible,
    pipelineGraphOut,
    pipelineTelemetryFile,
    pipelineIacEngine,
    pipelineTofuImage,
    pipelineAnsibleImage,
    pipelinePulumiImage,
    inventoryTfStateText,
    inventoryPlanJsonText,
    inventoryAnsibleIniText,
    postureSecurityJson,
    serveApiToken,
  ])

  const resetWorkspace = () => {
    if (!window.confirm('Reset local workspace? Unsaved browser state will revert to defaults.')) {
      return
    }
    clearWorkspaceStorage()
    const d = defaultWorkspaceSnapshot()
    setSchemaText(d.schemaText)
    setGraphText(d.graphText)
    setHclText(d.hclText)
    setProjectLabel(d.projectLabel)
    setGitRepoRoot('')
    setSchemaManifestRelativePath(d.schemaManifestRelativePath)
    setSchemaFileNameHint(undefined)
    setGraphFileNameHint(undefined)
    setPipelineWorkdir('')
    setPipelineAnsibleRoot('')
    setPipelinePlaybookRel('site.yml')
    setPipelinePlaybookOverride('')
    setPipelineUsePlaybookOverride(false)
    setPipelineSchema('.omnigraph.schema')
    setPipelineTfBinary('tofu')
    setPipelinePlanFile('tfplan')
    setPipelineStateFile('terraform.tfstate')
    setPipelineRunner('exec')
    setPipelineContainerRuntime('')
    setPipelineAutoApprove(false)
    setPipelineSkipAnsible(false)
    setPipelineGraphOut('')
    setPipelineTelemetryFile('')
    setPipelineIacEngine('')
    setPipelineTofuImage('')
    setPipelineAnsibleImage('')
    setPipelinePulumiImage('')
    setInventoryTfStateText('')
    setInventoryPlanJsonText('')
    setInventoryAnsibleIniText('')
    setPostureSecurityJson(defaultPostureSecurityJson)
    setRepoScanSession(null)
    setServerSummary(null)
    setReconciliationSnapshot(null)
    setServerError(null)
    setSelectedGraphNode(null)
  }

  const loadFromOmnigraphServer = useCallback(async () => {
    setServerError(null)
    setServerLoading(true)
    try {
      const s = await fetchWorkspaceSummary('.')
      setServerSummary(s)
      try {
        const snap = await fetchReconciliationSnapshot('.', serveApiToken)
        setReconciliationSnapshot(snap)
      } catch {
        setReconciliationSnapshot(null)
      }
    } catch (e: unknown) {
      setServerSummary(null)
      setReconciliationSnapshot(null)
      setServerError(e instanceof Error ? e.message : String(e))
    } finally {
      setServerLoading(false)
    }
  }, [serveApiToken])

  const openRepositoryFolder = useCallback(async () => {
    const s = await scanRepositoryFolder()
    if (!s) {
      return
    }
    setRepoScanSession(s)
    const sch = s.schemas[0]
    if (sch) {
      setSchemaText(sch.text)
      setSchemaFileNameHint(sch.path)
    }
  }, [])

  const downloadWorkspaceManifest = () => {
    const m = buildWorkspaceManifest({
      gitRepositoryRoot: gitRepoRoot,
      pipelineWorkdir,
      pipelineAnsibleRoot,
      pipelinePlaybookRel,
      schemaManifestRelativePath,
    })
    const blob = new Blob([manifestToJson(m)], { type: 'application/json;charset=utf-8' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = 'omnigraph.workspace.json'
    a.click()
    URL.revokeObjectURL(url)
  }

  const syncNameFromSchema = () => {
    const n = tryParseMetadataName(schemaText)
    if (n) {
      setProjectLabel(n)
    } else {
      window.alert('Could not read metadata.name from the current document (parse error or missing field).')
    }
  }

  return (
    <div className="flex h-dvh min-h-dvh overflow-hidden bg-gray-950 font-sans text-gray-100 selection:bg-blue-500/30">
      <div className="flex w-16 shrink-0 flex-col border-r border-gray-800 bg-gray-900 transition-all duration-300 md:w-64">
        <div className="flex items-center justify-center gap-3 border-b border-gray-800 p-4 md:justify-start">
          <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-gradient-to-br from-blue-500 to-indigo-600 shadow-lg shadow-blue-500/20">
            <Layers size={18} className="text-white" aria-hidden />
          </div>
          <span className="hidden bg-gradient-to-r from-gray-100 to-gray-400 bg-clip-text text-lg font-bold tracking-wide text-transparent md:block">
            OmniGraph
          </span>
        </div>

        <nav className="flex flex-1 flex-col gap-1 px-2 py-6">
          <NavSectionLabel>Operational contexts</NavSectionLabel>
          <NavItem
            icon={<Network size={20} aria-hidden />}
            label="Topology"
            active={activeTab === 'visualizer'}
            onClick={() => switchTab('visualizer')}
          />
          <NavSectionLabel>Reconciliation</NavSectionLabel>
          <NavItem
            icon={<Server size={20} aria-hidden />}
            label="Inventory"
            active={activeTab === 'inventory'}
            onClick={() => switchTab('inventory')}
            indent
          />
          <NavItem
            icon={<Workflow size={20} aria-hidden />}
            label="Pipeline"
            active={activeTab === 'pipeline'}
            onClick={() => switchTab('pipeline')}
            indent
          />
          <NavItem
            icon={<Shield size={20} aria-hidden />}
            label="Posture"
            active={activeTab === 'posture'}
            onClick={() => switchTab('posture')}
          />
          <NavSectionLabel>Supporting editors</NavSectionLabel>
          <NavItem
            icon={<Settings size={20} aria-hidden />}
            label="Schema Contract"
            active={activeTab === 'schema'}
            onClick={() => switchTab('schema')}
          />
          <NavItem
            icon={<TerminalSquare size={20} aria-hidden />}
            label="Web IDE"
            active={activeTab === 'ide'}
            onClick={() => switchTab('ide')}
          />
        </nav>

        <div className="hidden flex-col gap-2 border-t border-gray-800 p-4 text-xs text-gray-500 md:flex">
          <div className="flex items-center gap-2">
            <span className={`h-2 w-2 shrink-0 rounded-full ${wasmStatusDotClass(hclWasm)}`} aria-hidden />
            <span>{wasmStatusLabel(hclWasm)}</span>
          </div>
          <button
            type="button"
            onClick={resetWorkspace}
            className="flex items-center gap-2 rounded-lg border border-gray-800 bg-gray-950 px-2 py-2 text-left text-gray-400 hover:bg-gray-800 hover:text-gray-200"
          >
            <RotateCcw size={14} aria-hidden />
            Reset workspace
          </button>
        </div>
      </div>

      <div className="relative flex min-w-0 flex-1 flex-col overflow-hidden bg-[radial-gradient(ellipse_at_top,_var(--tw-gradient-stops))] from-gray-900 via-gray-950 to-gray-950">
        <header className="z-10 flex h-14 shrink-0 items-center justify-between gap-4 border-b border-gray-800 bg-gray-900/50 px-4 backdrop-blur-sm md:px-6">
          <div className="flex min-w-0 flex-1 flex-col gap-1 sm:flex-row sm:items-center sm:gap-3">
            <h1 className="flex min-w-0 items-center gap-2 text-sm font-medium text-gray-300">
              <span className="shrink-0 text-gray-500">workspace /</span>
              <input
                type="text"
                value={projectLabel}
                onChange={(e) => setProjectLabel(e.target.value)}
                className="min-w-0 max-w-[200px] rounded border border-transparent bg-transparent px-1 py-0.5 font-medium text-gray-100 outline-none focus:border-gray-700 sm:max-w-xs"
                aria-label="Project label"
              />
              <span className="shrink-0 text-gray-500">/</span>
              <span className="truncate text-gray-400">{mvpTabDisplayName(activeTab)}</span>
            </h1>
            <button
              type="button"
              onClick={syncNameFromSchema}
              className="shrink-0 self-start rounded border border-gray-700 bg-gray-900 px-2 py-1 text-xs text-gray-400 hover:bg-gray-800 sm:self-auto"
            >
              Sync name from schema
            </button>
          </div>
          <span className="shrink-0 rounded border border-gray-800 bg-gray-900/80 px-2 py-1 font-mono text-xs text-gray-500">
            web v{webVersion}
          </span>
        </header>

        <div className="flex shrink-0 flex-wrap items-end gap-3 border-b border-gray-800 bg-gray-950/90 px-4 py-2.5 md:items-center md:px-6">
          <div className="flex min-w-0 flex-1 flex-col gap-1 sm:flex-row sm:items-center sm:gap-3">
            <span className="flex shrink-0 items-center gap-1.5 text-[11px] font-semibold uppercase tracking-wide text-gray-500">
              <FolderGit2 size={14} className="text-gray-600" aria-hidden />
              Git repository root
            </span>
            <input
              type="text"
              value={gitRepoRoot}
              onChange={(e) => setGitRepoRoot(e.target.value)}
              placeholder="e.g. /path/to/repo"
              className="min-w-0 flex-1 rounded-lg border border-gray-800 bg-gray-900/80 px-3 py-1.5 font-mono text-xs text-gray-200 placeholder:text-gray-600 focus:border-gray-700 focus:outline-none focus:ring-1 focus:ring-blue-500/40"
              aria-label="Git repository root path"
            />
          </div>
          <button
            type="button"
            onClick={downloadWorkspaceManifest}
            className="shrink-0 rounded-lg border border-gray-700 bg-gray-900 px-3 py-1.5 text-xs font-medium text-gray-200 hover:bg-gray-800"
          >
            Export omnigraph.workspace.json
          </button>
        </div>

        <main className="min-h-0 flex-1 overflow-hidden">
          {activeTab === 'visualizer' ? (
            <GraphVisualizerTab
              graphText={graphText}
              onGraphTextChange={setGraphText}
              selectedNode={selectedGraphNode}
              onNodeSelect={setSelectedGraphNode}
              reconciliationSnapshot={reconciliationSnapshot}
              graphFileNameHint={graphFileNameHint}
              onGraphFileNameHintChange={setGraphFileNameHint}
            />
          ) : null}
          {activeTab === 'schema' ? (
            <SchemaTab
              schemaText={schemaText}
              onSchemaChange={setSchemaText}
              schemaManifestRelativePath={schemaManifestRelativePath}
              onSchemaManifestRelativePathChange={setSchemaManifestRelativePath}
              schemaFileNameHint={schemaFileNameHint}
              onSchemaFileNameHintChange={setSchemaFileNameHint}
              pipelineSchemaPath={pipelineSchema}
              onApplyManifestPathToPipeline={() =>
                setPipelineSchema(schemaManifestRelativePath.trim() || '.omnigraph.schema')
              }
            />
          ) : null}
          {activeTab === 'ide' ? <WebIDETab hclWasm={hclWasm} hclText={hclText} onHclChange={setHclText} /> : null}
          {activeTab === 'inventory' ? (
            <InventoryTab
              tfStateText={inventoryTfStateText}
              onTfStateTextChange={setInventoryTfStateText}
              planJsonText={inventoryPlanJsonText}
              onPlanJsonTextChange={setInventoryPlanJsonText}
              ansibleIniText={inventoryAnsibleIniText}
              onAnsibleIniTextChange={setInventoryAnsibleIniText}
              gitRepoRoot={gitRepoRoot}
              pipelineWorkdir={pipelineWorkdir}
              pipelineAnsibleRoot={pipelineAnsibleRoot}
              pipelinePlanFile={pipelinePlanFile}
              pipelineStateFile={pipelineStateFile}
              repoSession={repoScanSession}
              onOpenRepository={() => void openRepositoryFolder()}
              onClearRepository={() => setRepoScanSession(null)}
              serverSummary={serverSummary}
              onClearServer={() => {
                setServerSummary(null)
                setReconciliationSnapshot(null)
                setServerError(null)
              }}
              onLoadServer={loadFromOmnigraphServer}
              serverLoading={serverLoading}
              serverError={serverError}
              reconciliationSnapshot={reconciliationSnapshot}
              workspaceStreamConnected={workspaceStreamConnected}
              workspaceStreamError={workspaceStreamError}
              serveApiToken={serveApiToken}
              onServeApiTokenChange={setServeApiToken}
            />
          ) : null}
          {activeTab === 'posture' ? (
            <PostureTab
              securityJsonText={postureSecurityJson}
              onSecurityJsonTextChange={setPostureSecurityJson}
              serveApiToken={serveApiToken}
              onServeApiTokenChange={setServeApiToken}
            />
          ) : null}
          {activeTab === 'pipeline' ? (
            <PipelineTab
              graphText={graphText}
              workdir={pipelineWorkdir}
              onWorkdirChange={setPipelineWorkdir}
              ansibleRoot={pipelineAnsibleRoot}
              onAnsibleRootChange={setPipelineAnsibleRoot}
              playbookRel={pipelinePlaybookRel}
              onPlaybookRelChange={setPipelinePlaybookRel}
              playbookOverride={pipelinePlaybookOverride}
              onPlaybookOverrideChange={setPipelinePlaybookOverride}
              usePlaybookOverride={pipelineUsePlaybookOverride}
              onUsePlaybookOverrideChange={setPipelineUsePlaybookOverride}
              schema={pipelineSchema}
              onSchemaChange={setPipelineSchema}
              tfBinary={pipelineTfBinary}
              onTfBinaryChange={setPipelineTfBinary}
              planFile={pipelinePlanFile}
              onPlanFileChange={setPipelinePlanFile}
              stateFile={pipelineStateFile}
              onStateFileChange={setPipelineStateFile}
              runner={pipelineRunner}
              onRunnerChange={setPipelineRunner}
              containerRuntime={pipelineContainerRuntime}
              onContainerRuntimeChange={setPipelineContainerRuntime}
              autoApprove={pipelineAutoApprove}
              onAutoApproveChange={setPipelineAutoApprove}
              skipAnsible={pipelineSkipAnsible}
              onSkipAnsibleChange={setPipelineSkipAnsible}
              graphOut={pipelineGraphOut}
              onGraphOutChange={setPipelineGraphOut}
              telemetryFile={pipelineTelemetryFile}
              onTelemetryFileChange={setPipelineTelemetryFile}
              iacEngine={pipelineIacEngine}
              onIacEngineChange={setPipelineIacEngine}
              tofuImage={pipelineTofuImage}
              onTofuImageChange={setPipelineTofuImage}
              ansibleImage={pipelineAnsibleImage}
              onAnsibleImageChange={setPipelineAnsibleImage}
              pulumiImage={pipelinePulumiImage}
              onPulumiImageChange={setPipelinePulumiImage}
            />
          ) : null}
        </main>

        {wasmNote ? (
          <p className="shrink-0 border-t border-gray-800 px-6 py-2 text-xs text-gray-500" data-testid="wasm-spike-note">
            {wasmNote}
          </p>
        ) : null}
      </div>
    </div>
  )
}
