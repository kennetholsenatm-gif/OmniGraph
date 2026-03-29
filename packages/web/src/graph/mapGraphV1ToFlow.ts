import dagre from 'dagre'
import type { Edge, Node } from '@xyflow/react'
import type { GraphDocument } from './types'

export type DrillHighlightKind = 'incident' | 'downstream' | 'upstream'

export type MapGraphFlowOptions = {
  drillHighlight?: Record<string, DrillHighlightKind>
}
import {
  FLOW_NODE_HEIGHT,
  FLOW_NODE_WIDTH,
  enclaveLabelFromAttributes,
  flowTypeForMeshOrKind,
} from './graphConventions'

export { FLOW_NODE_WIDTH, FLOW_NODE_HEIGHT } from './graphConventions'

function debugLogFromAttributes(attr: Record<string, unknown> | undefined): string[] {
  if (!attr) {
    return []
  }
  const v = attr.debugLog
  if (Array.isArray(v)) {
    return v.filter((x): x is string => typeof x === 'string')
  }
  if (typeof v === 'string' && v.trim()) {
    return v.split(/\n/).map((s) => s.replace(/\r$/, ''))
  }
  return []
}

function cloneAttributes(attr: Record<string, unknown> | undefined): Record<string, unknown> {
  if (!attr) {
    return {}
  }
  return { ...attr }
}

/** Maps omnigraph/graph/v1 nodes and edges into React Flow models with Dagre layout. */
export function mapGraphV1ToFlow(doc: GraphDocument, opts?: MapGraphFlowOptions): { nodes: Node[]; edges: Edge[] } {
  const g = new dagre.graphlib.Graph()
  g.setDefaultEdgeLabel(() => ({}))
  g.setGraph({ rankdir: 'TB', nodesep: 48, ranksep: 64, marginx: 24, marginy: 24 })

  const rawNodes: Node[] = doc.spec.nodes.map((n) => {
    const attr = n.attributes as Record<string, unknown> | undefined
    const ansibleHost = attr?.ansible_host
    const src = attr?.source
    const ip = attr?.ip
    let subtitle = ''
    if (ansibleHost != null) {
      subtitle = String(ansibleHost)
    } else if (ip != null) {
      subtitle = String(ip)
    } else if (src != null) {
      subtitle = String(src)
    }
    const dbg = debugLogFromAttributes(attr)
    const enclave = enclaveLabelFromAttributes(attr)
    const flowType = flowTypeForMeshOrKind(n.kind, attr)
    const drillHighlight = opts?.drillHighlight?.[n.id]
    return {
      id: n.id,
      type: flowType,
      position: { x: 0, y: 0 },
      data: {
        label: n.label,
        kind: n.kind,
        state: n.state ?? '',
        subtitle,
        source: src != null ? String(src) : '',
        debugLog: dbg,
        enclave: enclave ?? '',
        attributes: cloneAttributes(attr),
        drillHighlight,
      },
    }
  })

  for (const n of rawNodes) {
    g.setNode(n.id, { width: FLOW_NODE_WIDTH, height: FLOW_NODE_HEIGHT })
  }

  const edges: Edge[] = doc.spec.edges.map((e, i) => {
    g.setEdge(e.from, e.to)
    return {
      id: `e-${e.from}-${e.to}-${i}`,
      source: e.from,
      target: e.to,
      label: e.kind ?? '',
    }
  })

  dagre.layout(g)

  const nodes = rawNodes.map((n) => {
    const pos = g.node(n.id)
    if (!pos) {
      return n
    }
    return {
      ...n,
      position: { x: pos.x - FLOW_NODE_WIDTH / 2, y: pos.y - FLOW_NODE_HEIGHT / 2 },
    }
  })

  return { nodes, edges }
}
