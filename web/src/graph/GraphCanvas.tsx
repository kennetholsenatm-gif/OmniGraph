import { useCallback, useEffect, useMemo } from 'react'
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

function OmniNode({ data }: NodeProps) {
  const kind = String(data.kind ?? '')
  const border =
    kind === 'project'
      ? 'border-emerald-600/60'
      : kind === 'host'
        ? 'border-amber-600/60'
        : 'border-slate-500/60'
  return (
    <div
      className={`min-w-[168px] rounded-lg border px-3 py-2 text-left text-xs ${border} bg-slate-900/95 text-slate-100 shadow-md`}
    >
      <Handle type="target" position={Position.Top} className="!h-2 !w-2 !bg-slate-400" />
      <div className="font-medium">{String(data.label)}</div>
      {data.subtitle ? (
        <div className="mt-0.5 font-mono text-[10px] text-slate-400">{String(data.subtitle)}</div>
      ) : null}
      <div className="mt-1 text-[10px] uppercase tracking-wide text-slate-500">{kind}</div>
      <Handle type="source" position={Position.Bottom} className="!h-2 !w-2 !bg-slate-400" />
    </div>
  )
}

const nodeTypes = {
  project: OmniNode,
  tool: OmniNode,
  host: OmniNode,
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

function GraphCanvasInner({ graphText }: { graphText: string }) {
  const [nodes, setNodes, onNodesChange] = useNodesState<Node>([])
  const [edges, setEdges, onEdgesChange] = useEdgesState<Edge>([])

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

export function GraphCanvas({ graphText }: { graphText: string }) {
  return (
    <ReactFlowProvider>
      <GraphCanvasInner graphText={graphText} />
    </ReactFlowProvider>
  )
}
