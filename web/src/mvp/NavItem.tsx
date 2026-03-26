import type { ReactNode } from 'react'

export type NavItemProps = {
  icon: ReactNode
  label: string
  active: boolean
  onClick: () => void
}

export function NavItem({ icon, label, active, onClick }: NavItemProps) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={`group flex w-full items-center gap-3 rounded-lg border px-3 py-2.5 transition-all duration-200 ${
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
