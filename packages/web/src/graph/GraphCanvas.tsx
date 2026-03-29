import { Bot, Maximize2, Minimize2, Radio } from 'lucide-react'
import { useCallback, useEffect, useMemo, useRef, useState, type RefObject } from 'react'
import {
  Background,
  Controls,
  Handle,
  MiniMap,
  Panel,
  Position,
  ReactFlow,
  ViewportPortal,
  type Edge,
  type Node,
  type NodeProps,
  ReactFlowProvider,
  useEdgesState,
  useNodesState,
  useReactFlow,
  useStore,
} from '@xyflow/react'
import '@xyflow/react/dist/style.css'

import { clusterPaletteClass, computeEnclaveClusters } from './graphConventions'
import { mapGraphV1ToFlow, type DrillHighlightKind } from './mapGraphV1ToFlow'
import { parseGraphDocument } from './parseGraphDocument'
import type { GraphDocument } from './types'
import { TopologyNodeFrame } from '../triage/TopologyNode'

/** Selected React Flow node summary for parent inspectors. */
export type GraphNodeSelection = {
  id: string
  label: string
  kind: string
  state: string
  subtitle: string
  /** Imperative runner lines from graph attributes.debugLog (contextual debugging). */
  debugLog: string[]
  /** Enclave / trust zone label when `attributes.enclave` or `trustZone` is set. */
  enclave: string
  /** Copy of graph node attributes for mesh telemetry and future fields. */
  attributes: Record<string, unknown>
}

function drillHighlightRingClass(h: unknown): string {
  if (typeof h !== 'string' || !h) {
    return ''
  }
  switch (h) {
    case 'incident':
      return 'ring-2 ring-rose-500/90 shadow-[0_0_22px_rgba(244,63,94,0.4)]'
    case 'downstream':
      return 'ring-2 ring-orange-500/80 shadow-[0_0_16px_rgba(249,115,22,0.25)]'
    case 'upstream':
      return 'ring-2 ring-cyan-500/65 shadow-[0_0_14px_rgba(34,211,238,0.2)]'
    default:
      return ''
  }
}

function statusRingClass(state: string): string {
  switch (state) {
    case 'active':
    case 'live':
      return 'ring-2 ring-emerald-500/75'
    case 'degraded':
    case 'partial':
      return 'ring-2 ring-amber-500/75'
    case 'gray':
      return 'ring-1 ring-slate-500/45 border-dashed'
    case 'pending':
    case 'planned':
      return 'ring-2 ring-blue-500/55 animate-pulse'
    default:
      return 'ring-1 ring-violet-500/40'
  }
}

function OmniNode({ data, id }: NodeProps) {
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
  const drill = drillHighlightRingClass(data.drillHighlight)
  return (
    <TopologyNodeFrame nodeId={id}>
    <div
      className={`min-w-[168px] rounded-lg border px-3 py-2 text-left text-xs ${border} ${grayRing} ${drill} bg-slate-900/95 text-slate-100 shadow-md`}
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
    </TopologyNodeFrame>
  )
}

/** Telemetry / CMDB context: SVG accent + muted chrome (React Flow depth; no parallel D3 engine). */
function TelemetryNode({ data, id }: NodeProps) {
  const kind = String(data.kind ?? 'telemetry')
  const state = String(data.state ?? 'gray')
  const drill = drillHighlightRingClass(data.drillHighlight)
  return (
    <TopologyNodeFrame nodeId={id}>
    <div className={`relative min-w-[188px] rounded-lg border border-dashed border-sky-500/35 bg-slate-900/80 px-3 py-2 pl-11 text-left text-xs text-slate-100 shadow-md ring-1 ring-sky-500/15 ${drill}`}>
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
    </TopologyNodeFrame>
  )
}

function MeshBrokerNode({ data, id }: NodeProps) {
  const state = String(data.state ?? '')
  const ring = statusRingClass(state)
  const drill = drillHighlightRingClass(data.drillHighlight)
  return (
    <TopologyNodeFrame nodeId={id}>
    <div
      className={`relative min-w-[180px] rounded-lg border border-violet-500/45 bg-slate-900/95 px-3 py-2 pl-10 text-left text-xs text-slate-100 shadow-md ${ring} ${drill}`}
    >
      <Radio
        className="pointer-events-none absolute left-2.5 top-1/2 h-5 w-5 -translate-y-1/2 text-violet-400/90"
        aria-hidden
      />
      <Handle type="target" position={Position.Top} className="!h-2 !w-2 !bg-violet-400" />
      <div className="font-medium text-violet-100">{String(data.label)}</div>
      {data.subtitle ? (
        <div className="mt-0.5 font-mono text-[10px] text-violet-200/65">{String(data.subtitle)}</div>
      ) : null}
      <div className="mt-1 flex flex-wrap items-center gap-1 text-[10px] uppercase tracking-wide text-violet-300/55">
        <span className="rounded bg-violet-950/55 px-1 text-[9px] font-semibold text-violet-200/90">broker</span>
        {state ? <span className="font-normal normal-case text-violet-200/80">{state}</span> : null}
      </div>
      <Handle type="source" position={Position.Bottom} className="!h-2 !w-2 !bg-violet-400" />
    </div>
    </TopologyNodeFrame>
  )
}

