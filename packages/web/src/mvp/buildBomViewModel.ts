import type { BomDocument, BomEntityClass, BomRelationType } from './omnigraphApi'

export type BomViewModel = {
  totalEntities: number
  totalRelations: number
  byClass: Record<BomEntityClass, number>
  byRelationType: Record<BomRelationType, number>
}

export function buildBomViewModel(bom: BomDocument | null): BomViewModel {
  const byClass: Record<BomEntityClass, number> = {
    software_component: 0,
    hardware_asset: 0,
    service_endpoint: 0,
  }
  const byRelationType: Record<BomRelationType, number> = {
    depends_on: 0,
    runs_on: 0,
    hosts: 0,
    connects_to: 0,
  }
  if (!bom) {
    return { totalEntities: 0, totalRelations: 0, byClass, byRelationType }
  }
  for (const e of bom.spec.entities ?? []) {
    byClass[e.class] = (byClass[e.class] ?? 0) + 1
  }
  for (const r of bom.spec.relations ?? []) {
    byRelationType[r.type] = (byRelationType[r.type] ?? 0) + 1
  }
  return {
    totalEntities: bom.spec.entities?.length ?? 0,
    totalRelations: bom.spec.relations?.length ?? 0,
    byClass,
    byRelationType,
  }
}
