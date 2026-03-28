import { useCallback, useEffect, useMemo, useRef } from 'react'
import {
  Background,
  Controls,
  Handle,
  MiniMap,
  Panel,
  Position,
  ReactFlow,
  type Edge,
  type Node,
  type NodeProps,
  ReactFlowProvider,
  useEdgesState,
  useNodesState,
} from '@xyflow/react'
import '@xyflow/react/dist/style.css'

import { mapGraphV1ToFlow } from './mapGraphV1ToFlow'
import type { GraphDocument } from './types'

/** Selected React Flow node summary for parent inspectors. */
export type GraphNodeSelection = {
  id: string
  label: string
  kind: string
  state: string
  subtitle: string
  /** Imperative runner lines from graph attributes.debugLog (contextual debugging). */
  debugLog: string[]
}

function OmniNode({ data }: NodeProps) {
  const kind = String(data.kind ?? '')
  const state = String(data.state ?? '')
  const isGray = state === 'gray'
  const border =
    kind === 'project'
      ? 'border-emerald-600/60'
      : kind === 'host'
        ? 'border-amber-600/60'
        : 'border-slate-500/60'
  const grayRing = isGray ? 'ring-1 ring-slate-600/40 border-dashed opacity-90' : ''
  return (
    <div
      className={`min-w-[168px] rounded-lg border px-3 py-2 text-left text-xs ${border} ${grayRing} bg-slate-900/95 text-slate-100 shadow-md`}
    >
      <Handle type="target" position={Position.Top} className="!h-2 !w-2 !bg-slate-400" />
      <div className="font-medium">{String(data.label)}</div>
      {data.subtitle ? (
        <div className="mt-0.5 font-mono text-[10px] text-slate-400">{String(data.subtitle)}</div>
      ) : null}
      <div className="mt-1 flex items-center gap-1 text-[10px] uppercase tracking-wide text-slate-500">
        {kind}
        {state ? <span className="rounded bg-slate-800/80 px-1 font-normal normal-case text-slate-400">{state}</span> : null}
      </div>
      <Handle type="source" position={Position.Bottom} className="!h-2 !w-2 !bg-slate-400" />
    </div>
  )
}

/** Telemetry / CMDB context: SVG accent + muted chrome (React Flow depth; no parallel D3 engine). */
function TelemetryNode({ data }: NodeProps) {
  const kind = String(data.kind ?? 'telemetry')
  const state = String(data.state ?? 'gray')
  return (
    <div className="relative min-w-[188px] rounded-lg border border-dashed border-sky-500/35 bg-slate-900/80 px-3 py-2 pl-11 text-left text-xs text-slate-100 shadow-md ring-1 ring-sky-500/15">
      <svg
        className="pointer-events-none absolute left-2 top-1/2 h-9 w-9 -translate-y-1/2 text-sky-400/85"
        viewBox="0 0 36 36"
        aria-hidden
      >
        <circle cx="18" cy="18" r="2.8" fill="currentColor" />
        <path
          d="M6 18c4-7 8-7 12 0s8 7 12 0"
          fill="none"
          stroke="currentColor"
          strokeWidth="1.3"
          strokeLinecap="round"
          opacity="0.55"
        />
        <path
          d="M6 24c4-6 8-6 12 0s8 6 12 0"
          fill="none"
          stroke="currentColor"
          strokeWidth="1.1"
          strokeLinecap="round"
          opacity="0.35"
        />
      </svg>
      <Handle type="target" position={Position.Top} className="!h-2 !w-2 !bg-sky-400/80" />
      <div className="font-medium text-sky-100">{String(data.label)}</div>
      {data.subtitle ? (
        <div className="mt-0.5 font-mono text-[10px] text-sky-200/70">{String(data.subtitle)}</div>
      ) : null}
      <div className="mt-1 flex items-center gap-1 text-[10px] uppercase tracking-wide text-sky-300/60">
        {kind}
        {state ? (
          <span className="rounded bg-sky-950/50 px-1 font-normal normal-case text-sky-200/80">{state}</span>
        ) : null}
      </div>
      <Handle type="source" position={Position.Bottom} className="!h-2 !w-2 !bg-sky-400/80" />
    </div>
  )
}

