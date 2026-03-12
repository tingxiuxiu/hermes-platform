import { useState } from 'react'
import { Link, useLocation } from 'react-router'
import { Bell, Menu, X, Home, ClipboardList, Users, Settings } from 'lucide-react'
import { cn } from '@/lib/utils'
import { UserMenu } from '@/components/UserMenu'

interface MainLayoutProps {
  children: React.ReactNode
}

const navItems = [
  { path: '/', label: '仪表盘', icon: Home },
  { path: '/tasks', label: '任务管理', icon: ClipboardList },
  { path: '/admin/users', label: '用户管理', icon: Users },
  { path: '/settings', label: '系统设置', icon: Settings },
]

export function MainLayout({ children }: MainLayoutProps) {
  const [isSidebarCollapsed, setIsSidebarCollapsed] = useState(false)
  const location = useLocation()

  const sidebarWidth = isSidebarCollapsed ? 'w-16' : 'w-60'

  return (
    <div className="flex flex-col h-screen bg-slate-50 dark:bg-slate-900">
      <header className="flex items-center justify-between h-16 px-4 bg-white dark:bg-slate-800 border-b border-slate-200 dark:border-slate-700 shrink-0">
        <div className="flex items-center gap-3">
          <button
            onClick={() => setIsSidebarCollapsed(!isSidebarCollapsed)}
            className="p-2 rounded-lg text-slate-600 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-700 transition-colors"
            aria-label={isSidebarCollapsed ? '展开菜单' : '折叠菜单'}
          >
            {isSidebarCollapsed ? (
              <Menu className="w-5 h-5" />
            ) : (
              <X className="w-5 h-5" />
            )}
          </button>
          <Link to="/" className="flex items-center gap-2">
            <div className="w-8 h-8 bg-gradient-to-br from-blue-500 to-indigo-600 rounded-lg flex items-center justify-center">
              <span className="text-white font-bold text-sm">H</span>
            </div>
            <span className="text-lg font-semibold text-slate-900 dark:text-slate-100">
              Hermes Platform
            </span>
          </Link>
        </div>

        <div className="flex items-center gap-2">
          <button
            className="relative p-2 rounded-lg text-slate-600 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-700 transition-colors"
            aria-label="通知"
          >
            <Bell className="w-5 h-5" />
            <span className="absolute top-1 right-1 w-2 h-2 bg-red-500 rounded-full" />
          </button>
          <UserMenu />
        </div>
      </header>

      <div className="flex flex-1 overflow-hidden">
        <aside
          className={cn(
            'shrink-0 bg-white dark:bg-slate-800 border-r border-slate-200 dark:border-slate-700 transition-all duration-300',
            sidebarWidth
          )}
        >
          <nav className="h-full overflow-y-auto p-2">
            <div className="space-y-1">
              {navItems.map((item) => {
                const isActive = location.pathname === item.path
                const Icon = item.icon
                return (
                  <SidebarItem
                    key={item.path}
                    path={item.path}
                    icon={<Icon className="w-5 h-5" />}
                    label={item.label}
                    collapsed={isSidebarCollapsed}
                    active={isActive}
                  />
                )
              })}
            </div>
          </nav>
        </aside>

        <main className="flex-1 overflow-auto p-4">
          {children}
        </main>
      </div>

      <footer className="flex items-center justify-center h-12 px-4 bg-white dark:bg-slate-800 border-t border-slate-200 dark:border-slate-700 shrink-0">
        <p className="text-sm text-slate-500 dark:text-slate-400">
          © 2024 Hermes Platform. All rights reserved.
        </p>
      </footer>
    </div>
  )
}

interface SidebarItemProps {
  path: string
  icon: React.ReactNode
  label: string
  collapsed: boolean
  active?: boolean
}

function SidebarItem({ path, icon, label, collapsed, active }: SidebarItemProps) {
  return (
    <Link
      to={path}
      className={cn(
        'flex items-center w-full gap-3 px-3 py-2 rounded-lg text-sm font-medium transition-colors',
        active
          ? 'bg-blue-50 dark:bg-blue-950/30 text-blue-600 dark:text-blue-400'
          : 'text-slate-600 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-700'
      )}
      title={collapsed ? label : undefined}
    >
      <span className="shrink-0">{icon}</span>
      {!collapsed && <span>{label}</span>}
    </Link>
  )
}
