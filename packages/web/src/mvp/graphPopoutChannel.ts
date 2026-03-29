/** Same-origin graph mirror for a second browser window (e.g. second monitor). */

export const GRAPH_POPOUT_BROADCAST_CHANNEL = 'omnigraph-graph-popout'

export type GraphPopoutMessage = { type: 'graph'; text: string }

export function createGraphPopoutChannel(): BroadcastChannel | null {
  try {
    return new BroadcastChannel(GRAPH_POPOUT_BROADCAST_CHANNEL)
  } catch {
    return null
  }
}

export function postGraphToPopouts(channel: BroadcastChannel | null, text: string): void {
  if (!channel) {
    return
  }
  const msg: GraphPopoutMessage = { type: 'graph', text }
  channel.postMessage(msg)
}

export function isGraphPopoutMessage(data: unknown): data is GraphPopoutMessage {
  return (
    typeof data === 'object' &&
    data !== null &&
    (data as GraphPopoutMessage).type === 'graph' &&
    typeof (data as GraphPopoutMessage).text === 'string'
  )
}
