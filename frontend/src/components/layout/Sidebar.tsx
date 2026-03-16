import { useState } from 'react'
import { NavLink } from 'react-router'
import { LayoutDashboard, ClipboardList, Users, ChevronDown, ChevronRight, Folder, FileText, CheckSquare } from 'lucide-react'
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
}

export function Sidebar({ collapsed }: SidebarProps) {
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
      path: '/projects',
      label: t('sidebar.projects'),
      icon: <Folder className="w-5 h-5" />
    },
    {
      path: '/test-plans',
      label: t('sidebar.testPlans'),
      icon: <FileText className="w-5 h-5" />
    },
    {
      path: '/test-cases',
      label: t('sidebar.testCases'),
      icon: <CheckSquare className="w-5 h-5" />
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
        <div key={item.label} className="mb-0.5">
          <button
            onClick={() => toggleSubMenu(item.label)}
            className={cn(
              "w-full flex items-center gap-3 px-4 py-3 rounded-full text-sm font-medium transition-all duration-200 relative overflow-hidden",
              "text-on-surface-variant hover:text-on-surface",
              "hover:bg-surface-container-hov",
              collapsed && "justify-center px-0"
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
            <div className="ml-4 mt-1 pl-4 border-l border-outline/30">
              {item.children.map(child => (
                <NavLink
                  key={child.path}
                  to={child.path!}
                  className={({ isActive }) =>
                    cn(
                      "flex items-center gap-3 px-4 py-2.5 rounded-full text-sm transition-all duration-200 relative overflow-hidden",
                      isActive
                        ? "text-primary bg-primary-container font-medium"
                        : "text-on-surface-variant hover:text-on-surface hover:bg-surface-container-hov"
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
            "flex items-center gap-3 px-4 py-3 rounded-full text-sm font-medium transition-all duration-200 mb-0.5 relative overflow-hidden",
            isActive
              ? "text-primary bg-primary-container"
              : "text-on-surface-variant hover:text-on-surface hover:bg-surface-container-hov",
            collapsed && "justify-center px-0"
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
        "h-full bg-surface border-r border-border flex flex-col transition-all duration-300 shadow-level-1 z-20",
        collapsed ? "w-20" : "w-72"
      )}
    >
      <nav className="flex-1 p-3 overflow-y-auto">
        {menuItems.map(renderMenuItem)}
      </nav>
    </aside>
  )
}
