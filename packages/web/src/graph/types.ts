/** Matches omnigraph/graph/v1 (see schemas/graph.v1.schema.json). */
export interface GraphDocument {
  apiVersion: string
  kind: string
  metadata: GraphMetadata
  spec: GraphSpec
}

export interface GraphMetadata {
  generatedAt: string
  project?: string
  environment?: string
}

export interface GraphSpec {
  phase: string
  nodes: GraphNodeV1[]
  edges: GraphEdgeV1[]
  phases?: GraphPhaseInfo[]
  summary?: GraphSummary
}

export interface GraphNodeV1 {
  id: string
  kind: string
  label: string
  state?: string
  attributes?: Record<string, unknown>
}

export interface GraphEdgeV1 {
  from: string
  to: string
  kind?: string
  /** necessary (default) | sufficient — see docs/guides/graph-dependencies-and-blast-radius.md */
  dependencyRole?: string
}

export interface GraphPhaseInfo {
  name: string
  status: string
  detail?: string
}

export interface GraphSummary {
  validateOk?: boolean
  coerceOk?: boolean
  inventoryPreview?: string
}