function MeshAgentNode({ data, id }: NodeProps) {
  const state = String(data.state ?? '')
  const ring = statusRingClass(state)
  const drill = drillHighlightRingClass(data.drillHighlight)
  return (
    <TopologyNodeFrame nodeId={id}>
    <div
      className={`relative min-w-[180px] rounded-lg border border-indigo-500/45 bg-slate-900/95 px-3 py-2 pl-10 text-left text-xs text-slate-100 shadow-md ${ring} ${drill}`}
    >
      <Bot
        className="pointer-events-none absolute left-2.5 top-1/2 h-5 w-5 -translate-y-1/2 text-indigo-400/90"
        aria-hidden
      />
      <Handle type="target" position={Position.Top} className="!h-2 !w-2 !bg-indigo-400" />
      <div className="font-medium text-indigo-100">{String(data.label)}</div>
      {data.subtitle ? (
        <div className="mt-0.5 font-mono text-[10px] text-indigo-200/65">{String(data.subtitle)}</div>
      ) : null}
      <div className="mt-1 flex flex-wrap items-center gap-1 text-[10px] uppercase tracking-wide text-indigo-300/55">
        <span className="rounded bg-indigo-950/55 px-1 text-[9px] font-semibold text-indigo-200/90">agent</span>
        {state ? <span className="font-normal normal-case text-indigo-200/80">{state}</span> : null}
      </div>
      <Handle type="source" position={Position.Bottom} className="!h-2 !w-2 !bg-indigo-400" />
    </div>
    </TopologyNodeFrame>
  )
}

const nodeTypes = {
  project: OmniNode,
  tool: OmniNode,
  host: OmniNode,
  telemetry: TelemetryNode,
  broker: MeshBrokerNode,
  agent: MeshAgentNode,
  default: OmniNode,
}

function EnclaveClusterLayer() {
  const nodes = useStore((s) => s.nodes)
  const clusters = useMemo(() => computeEnclaveClusters(nodes), [nodes])

  return (
    <ViewportPortal>
      <svg
        className="pointer-events-none"
        style={{
          position: 'absolute',
          left: 0,
          top: 0,
          width: '100%',
          height: '100%',
          overflow: 'visible',
          zIndex: -1,
        }}
        aria-hidden
      >
        {clusters.map((c, i) => {
          const p = clusterPaletteClass(i)
          const w = c.maxX - c.minX
          const h = c.maxY - c.minY
          return (
            <g key={c.id}>
              <rect
                x={c.minX}
                y={c.minY}
                width={w}
                height={h}
                rx={14}
                fill={p.fill}
                stroke={p.stroke}
                strokeWidth={1}
              />
              <text
                x={c.minX + 10}
                y={c.minY + 15}
                fill={p.text}
                fontSize={11}
                fontFamily="ui-sans-serif, system-ui, sans-serif"
                fontWeight={600}
              >
                {c.label}
              </text>
            </g>
          )
        })}
      </svg>
    </ViewportPortal>
  )
}

function GraphViewportControls({ containerRef }: { containerRef: RefObject<HTMLDivElement | null> }) {
  const rf = useReactFlow()
  const [fullscreen, setFullscreen] = useState(false)

  const scheduleFitView = useCallback(() => {
    requestAnimationFrame(() => {
      rf.fitView({ padding: 0.2, duration: 200 })
    })
  }, [rf])

  useEffect(() => {
    const onFullscreenChange = () => {
      setFullscreen(document.fullscreenElement === containerRef.current)
      scheduleFitView()
    }
    document.addEventListener('fullscreenchange', onFullscreenChange)
    return () => document.removeEventListener('fullscreenchange', onFullscreenChange)
  }, [containerRef, scheduleFitView])

  useEffect(() => {
    const el = containerRef.current
    if (!el) {
      return
    }
    let timeoutId = 0
    const ro = new ResizeObserver(() => {
      window.clearTimeout(timeoutId)
      timeoutId = window.setTimeout(() => scheduleFitView(), 120)
    })
    ro.observe(el)
    return () => {
      window.clearTimeout(timeoutId)
      ro.disconnect()
    }
  }, [containerRef, scheduleFitView])

  const toggleFullscreen = () => {
    const el = containerRef.current
    if (!el) {
      return
    }
    const fsApi = el.requestFullscreen?.bind(el)
    if (!document.fullscreenElement) {
      if (fsApi) {
        void fsApi().catch(() => {})
      }
    } else {
      void document.exitFullscreen?.().catch(() => {})
    }
  }

  const fsSupported = typeof document !== 'undefined' && Boolean(document.documentElement.requestFullscreen)

  return (
    <Panel position="top-right" className="m-0 flex gap-1">
      {fsSupported ? (
        <button
          type="button"
          onClick={toggleFullscreen}
          className="flex items-center gap-1.5 rounded border border-slate-600 bg-slate-800/95 px-2 py-1.5 text-[11px] text-slate-200 shadow-sm hover:bg-slate-700"
          title={fullscreen ? 'Exit fullscreen' : 'Fullscreen graph on this display'}
        >
          {fullscreen ? <Minimize2 size={14} aria-hidden /> : <Maximize2 size={14} aria-hidden />}
          <span className="hidden sm:inline">{fullscreen ? 'Exit' : 'Fullscreen'}</span>
        </button>
      ) : null}
    </Panel>
  )
}

