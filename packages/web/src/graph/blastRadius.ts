import type { GraphDocument, GraphEdgeV1 } from './types'

/** Matches Go internal/graph EffectiveDependencyRole (omitted = necessary). */
export function effectiveDependencyRole(e: GraphEdgeV1): 'necessary' | 'sufficient' {
  const s = (e.dependencyRole ?? '').trim()
  if (s === 'sufficient') {
    return 'sufficient'
  }
  return 'necessary'
}

function nodeIDSet(doc: GraphDocument): Set<string> {
  return new Set(doc.spec.nodes.map((n) => n.id))
}

/** Transitive closure along outgoing necessary edges (includes incidents). Sorted ids. */
export function downstreamBlastNodeIds(doc: GraphDocument, incidents: string[]): string[] {
  if (incidents.length === 0) {
    return []
  }
  const ids = nodeIDSet(doc)
  const uniq = [...new Set(incidents)].filter((id) => {
    if (!id || !ids.has(id)) {
      throw new Error(`unknown incident node id: ${id}`)
    }
    return true
  })
  const adj = new Map<string, GraphEdgeV1[]>()
  for (const e of doc.spec.edges) {
    if (effectiveDependencyRole(e) !== 'necessary') {
      continue
    }
    const list = adj.get(e.from) ?? []
    list.push(e)
    adj.set(e.from, list)
  }
  const seen = new Set<string>(uniq)
  const queue = [...uniq]
  while (queue.length > 0) {
    const u = queue.shift()!
    for (const e of adj.get(u) ?? []) {
      if (seen.has(e.to)) {
        continue
      }
      seen.add(e.to)
      queue.push(e.to)
    }
  }
  return [...seen].sort()
}

/** Nodes that can reach incidents via necessary edges backward (includes incidents). Sorted ids. */
export function upstreamBlastNodeIds(doc: GraphDocument, incidents: string[]): string[] {
  if (incidents.length === 0) {
    return []
  }
  const ids = nodeIDSet(doc)
  const uniq = [...new Set(incidents)].filter((id) => {
    if (!id || !ids.has(id)) {
      throw new Error(`unknown incident node id: ${id}`)
    }
    return true
  })
  const rev = new Map<string, GraphEdgeV1[]>()
  for (const e of doc.spec.edges) {
    if (effectiveDependencyRole(e) !== 'necessary') {
      continue
    }
    const list = rev.get(e.to) ?? []
    list.push(e)
    rev.set(e.to, list)
  }
  const seen = new Set<string>(uniq)
  const queue = [...uniq]
  while (queue.length > 0) {
    const u = queue.shift()!
    for (const e of rev.get(u) ?? []) {
      if (seen.has(e.from)) {
        continue
      }
      seen.add(e.from)
      queue.push(e.from)
    }
  }
  return [...seen].sort()
}
