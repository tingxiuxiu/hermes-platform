import { useState } from 'react'
import { NavLink } from 'react-router'
import { LayoutDashboard, ClipboardList, Users, ChevronDown, ChevronRight, PanelLeftClose, PanelLeft } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useTranslation } from 'react-i18next'

interface MenuItem {
  path?: string
  label: string
  icon: React.ReactNode
  children?: MenuItem[]
}

interface SidebarProps {
  collapsed: boolean
  onToggle: () => void
}

export function Sidebar({ collapsed, onToggle }: SidebarProps) {
  const { t } = useTranslation()
  const [expandedMenus, setExpandedMenus] = useState<string[]>([])

  const menuItems: MenuItem[] = [
    {
      path: '/',
      label: t('sidebar.dashboard'),
      icon: <LayoutDashboard className="w-5 h-5" />
    },
    {
      path: '/tasks',
      label: t('sidebar.tasks'),
      icon: <ClipboardList className="w-5 h-5" />
    },
    {
      label: t('sidebar.systemConfig'),
      icon: <Users className="w-5 h-5" />,
      children: [
        {
          path: '/admin/users',
          label: t('sidebar.userManagement'),
          icon: <Users className="w-4 h-4" />
        }
      ]
    }
  ]

  const toggleSubMenu = (label: string) => {
    if (collapsed) return
    setExpandedMenus(prev =>
      prev.includes(label)
        ? prev.filter(item => item !== label)
        : [...prev, label]
    )
  }

  const renderMenuItem = (item: MenuItem) => {
    if (item.children) {
      const isExpanded = expandedMenus.includes(item.label)
      
      return (
        <div key={item.label} className="mb-1">
          <button
            onClick={() => toggleSubMenu(item.label)}
            className={cn(
              "w-full flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm font-medium transition-all duration-300",
              "text-slate-600 dark:text-slate-300 hover:text-blue-600 dark:hover:text-blue-400",
              "hover:bg-slate-100 dark:hover:bg-slate-800",
              collapsed && "justify-center"
            )}
          >
            {item.icon}
            {!collapsed && (
              <>
                <span className="flex-1 text-left">{item.label}</span>
                {isExpanded ? (
                  <ChevronDown className="w-4 h-4" />
                ) : (
                  <ChevronRight className="w-4 h-4" />
                )}
              </>
            )}
          </button>
          
          {!collapsed && isExpanded && (
            <div className="ml-4 mt-1 pl-4 border-l border-slate-200 dark:border-slate-700">
              {item.children.map(child => (
                <NavLink
                  key={child.path}
                  to={child.path!}
                  className={({ isActive }) =>
                    cn(
                      "flex items-center gap-3 px-3 py-2 rounded-xl text-sm transition-all duration-300",
                      isActive
                        ? "text-blue-600 dark:text-blue-400 bg-blue-50 dark:bg-blue-950/40 font-medium"
                        : "text-slate-600 dark:text-slate-300 hover:text-blue-600 dark:hover:text-blue-400 hover:bg-slate-100 dark:hover:bg-slate-800"
                    )
                  }
                >
                  {child.icon}
                  <span>{child.label}</span>
                </NavLink>
              ))}
            </div>
          )}
        </div>
      )
    }

    return (
      <NavLink
        key={item.path}
        to={item.path!}
        className={({ isActive }) =>
          cn(
            "flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm font-medium transition-all duration-300 mb-1",
            isActive
              ? "text-blue-600 dark:text-blue-400 bg-blue-50 dark:bg-blue-950/40"
              : "text-slate-600 dark:text-slate-300 hover:text-blue-600 dark:hover:text-blue-400 hover:bg-slate-100 dark:hover:bg-slate-800",
            collapsed && "justify-center"
          )
        }
        title={collapsed ? item.label : undefined}
      >
        {item.icon}
        {!collapsed && <span>{item.label}</span>}
      </NavLink>
    )
  }

  return (
    <aside
      className={cn(
        "h-screen bg-white dark:bg-slate-900 border-r border-slate-200 dark:border-slate-800 flex flex-col transition-all duration-300",
        collapsed ? "w-16" : "w-60"
      )}
    >
      <div className="flex items-center justify-between h-16 px-4 border-b border-slate-200 dark:border-slate-800">
        {!collapsed && (
          <span className="text-lg font-bold bg-gradient-to-r from-blue-600 to-indigo-600 bg-clip-text text-transparent">
            {t('navbar.brand')}
          </span>
        )}
        <button
          onClick={onToggle}
          className={cn(
            "p-2 rounded-xl text-slate-600 dark:text-slate-300 hover:text-blue-600 dark:hover:text-blue-400 hover:bg-slate-100 dark:hover:bg-slate-800 transition-all duration-300",
            collapsed && "mx-auto"
          )}
          title={collapsed ? t('sidebar.expand') : t('sidebar.collapse')}
        >
          {collapsed ? (
            <PanelLeft className="w-5 h-5" />
          ) : (
            <PanelLeftClose className="w-5 h-5" />
          )}
        </button>
      </div>

      <nav className="flex-1 p-3 overflow-y-auto">
        {menuItems.map(renderMenuItem)}
      </nav>
    </aside>
  )
}
