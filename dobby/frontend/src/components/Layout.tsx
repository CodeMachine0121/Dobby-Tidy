import { NavLink, Outlet } from 'react-router-dom'
import { LayoutDashboard, ListChecks, ScrollText, Settings } from 'lucide-react'

const navItems = [
  { to: '/', icon: LayoutDashboard, label: '儀表板' },
  { to: '/rules', icon: ListChecks, label: '規則管理' },
  { to: '/logs', icon: ScrollText, label: '操作紀錄' },
  { to: '/settings', icon: Settings, label: '設定' },
]

export function Layout() {
  return (
    <div className="flex h-screen w-screen overflow-hidden">
      {/* Sidebar */}
      <aside className="w-56 flex-shrink-0 bg-sidebar flex flex-col">
        {/* Logo */}
        <div className="h-14 flex items-center px-5 border-b border-slate-800">
          <div className="flex items-center gap-2.5">
            <div className="w-7 h-7 rounded-lg bg-primary flex items-center justify-center">
              <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
                <path d="M2 4h12M2 8h8M2 12h10" stroke="white" strokeWidth="1.5" strokeLinecap="round"/>
              </svg>
            </div>
            <span className="text-white font-semibold text-sm tracking-wide">Dobby</span>
          </div>
        </div>

        {/* Navigation */}
        <nav className="flex-1 px-3 py-4 space-y-0.5">
          {navItems.map(({ to, icon: Icon, label }) => (
            <NavLink
              key={to}
              to={to}
              end={to === '/'}
              className={({ isActive }) =>
                `flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium
                 transition-all duration-150 cursor-pointer
                 ${isActive
                   ? 'bg-sidebar-active text-white'
                   : 'text-slate-400 hover:bg-sidebar-hover hover:text-white'
                 }`
              }
            >
              <Icon size={16} strokeWidth={1.75} />
              {label}
            </NavLink>
          ))}
        </nav>

        {/* Footer */}
        <div className="px-5 py-3 border-t border-slate-800">
          <p className="text-xs text-slate-600">File Manager v1.0</p>
        </div>
      </aside>

      {/* Main content */}
      <main className="flex-1 overflow-auto bg-surface">
        <Outlet />
      </main>
    </div>
  )
}