function selectionFromNodeData(n: Node): GraphNodeSelection {
  const d = n.data as Record<string, unknown>
  const rawLog = d.debugLog
  let debugLog: string[] = []
  if (Array.isArray(rawLog)) {
    debugLog = rawLog.filter((x): x is string => typeof x === 'string')
  }
  const rawAttr = d.attributes
  const attributes =
    rawAttr && typeof rawAttr === 'object' && !Array.isArray(rawAttr)
      ? (rawAttr as Record<string, unknown>)
      : {}
  return {
    id: n.id,
    label: String(d.label ?? ''),
    kind: String(d.kind ?? ''),
    state: String(d.state ?? ''),
    subtitle: String(d.subtitle ?? ''),
    debugLog,
    enclave: String(d.enclave ?? ''),
    attributes,
  }
}

function GraphCanvasInner({
  graphText,
  onNodeSelect,
  className,
  drillHighlight,
}: {
  graphText: string
  onNodeSelect?: (node: GraphNodeSelection | null) => void
  className?: string
  drillHighlight?: Record<string, DrillHighlightKind>
}) {
  const wrapRef = useRef<HTMLDivElement>(null)
  const [nodes, setNodes, onNodesChange] = useNodesState<Node>([])
  const [edges, setEdges, onEdgesChange] = useEdgesState<Edge>([])
  const onSelectRef = useRef(onNodeSelect)
  useEffect(() => {
    onSelectRef.current = onNodeSelect
  }, [onNodeSelect])

  const parsed = useMemo(() => parseGraphDocument(graphText), [graphText])

  const applyLayout = useCallback(
    (doc: GraphDocument) => {
      try {
        const { nodes: n, edges: e } = mapGraphV1ToFlow(doc, { drillHighlight })
        setNodes(n)
        setEdges(e)
      } catch {
        setNodes([])
        setEdges([])
      }
    },
    [drillHighlight, setEdges, setNodes],
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
      <div
        className={`flex min-h-[280px] w-full flex-1 items-center justify-center rounded-xl border border-rose-900/50 bg-rose-950/20 p-4 text-sm text-rose-200 ${className ?? ''}`}
      >
        {parsed.error}
      </div>
    )
  }

  return (
    <div
      ref={wrapRef}
      className={`flex h-full min-h-[280px] w-full flex-col rounded-xl border border-slate-700 bg-slate-950 ${className ?? ''}`}
    >
      <ReactFlow
        className="!h-full min-h-0 flex-1"
        style={{ height: '100%' }}
        nodes={nodes}
        edges={edges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        nodeTypes={nodeTypes}
        onNodeClick={(_, n) => {
          onSelectRef.current?.(selectionFromNodeData(n))
        }}
        onPaneClick={() => onSelectRef.current?.(null)}
        fitView
        fitViewOptions={{ padding: 0.2 }}
        proOptions={{ hideAttribution: true }}
      >
        <GraphViewportControls containerRef={wrapRef} />
        <EnclaveClusterLayer />
        <Background gap={16} color="#334155" />
        <Controls className="!bg-slate-800 !border-slate-600" />
        <MiniMap
          className="!bg-slate-900 !border-slate-600"
          nodeStrokeWidth={2}
          maskColor="rgb(15 23 42 / 0.7)"
        />
        <Panel position="top-left" className="max-w-[min(100%,280px)] rounded bg-slate-900/90 px-2 py-1.5 text-xs text-slate-400">
          <div>
            phase: {parsed.doc.spec.phase}
            {parsed.doc.metadata.project ? ` · ${parsed.doc.metadata.project}` : ''}
          </div>
          <div className="mt-0.5 text-[10px] leading-snug text-slate-500">
            Shaded regions = enclave / trust zone (<code className="text-slate-500">attributes.enclave</code> or{' '}
            <code className="text-slate-500">trustZone</code>). Unlabeled group = Unscoped.
          </div>
        </Panel>
      </ReactFlow>
    </div>
  )
}

export function GraphCanvas({
  graphText,
  onNodeSelect,
  className,
  drillHighlight,
}: {
  graphText: string
  onNodeSelect?: (node: GraphNodeSelection | null) => void
  /** Merged onto the graph container (e.g. `min-h-0 flex-1` in popout layout). */
  className?: string
  /** Fire-drill / blast-radius overlay kinds keyed by graph node id. */
  drillHighlight?: Record<string, DrillHighlightKind>
}) {
  return (
    <ReactFlowProvider>
      <GraphCanvasInner
        graphText={graphText}
        onNodeSelect={onNodeSelect}
        className={className}
        drillHighlight={drillHighlight}
      />
    </ReactFlowProvider>
  )
}

export type { DrillHighlightKind }
