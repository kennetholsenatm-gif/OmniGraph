import type { ReactNode } from 'react'

export type CodeBlockProps = {
  title: string
  icon: ReactNode
  children: string
}

export function CodeBlock({ title, icon, children }: CodeBlockProps) {
  return (
    <div className="overflow-hidden rounded-xl border border-gray-800 bg-gray-950 shadow-lg">
      <div className="flex items-center gap-2 border-b border-gray-800 bg-gray-900 px-4 py-2 text-sm font-medium text-gray-300">
        {icon}
        {title}
      </div>
      <pre className="overflow-x-auto p-4 font-mono text-sm leading-relaxed text-gray-300">{children}</pre>
    </div>
  )
}
