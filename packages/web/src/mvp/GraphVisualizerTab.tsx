import { Download, Eye, Monitor, Network, Siren, Upload } from 'lucide-react'
import { useEffect, useRef, type ChangeEvent } from 'react'

import { GraphCanvas, type GraphNodeSelection } from '../graph/GraphCanvas'
import { buildReconciliationViewModel } from './buildReconciliationViewModel'
import type { ReconciliationSnapshot } from './omnigraphApi'
import { NodeContextPanel } from '../triage/NodeContextPanel'
import { TriageSessionProvider, postTriageSelectionDetached, useTriageSession } from '../triage/TriageSessionContext'
import { createGraphPopoutChannel, postGraphToPopouts } from './graphPopoutChannel'
import { GRAPH_V1_ATTR } from '../graph/graphConventions'

function formatAttrSnippet(value: unknown): string {
  if (value === undefined || value === null) {
    return ''
  }
  if (typeof value === 'string') {
    return value
  }
  try {
    return JSON.stringify(value, null, 2)
  } catch {
    return String(value)
  }
}

function isAgentMeshNode(kind: string, attr: Record<string, unknown>): boolean {
  if (kind === 'broker' || kind === 'agent') {
    return true
  }
  const r = attr[GRAPH_V1_ATTR.meshRole]
  return r === 'broker' || r === 'agent'
}

export type GraphVisualizerTabProps = {
  graphText: string
  onGraphTextChange: (value: string) => void
  selectedNode: GraphNodeSelection | null
  onNodeSelect: (node: GraphNodeSelection | null) => void
  reconciliationSnapshot: ReconciliationSnapshot | null
  graphFileNameHint?: string
  onGraphFileNameHintChange?: (value: string | undefined) => void
}

