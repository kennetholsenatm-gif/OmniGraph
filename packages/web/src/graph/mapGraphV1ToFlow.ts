import dagre from 'dagre'
import type { Edge, Node } from '@xyflow/react'
import type { GraphDocument } from './types'

const nodeWidth = 200
const nodeHeight = 52

function flowNodeType(kind: string): string {
  switch (kind) {
    case 'project':
      return 'project'
    case 'tool':
      return 'tool'
    case 'host':
      return 'host'
    case 'telemetry':
      return 'telemetry'
    default:
      return 'default'
  }
}

/** Maps omnigraph/graph/v1 nodes and edges into React Flow models with Dagre layout. */
export function mapGraphV1ToFlow(doc: GraphDocument): { nodes: Node[]; edges: Edge[] } {
  const g = new dagre.graphlib.Graph()
  g.setDefaultEdgeLabel(() => ({}))
  g.setGraph({ rankdir: 'TB', nodesep: 48, ranksep: 64, marginx: 24, marginy: 24 })

  const rawNodes: Node[] = doc.spec.nodes.map((n) => {
    const ansibleHost = n.attributes?.ansible_host
    const src = n.attributes?.source
    const ip = n.attributes?.ip
    let subtitle = ''
    if (ansibleHost != null) {
      subtitle = String(ansibleHost)
    } else if (ip != null) {
      subtitle = String(ip)
    } else if (src != null) {
      subtitle = String(src)
    }
    return {
      id: n.id,
      type: flowNodeType(n.kind),
      position: { x: 0, y: 0 },
      data: {
        label: n.label,
        kind: n.kind,
        state: n.state ?? '',
        subtitle,
        source: src != null ? String(src) : '',
      },
    }
  })

  for (const n of rawNodes) {
    g.setNode(n.id, { width: nodeWidth, height: nodeHeight })
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
      position: { x: pos.x - nodeWidth / 2, y: pos.y - nodeHeight / 2 },
    }
  })

  return { nodes, edges }
}
