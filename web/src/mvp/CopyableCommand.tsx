import { Check, Copy } from 'lucide-react'
import { useState } from 'react'

export type CopyableCommandProps = {
  label: string
  command: string
}

export function CopyableCommand({ label, command }: CopyableCommandProps) {
  const [copied, setCopied] = useState(false)

  const copy = async () => {
    try {
      await navigator.clipboard.writeText(command)
      setCopied(true)
      window.setTimeout(() => setCopied(false), 2000)
    } catch {
      setCopied(false)
    }
  }

  return (
    <div className="rounded-lg border border-gray-800 bg-gray-950">
      <div className="flex items-center justify-between gap-2 border-b border-gray-800 px-3 py-2">
        <span className="text-xs font-medium text-gray-400">{label}</span>
        <button
          type="button"
          onClick={() => void copy()}
          className="flex items-center gap-1 rounded border border-gray-700 bg-gray-900 px-2 py-1 text-xs text-gray-300 hover:bg-gray-800"
        >
          {copied ? <Check size={14} className="text-emerald-400" aria-hidden /> : <Copy size={14} aria-hidden />}
          {copied ? 'Copied' : 'Copy'}
        </button>
      </div>
      <pre className="overflow-x-auto p-3 font-mono text-xs leading-relaxed text-gray-200">{command}</pre>
    </div>
  )
}
