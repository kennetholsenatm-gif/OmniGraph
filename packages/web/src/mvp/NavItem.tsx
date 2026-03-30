import type { ReactNode } from 'react'

export type NavItemProps = {
  icon: ReactNode
  label: string
  active: boolean
  onClick: () => void
  /** Indent under a section (e.g. Reconciliation). */
  indent?: boolean
}

export function NavItem({ icon, label, active, onClick, indent }: NavItemProps) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={`group flex w-full items-center gap-3 rounded-lg border py-2.5 transition-all duration-200 ${
        indent ? 'border-transparent pl-5 pr-3' : 'border-transparent px-3'
      } ${
        active
          ? 'border-blue-500/20 bg-blue-600/10 text-blue-400 shadow-inner'
          : 'border-transparent text-gray-400 hover:bg-gray-800 hover:text-gray-200'
      }`}
    >
      <div className={active ? 'text-blue-400' : 'text-gray-400 group-hover:text-gray-200'}>{icon}</div>
      <span className="hidden font-medium text-sm md:block">{label}</span>
    </button>
  )
}