function GraphVisualizerTabInner({
  graphText,
  onGraphTextChange,
  selectedNode,
  onNodeSelect,
  reconciliationSnapshot,
  graphFileNameHint,
  onGraphFileNameHintChange,
}: GraphVisualizerTabProps) {
  const fileInputRef = useRef<HTMLInputElement>(null)
  const popoutChannelRef = useRef<BroadcastChannel | null>(null)

  useEffect(() => {
    popoutChannelRef.current = createGraphPopoutChannel()
    return () => {
      popoutChannelRef.current?.close()
      popoutChannelRef.current = null
    }
  }, [])

  useEffect(() => {
    postGraphToPopouts(popoutChannelRef.current, graphText)
  }, [graphText])

  const {
    triageByNodeId,
    triageModeEnabled,
    setTriageModeEnabled,
    panelDetached,
    setPanelDetached,
    streamStatus,
  } = useTriageSession()

  useEffect(() => {
    if (panelDetached) {
      postTriageSelectionDetached(selectedNode?.id ?? null)
    }
  }, [selectedNode?.id, panelDetached])

  const openGraphPopout = () => {
    const u = new URL(window.location.href)
    u.searchParams.set('popout', 'graph')
    window.open(u.toString(), '_blank', 'noopener,noreferrer,width=1280,height=800')
  }

  const openTriagePanelPopout = () => {
    setPanelDetached(true)
    const u = new URL(window.location.href)
    u.searchParams.set('popout', 'triage-panel')
    window.open(u.toString(), '_blank', 'noopener,noreferrer,width=480,height=900')
  }

  const onPickFile = () => fileInputRef.current?.click()

  const onFileSelected = (e: ChangeEvent<HTMLInputElement>) => {
    const f = e.target.files?.[0]
    e.target.value = ''
    if (!f) {
      return
    }
    const reader = new FileReader()
    reader.onload = () => {
      const text = typeof reader.result === 'string' ? reader.result : ''
      onGraphTextChange(text)
      onGraphFileNameHintChange?.(f.name)
    }
    reader.readAsText(f)
  }

  const triageForSelected = selectedNode ? triageByNodeId[selectedNode.id] : undefined
  const reconVm = buildReconciliationViewModel(reconciliationSnapshot)

  const downloadGraph = () => {
    const blob = new Blob([graphText], { type: 'application/json;charset=utf-8' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = graphFileNameHint?.endsWith('.json') ? graphFileNameHint : 'graph.v1.json'
    a.click()
    URL.revokeObjectURL(url)
  }

  return (
    <div className="flex h-full min-h-0 w-full flex-col gap-4 p-4 lg:flex-row lg:gap-0">
      <div className="flex min-h-0 min-w-0 flex-1 flex-col gap-3 lg:pr-4">
        <div className="flex flex-wrap items-center justify-between gap-2">
          <label htmlFor="mvp-graph-json" className="text-sm font-medium text-gray-300">
            Graph JSON <span className="text-gray-500">(omnigraph/graph/v1)</span>
          </label>
          <div className="flex flex-wrap gap-2">
            <button
              type="button"
              onClick={() => setTriageModeEnabled(!triageModeEnabled)}
              className={`flex items-center gap-2 rounded-lg border px-3 py-1.5 text-xs ${
                triageModeEnabled
                  ? 'border-amber-600/60 bg-amber-950/40 text-amber-100'
                  : 'border-gray-700 bg-gray-900 text-gray-200 hover:bg-gray-800'
              }`}
              title="Unified triage: canvas cues + aggregated hand-off, posture, drift for the selected node id"
            >
              <Siren size={14} aria-hidden />
              Triage mode
            </button>
            <input
              ref={fileInputRef}
              type="file"
              accept=".json,application/json,text/plain"
              className="hidden"
              aria-hidden
              onChange={onFileSelected}
            />
            <button
              type="button"
              onClick={onPickFile}
              className="flex items-center gap-2 rounded-lg border border-gray-700 bg-gray-900 px-3 py-1.5 text-xs text-gray-200 hover:bg-gray-800"
            >
              <Upload size={14} aria-hidden />
              Open
            </button>
            <button
              type="button"
              onClick={downloadGraph}
              className="flex items-center gap-2 rounded-lg border border-gray-700 bg-gray-900 px-3 py-1.5 text-xs text-gray-200 hover:bg-gray-800"
            >
              <Download size={14} aria-hidden />
              Download
            </button>
            <button
              type="button"
              onClick={openGraphPopout}
              className="flex items-center gap-2 rounded-lg border border-gray-700 bg-gray-900 px-3 py-1.5 text-xs text-gray-200 hover:bg-gray-800"
              title="Opens a second window you can place on another monitor; stays in sync with this graph JSON."
            >
              <Monitor size={14} aria-hidden />
              Graph in new window
            </button>
          </div>
        </div>
        {graphFileNameHint ? (
          <p className="font-mono text-xs text-gray-500">Loaded: {graphFileNameHint}</p>
        ) : null}
        <textarea
          id="mvp-graph-json"
          spellCheck={false}
          value={graphText}
          onChange={(e) => onGraphTextChange(e.target.value)}
          className={`w-full resize-y rounded-lg border border-gray-800 bg-gray-900/80 p-3 font-mono text-xs text-gray-100 outline-none transition-[max-height,min-height] duration-300 focus:ring-2 focus:ring-blue-500/40 ${
            selectedNode
              ? 'max-h-36 min-h-0 shrink-0 lg:max-h-28'
              : 'min-h-32 min-w-0 flex-1'
          }`}
          aria-label="Omnigraph graph v1 JSON"
        />
        <p className="text-xs text-gray-500">
          Use <strong className="text-gray-400">dependencyRole</strong> on edges (<code className="text-gray-500">necessary</code> vs{' '}
          <code className="text-gray-500">sufficient</code>) for blast-radius semantics — see{' '}
          <span className="text-gray-400">docs/guides/graph-dependencies-and-blast-radius.md</span> (including lifecycle, auditing,
          and roadmap UI notes). Provider artifacts to Inventory/Topology:{' '}
          <span className="text-gray-400">docs/core-concepts/data-handoff.md</span>. Automation that emits graph JSON:{' '}
          <span className="text-gray-400">docs/ci-and-contributor-automation.md</span>.
        </p>
        {reconciliationSnapshot ? (
          <p className="text-[11px] text-gray-500">
            Reconciliation snapshot: {reconVm.bom.totalEntities} BOM entities, {reconVm.bom.totalRelations} dependency
            relations, {reconVm.relationDriftCount} relation drift cues.
          </p>
        ) : null}
        <div className="flex min-h-0 flex-1 flex-col">
          <GraphCanvas graphText={graphText} onNodeSelect={onNodeSelect} className="min-h-0 flex-1" />
        </div>
      </div>

      <aside
        className={`flex min-h-0 shrink-0 flex-col border-t border-gray-800 bg-gray-900/80 backdrop-blur-md lg:border-l lg:border-t-0 ${
          triageModeEnabled ? 'lg:w-[28rem]' : 'lg:w-80'
        } w-full`}
      >
        {triageModeEnabled ? (
          panelDetached ? (
            <div className="flex flex-col gap-3 p-6 text-sm text-gray-400">
              <p>Triage panel is open in another window (BroadcastChannel sync).</p>
              <button
                type="button"
                className="rounded-lg border border-gray-700 bg-gray-900 px-3 py-2 text-xs text-gray-200 hover:bg-gray-800"
                onClick={() => setPanelDetached(false)}
              >
                Re-attach panel here
              </button>
            </div>
          ) : (
            <NodeContextPanel
              open
              selectedNode={selectedNode}
              triage={triageForSelected}
              streamStatus={streamStatus}
              onClose={() => setTriageModeEnabled(false)}
              onDetach={openTriagePanelPopout}
            />
          )
        ) : (
          <div className="flex min-h-0 flex-1 flex-col p-6">
            <h2 className="mb-6 flex items-center gap-2 text-lg font-bold text-gray-100">
              <Eye className="text-blue-400" size={20} aria-hidden />
              Inspector
            </h2>
            <div className="min-h-0 flex-1 overflow-y-auto">
{selectedNode ? (
          <div className="space-y-6 transition-opacity duration-300">
            <div className="flex items-center justify-between gap-2">
              <span className="text-sm text-gray-400">Node id</span>
              <span className="rounded bg-gray-800 px-2 py-1 font-mono text-xs text-gray-200">{selectedNode.id}</span>
            </div>
            <div className="flex items-center justify-between gap-2">
              <span className="text-sm text-gray-400">Label</span>
              <span className="font-mono text-sm text-gray-200">{selectedNode.label}</span>
            </div>
            <div className="flex items-center justify-between gap-2">
              <span className="text-sm text-gray-400">Kind</span>
              <span className="text-xs font-bold uppercase tracking-wider text-blue-400">{selectedNode.kind || '—'}</span>
            </div>
            <div className="flex items-center justify-between gap-2">
              <span className="text-sm text-gray-400">State</span>
              <span className="rounded bg-gray-800 px-2 py-1 text-xs text-gray-300">{selectedNode.state || '—'}</span>
            </div>
            {selectedNode.enclave ? (
              <div className="flex items-center justify-between gap-2">
                <span className="text-sm text-gray-400">Enclave</span>
                <span className="max-w-[160px] truncate rounded bg-indigo-950/50 px-2 py-1 font-mono text-xs text-indigo-200">
                  {selectedNode.enclave}
                </span>
              </div>
            ) : null}
            {isAgentMeshNode(selectedNode.kind, selectedNode.attributes) ? (
              <div className="space-y-3 border-t border-gray-800 pt-4">
                <span className="text-sm text-gray-400">AgentMesh</span>
                <p className="text-[11px] text-gray-500">
                  Telemetry and routing from <code className="text-gray-400">attributes.meshTelemetry</code>,{' '}
                  <code className="text-gray-400">routes</code>, <code className="text-gray-400">connectivity</code>.
                </p>
                {selectedNode.attributes[GRAPH_V1_ATTR.connectivity] != null ? (
                  <div>
                    <div className="mb-1 text-[10px] font-semibold uppercase tracking-wide text-gray-500">Connectivity</div>
                    <div className="rounded border border-violet-900/40 bg-violet-950/20 p-2 font-mono text-xs text-violet-100/90">
                      {formatAttrSnippet(selectedNode.attributes[GRAPH_V1_ATTR.connectivity])}
                    </div>
                  </div>
                ) : null}
                {selectedNode.attributes[GRAPH_V1_ATTR.routes] != null ? (
                  <div>
                    <div className="mb-1 text-[10px] font-semibold uppercase tracking-wide text-gray-500">Routes</div>
                    <pre
                      className="max-h-36 overflow-auto rounded border border-violet-900/40 bg-violet-950/20 p-2 font-mono text-[11px] text-violet-100/85"
                      tabIndex={0}
                    >
                      {formatAttrSnippet(selectedNode.attributes[GRAPH_V1_ATTR.routes])}
                    </pre>
                  </div>
                ) : null}
                {selectedNode.attributes[GRAPH_V1_ATTR.meshTelemetry] != null ? (
                  <div>
                    <div className="mb-1 text-[10px] font-semibold uppercase tracking-wide text-gray-500">Telemetry</div>
                    <pre
                      className="max-h-44 overflow-auto rounded border border-violet-900/40 bg-violet-950/20 p-2 font-mono text-[11px] text-violet-100/85"
                      tabIndex={0}
                    >
                      {formatAttrSnippet(selectedNode.attributes[GRAPH_V1_ATTR.meshTelemetry])}
                    </pre>
                  </div>
                ) : null}
              </div>
            ) : null}
            {selectedNode.subtitle ? (
              <div className="space-y-2 border-t border-gray-800 pt-4">
                <span className="text-sm text-gray-400">Detail</span>
                <div className="rounded border border-gray-800 bg-gray-950 p-3 font-mono text-sm text-gray-300">
                  {selectedNode.subtitle}
                </div>
              </div>
            ) : null}
            {selectedNode.debugLog.length > 0 ? (
              <div className="space-y-2 border-t border-gray-800 pt-4">
                <span className="text-sm text-gray-400">Execution log</span>
                <p className="text-[11px] text-gray-500">
                  Imperative lines mapped to this node via <code className="text-gray-400">attributes.debugLog</code>.
                </p>
                <pre
                  className="max-h-52 overflow-auto rounded border border-amber-900/45 bg-amber-950/25 p-3 font-mono text-[11px] leading-relaxed text-amber-100/90"
                  tabIndex={0}
                >
                  {selectedNode.debugLog.join('\n')}
                </pre>
              </div>
            ) : null}
          </div>
        ) : (
          <div className="mt-12 flex flex-col items-center gap-4 text-center text-gray-500">
            <Network size={48} className="text-gray-700 opacity-50" aria-hidden />
            <p className="text-sm">
              Click a node to inspect fields and optional <code className="text-gray-400">debugLog</code> lines from{' '}
              <code className="text-gray-400">omnigraph/graph/v1</code>.
            </p>
          </div>
        )}
            </div>
          </div>
        )}
      </aside>
    </div>
  )
}

export function GraphVisualizerTab(props: GraphVisualizerTabProps) {
  return (
    <TriageSessionProvider graphText={props.graphText} syncSelectionId={props.selectedNode?.id ?? null}>
      <GraphVisualizerTabInner {...props} />
    </TriageSessionProvider>
  )
}

