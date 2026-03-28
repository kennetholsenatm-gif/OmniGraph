import { Download, Eye, Network, Upload } from 'lucide-react'
import { useRef, type ChangeEvent } from 'react'

import { GraphCanvas, type GraphNodeSelection } from '../graph/GraphCanvas'

export type GraphVisualizerTabProps = {
  graphText: string
  onGraphTextChange: (value: string) => void
  selectedNode: GraphNodeSelection | null
  onNodeSelect: (node: GraphNodeSelection | null) => void
  graphFileNameHint?: string
  onGraphFileNameHintChange?: (value: string | undefined) => void
}

export function GraphVisualizerTab({
  graphText,
  onGraphTextChange,
  selectedNode,
  onNodeSelect,
  graphFileNameHint,
  onGraphFileNameHintChange,
}: GraphVisualizerTabProps) {
  const fileInputRef = useRef<HTMLInputElement>(null)

  const onPickFile = () => fileInputRef.current?.click()

  const onFileSelected = (e: ChangeEvent<HTMLInputElement>) => {
    const f = e.target.files?.[0]
    e.target.value = ''
    if (!f) {
      return
    }
    const reader = new FileReader()
    reader.onload = () => {
      const text = typeof reader.result === 'string' ? reader.result : ''
      onGraphTextChange(text)
      onGraphFileNameHintChange?.(f.name)
    }
    reader.readAsText(f)
  }

  const downloadGraph = () => {
    const blob = new Blob([graphText], { type: 'application/json;charset=utf-8' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = graphFileNameHint?.endsWith('.json') ? graphFileNameHint : 'graph.v1.json'
    a.click()
    URL.revokeObjectURL(url)
  }

  return (
    <div className="flex h-full min-h-0 w-full flex-col gap-4 p-4 lg:flex-row lg:gap-0">
      <div className="flex min-h-0 min-w-0 flex-1 flex-col gap-3 lg:pr-4">
        <div className="flex flex-wrap items-center justify-between gap-2">
          <label htmlFor="mvp-graph-json" className="text-sm font-medium text-gray-300">
            Graph JSON <span className="text-gray-500">(omnigraph/graph/v1)</span>
          </label>
          <div className="flex flex-wrap gap-2">
            <input
              ref={fileInputRef}
              type="file"
              accept=".json,application/json,text/plain"
              className="hidden"
              aria-hidden
              onChange={onFileSelected}
            />
            <button
              type="button"
              onClick={onPickFile}
              className="flex items-center gap-2 rounded-lg border border-gray-700 bg-gray-900 px-3 py-1.5 text-xs text-gray-200 hover:bg-gray-800"
            >
              <Upload size={14} aria-hidden />
              Open
            </button>
            <button
              type="button"
              onClick={downloadGraph}
              className="flex items-center gap-2 rounded-lg border border-gray-700 bg-gray-900 px-3 py-1.5 text-xs text-gray-200 hover:bg-gray-800"
            >
              <Download size={14} aria-hidden />
              Download
            </button>
          </div>
        </div>
        {graphFileNameHint ? (
          <p className="font-mono text-xs text-gray-500">Loaded: {graphFileNameHint}</p>
        ) : null}
        <textarea
          id="mvp-graph-json"
          spellCheck={false}
          value={graphText}
          onChange={(e) => onGraphTextChange(e.target.value)}
          className={`w-full resize-y rounded-lg border border-gray-800 bg-gray-900/80 p-3 font-mono text-xs text-gray-100 outline-none transition-[max-height,min-height] duration-300 focus:ring-2 focus:ring-blue-500/40 ${
            selectedNode
              ? 'max-h-36 min-h-0 shrink-0 lg:max-h-28'
              : 'min-h-32 min-w-0 flex-1'
          }`}
          aria-label="Omnigraph graph v1 JSON"
        />
        <p className="text-xs text-gray-500">
          Produce real graphs with <code className="text-gray-400">omnigraph graph emit</code> (optionally{' '}
          <code className="text-gray-400">--plan-json</code>, <code className="text-gray-400">--tfstate</code>,{' '}
          <code className="text-gray-400">--telemetry-file</code>) and paste the output here.
        </p>
        <div className="min-h-[420px] shrink-0">
          <GraphCanvas graphText={graphText} onNodeSelect={onNodeSelect} />
        </div>
      </div>

      <aside className="flex w-full shrink-0 flex-col border-t border-gray-800 bg-gray-900/80 p-6 backdrop-blur-md lg:w-80 lg:border-l lg:border-t-0">
        <h2 className="mb-6 flex items-center gap-2 text-lg font-bold text-gray-100">
          <Eye className="text-blue-400" size={20} aria-hidden />
          Inspector
        </h2>

        {selectedNode ? (
          <div className="space-y-6 transition-opacity duration-300">
            <div className="flex items-center justify-between gap-2">
              <span className="text-sm text-gray-400">Node id</span>
              <span className="rounded bg-gray-800 px-2 py-1 font-mono text-xs text-gray-200">{selectedNode.id}</span>
            </div>
            <div className="flex items-center justify-between gap-2">
              <span className="text-sm text-gray-400">Label</span>
              <span className="font-mono text-sm text-gray-200">{selectedNode.label}</span>
            </div>
            <div className="flex items-center justify-between gap-2">
              <span className="text-sm text-gray-400">Kind</span>
              <span className="text-xs font-bold uppercase tracking-wider text-blue-400">{selectedNode.kind || '—'}</span>
            </div>
            <div className="flex items-center justify-between gap-2">
              <span className="text-sm text-gray-400">State</span>
              <span className="rounded bg-gray-800 px-2 py-1 text-xs text-gray-300">{selectedNode.state || '—'}</span>
            </div>
            {selectedNode.subtitle ? (
              <div className="space-y-2 border-t border-gray-800 pt-4">
                <span className="text-sm text-gray-400">Detail</span>
                <div className="rounded border border-gray-800 bg-gray-950 p-3 font-mono text-sm text-gray-300">
                  {selectedNode.subtitle}
                </div>
              </div>
            ) : null}
            {selectedNode.debugLog.length > 0 ? (
              <div className="space-y-2 border-t border-gray-800 pt-4">
                <span className="text-sm text-gray-400">Execution log</span>
                <p className="text-[11px] text-gray-500">
                  Imperative lines mapped to this node via <code className="text-gray-400">attributes.debugLog</code>.
                </p>
                <pre
                  className="max-h-52 overflow-auto rounded border border-amber-900/45 bg-amber-950/25 p-3 font-mono text-[11px] leading-relaxed text-amber-100/90"
                  tabIndex={0}
                >
                  {selectedNode.debugLog.join('\n')}
                </pre>
              </div>
            ) : null}
          </div>
        ) : (
          <div className="mt-12 flex flex-col items-center gap-4 text-center text-gray-500">
            <Network size={48} className="text-gray-700 opacity-50" aria-hidden />
            <p className="text-sm">
              Click a node to inspect fields and optional <code className="text-gray-400">debugLog</code> lines from{' '}
              <code className="text-gray-400">omnigraph/graph/v1</code>.
            </p>
          </div>
        )}
      </aside>
    </div>
  )
}