const nodeTypes = {
  project: OmniNode,
  tool: OmniNode,
  host: OmniNode,
  telemetry: TelemetryNode,
  default: OmniNode,
}

function parseGraph(text: string): { ok: true; doc: GraphDocument } | { ok: false; error: string } {
  const t = text.trim()
  if (!t) {
    return { ok: false, error: 'Empty graph JSON' }
  }
  try {
    const doc = JSON.parse(t) as GraphDocument
    if (doc.apiVersion !== 'omnigraph/graph/v1' || doc.kind !== 'Graph') {
      return { ok: false, error: 'Not an omnigraph/graph/v1 document' }
    }
    if (!doc.spec?.nodes || !doc.spec?.edges) {
      return { ok: false, error: 'Missing spec.nodes or spec.edges' }
    }
    return { ok: true, doc }
  } catch (e) {
    const m = e instanceof Error ? e.message : String(e)
    return { ok: false, error: m }
  }
}

function GraphCanvasInner({
  graphText,
  onNodeSelect,
}: {
  graphText: string
  onNodeSelect?: (node: GraphNodeSelection | null) => void
}) {
  const [nodes, setNodes, onNodesChange] = useNodesState<Node>([])
  const [edges, setEdges, onEdgesChange] = useEdgesState<Edge>([])
  const onSelectRef = useRef(onNodeSelect)
  useEffect(() => {
    onSelectRef.current = onNodeSelect
  }, [onNodeSelect])

  const parsed = useMemo(() => parseGraph(graphText), [graphText])

  const applyLayout = useCallback(
    (doc: GraphDocument) => {
      try {
        const { nodes: n, edges: e } = mapGraphV1ToFlow(doc)
        setNodes(n)
        setEdges(e)
      } catch {
        setNodes([])
        setEdges([])
      }
    },
    [setEdges, setNodes],
  )

  useEffect(() => {
    if (parsed.ok) {
      applyLayout(parsed.doc)
    } else {
      setNodes([])
      setEdges([])
    }
    onSelectRef.current?.(null)
  }, [parsed, applyLayout, setEdges, setNodes])

  if (!parsed.ok) {
    return (
      <div className="flex h-96 w-full items-center justify-center rounded-xl border border-rose-900/50 bg-rose-950/20 p-4 text-sm text-rose-200">
        {parsed.error}
      </div>
    )
  }

  return (
    <div className="h-[420px] w-full rounded-xl border border-slate-700 bg-slate-950">
      <ReactFlow
        nodes={nodes}
        edges={edges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        nodeTypes={nodeTypes}
        onNodeClick={(_, n) => {
          const d = n.data as Record<string, unknown>
          const rawLog = d.debugLog
          let debugLog: string[] = []
          if (Array.isArray(rawLog)) {
            debugLog = rawLog.filter((x): x is string => typeof x === 'string')
          }
          onSelectRef.current?.({
            id: n.id,
            label: String(d.label ?? ''),
            kind: String(d.kind ?? ''),
            state: String(d.state ?? ''),
            subtitle: String(d.subtitle ?? ''),
            debugLog,
          })
        }}
        onPaneClick={() => onSelectRef.current?.(null)}
        fitView
        fitViewOptions={{ padding: 0.2 }}
        proOptions={{ hideAttribution: true }}
      >
        <Background gap={16} color="#334155" />
        <Controls className="!bg-slate-800 !border-slate-600" />
        <MiniMap
          className="!bg-slate-900 !border-slate-600"
          nodeStrokeWidth={2}
          maskColor="rgb(15 23 42 / 0.7)"
        />
        <Panel position="top-left" className="rounded bg-slate-900/90 px-2 py-1 text-xs text-slate-400">
          phase: {parsed.doc.spec.phase}
          {parsed.doc.metadata.project ? ` · ${parsed.doc.metadata.project}` : ''}
        </Panel>
      </ReactFlow>
    </div>
  )
}

export function GraphCanvas({
  graphText,
  onNodeSelect,
}: {
  graphText: string
  onNodeSelect?: (node: GraphNodeSelection | null) => void
}) {
  return (
    <ReactFlowProvider>
      <GraphCanvasInner graphText={graphText} onNodeSelect={onNodeSelect} />
    </ReactFlowProvider>
  )
}
