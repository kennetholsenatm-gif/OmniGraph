import dagre from 'dagre'
import type { Edge, Node } from '@xyflow/react'
import type { GraphDocument } from './types'

const nodeWidth = 180
const nodeHeight = 44

function flowNodeType(kind: string): string {
  switch (kind) {
    case 'project':
      return 'project'
    case 'tool':
      return 'tool'
    case 'host':
      return 'host'
    default:
      return 'default'
  }
}

/** Maps omnigraph/graph/v1 nodes and edges into React Flow models with Dagre layout. */
export function mapGraphV1ToFlow(doc: GraphDocument): { nodes: Node[]; edges: Edge[] } {
  const g = new dagre.graphlib.Graph()
  g.setDefaultEdgeLabel(() => ({}))
  g.setGraph({ rankdir: 'TB', nodesep: 48, ranksep: 64, marginx: 24, marginy: 24 })

  const rawNodes: Node[] = doc.spec.nodes.map((n) => ({
    id: n.id,
    type: flowNodeType(n.kind),
    position: { x: 0, y: 0 },
    data: {
      label: n.label,
      kind: n.kind,
      state: n.state ?? '',
      subtitle: n.attributes?.ansible_host != null ? String(n.attributes.ansible_host) : '',
    },
  }))

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
