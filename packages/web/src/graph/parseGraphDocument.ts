import type { GraphDocument } from './types'

export type ParseGraphResult =
  | { ok: true; doc: GraphDocument }
  | { ok: false; error: string }

/** Lightweight client parse for omnigraph/graph/v1 (structural checks only). */
export function parseGraphDocument(text: string): ParseGraphResult {
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
