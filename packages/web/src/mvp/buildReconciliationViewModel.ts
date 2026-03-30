import { buildBomViewModel, type BomViewModel } from './buildBomViewModel'
import type { ReconciliationSnapshot } from './omnigraphApi'

export type ReconciliationViewModel = {
  bom: BomViewModel
  degradedNodeCount: number
  fracturedEdgeCount: number
  relationDriftCount: number
  nextActions: string[]
}

export function buildReconciliationViewModel(snapshot: ReconciliationSnapshot | null): ReconciliationViewModel {
  if (!snapshot) {
    return {
      bom: buildBomViewModel(null),
      degradedNodeCount: 0,
      fracturedEdgeCount: 0,
      relationDriftCount: 0,
      nextActions: [],
    }
  }
  return {
    bom: buildBomViewModel(snapshot.spec.bom),
    degradedNodeCount: snapshot.spec.degradedNodes.length,
    fracturedEdgeCount: snapshot.spec.fracturedEdges.length,
    relationDriftCount: snapshot.spec.relationDrifts.length,
    nextActions: snapshot.spec.nextActions,
  }
}
