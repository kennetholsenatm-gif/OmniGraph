import type { GraphDocument } from './types'

/**
 * Optional omnigraph/graph/v1 node `attributes` keys the Topology UI understands.
 * The JSON schema keeps `attributes` open; emitters may populate these over time.
 */
export const GRAPH_V1_ATTR = {
  enclave: 'enclave',
  trustZone: 'trustZone',
  meshRole: 'meshRole',
  meshTelemetry: 'meshTelemetry',
  routes: 'routes',
  connectivity: 'connectivity',
} as const

/** Layout dimensions — keep in sync with Dagre node bounds in mapGraphV1ToFlow. */
export const FLOW_NODE_WIDTH = 200
export const FLOW_NODE_HEIGHT = 52

const UNSCOPED_LABEL = 'Unscoped'

export function enclaveLabelFromAttributes(attr: Record<string, unknown> | undefined): string | null {
  if (!attr) {
    return null
  }
  const e = attr[GRAPH_V1_ATTR.enclave]
  if (typeof e === 'string' && e.trim()) {
    return e.trim()
  }
  const z = attr[GRAPH_V1_ATTR.trustZone]
  if (typeof z === 'string' && z.trim()) {
    return z.trim()
  }
  return null
}

export type MeshRole = 'broker' | 'agent'

export function meshRoleFromNode(kind: string, attr: Record<string, unknown> | undefined): MeshRole | null {
  if (kind === 'broker' || kind === 'agent') {
    return kind
  }
  if (!attr) {
    return null
  }
  const r = attr[GRAPH_V1_ATTR.meshRole]
  if (r === 'broker' || r === 'agent') {
    return r
  }
  return null
}

export function flowTypeForMeshOrKind(kind: string, attr: Record<string, unknown> | undefined): string {
  const mesh = meshRoleFromNode(kind, attr)
  if (mesh === 'broker') {
    return 'broker'
  }
  if (mesh === 'agent') {
    return 'agent'
  }
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

export type EnclaveClusterBounds = {
  id: string
  label: string
  minX: number
  minY: number
  maxX: number
  maxY: number
}

const CLUSTER_PAD = 28
const LABEL_H = 18

/**
 * Group nodes by `data.enclave` (string) or Unscoped when null/empty.
 * Bounds are in flow coordinates including node card size.
 */
export function computeEnclaveClusters(
  nodes: { id: string; position: { x: number; y: number }; data?: Record<string, unknown> }[],
): EnclaveClusterBounds[] {
  const byLabel = new Map<string, string[]>()
  for (const n of nodes) {
    const raw = n.data?.enclave
    const label = typeof raw === 'string' && raw.trim() ? raw.trim() : UNSCOPED_LABEL
    const ids = byLabel.get(label) ?? []
    ids.push(n.id)
    byLabel.set(label, ids)
  }

  const idToPos = new Map(nodes.map((n) => [n.id, n] as const))
  const out: EnclaveClusterBounds[] = []

  for (const [label, ids] of byLabel) {
    let minX = Infinity
    let minY = Infinity
    let maxX = -Infinity
    let maxY = -Infinity
    for (const id of ids) {
      const n = idToPos.get(id)
      if (!n) {
        continue
      }
      const { x, y } = n.position
      minX = Math.min(minX, x)
      minY = Math.min(minY, y)
      maxX = Math.max(maxX, x + FLOW_NODE_WIDTH)
      maxY = Math.max(maxY, y + FLOW_NODE_HEIGHT)
    }
    if (!Number.isFinite(minX)) {
      continue
    }
    out.push({
      id: `enclave-cluster-${label}`,
      label,
      minX: minX - CLUSTER_PAD,
      minY: minY - CLUSTER_PAD - LABEL_H,
      maxX: maxX + CLUSTER_PAD,
      maxY: maxY + CLUSTER_PAD,
    })
  }
  return out
}

export function clusterPaletteClass(index: number): { fill: string; stroke: string; text: string } {
  const palettes = [
    { fill: 'rgba(99, 102, 241, 0.06)', stroke: 'rgba(129, 140, 248, 0.35)', text: '#a5b4fc' },
    { fill: 'rgba(16, 185, 129, 0.06)', stroke: 'rgba(52, 211, 153, 0.35)', text: '#6ee7b7' },
    { fill: 'rgba(245, 158, 11, 0.06)', stroke: 'rgba(251, 191, 36, 0.35)', text: '#fcd34d' },
    { fill: 'rgba(244, 63, 94, 0.06)', stroke: 'rgba(251, 113, 133, 0.35)', text: '#fda4af' },
    { fill: 'rgba(56, 189, 248, 0.06)', stroke: 'rgba(125, 211, 252, 0.35)', text: '#7dd3fc' },
  ]
  return palettes[index % palettes.length]!
}

export function tryParseGraphDocument(text: string): GraphDocument | null {
  const t = text.trim()
  if (!t) {
    return null
  }
  try {
    const doc = JSON.parse(t) as GraphDocument
    if (doc.apiVersion !== 'omnigraph/graph/v1' || doc.kind !== 'Graph') {
      return null
    }
    if (!doc.spec?.nodes || !doc.spec?.edges) {
      return null
    }
    return doc
  } catch {
    return null
  }
}
