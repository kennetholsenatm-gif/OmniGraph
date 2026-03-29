import type { NodeTriageState } from './nodeTriageTypes'

export const TRIAGE_PANEL_BROADCAST_CHANNEL = 'omnigraph-triage-panel'

export type TriagePanelBroadcastMessage =
  | {
      type: 'full_sync'
      triageByNodeId: Record<string, NodeTriageState>
      selectedNodeId: string | null
      triageModeEnabled: boolean
    }
  | { type: 'selection'; selectedNodeId: string | null }

export function createTriagePanelChannel(): BroadcastChannel | null {
  try {
    return new BroadcastChannel(TRIAGE_PANEL_BROADCAST_CHANNEL)
  } catch {
    return null
  }
}

export function isTriagePanelMessage(data: unknown): data is TriagePanelBroadcastMessage {
  if (typeof data !== 'object' || data === null) {
    return false
  }
  const t = (data as { type?: string }).type
  return t === 'full_sync' || t === 'selection'
}
