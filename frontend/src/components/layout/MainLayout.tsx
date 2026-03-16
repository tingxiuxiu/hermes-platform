import { useState } from 'react'
import { Link } from 'react-router'
import { PanelLeft, PanelLeftClose } from 'lucide-react'
import { UserMenu } from '@/components/UserMenu'
import { ThemeToggle } from '@/components/ui/ThemeToggle'
import { Sidebar } from '@/components/layout/Sidebar'
import { Footer } from '@/components/layout/Footer'
import { NotificationBell } from '@/components/layout/NotificationBell'
import { LanguageSwitcher } from '@/components/ui/languageSwitcher'
import { useTranslation } from 'react-i18next'

interface MainLayoutProps {
  children: React.ReactNode
}

export function MainLayout({ children }: MainLayoutProps) {
  const [isSidebarCollapsed, setIsSidebarCollapsed] = useState(false)
  const { t } = useTranslation()

  return (
    <div className="flex flex-col h-screen bg-background text-foreground overflow-hidden">
      {/* 上部 - 顶部导航栏 */}
      <header className="flex items-center justify-between h-16 px-4 bg-surface border-b border-border shadow-level-2 shrink-0 z-10">
        <div className="flex items-center gap-3">
          <button
            onClick={() => setIsSidebarCollapsed(!isSidebarCollapsed)}
            className="p-2 rounded-full text-on-surface-variant hover:bg-surface-variant/80 transition-colors focus:outline-none focus:ring-2 focus:ring-primary/40 focus:ring-offset-2 focus:ring-offset-background"
            aria-label={isSidebarCollapsed ? t('sidebar.expand') : t('sidebar.collapse')}
          >
            {isSidebarCollapsed ? (
              <PanelLeft className="w-5 h-5" />
            ) : (
              <PanelLeftClose className="w-5 h-5" />
            )}
          </button>
          <Link to="/" className="flex items-center gap-2">
            <div className="w-10 h-10 bg-primary-container rounded-2xl flex items-center justify-center">
              <span className="text-primary font-bold text-lg">H</span>
            </div>
            <span className="text-xl font-medium text-on-surface">
              {t('navbar.brand')}
            </span>
          </Link>
        </div>

        <div className="flex items-center gap-1">
          <NotificationBell />
          <LanguageSwitcher />
          <ThemeToggle />
          <UserMenu />
        </div>
      </header>

      {/* 中间 - 左侧菜单 + 右侧内容 */}
      <div className="flex flex-1 min-h-0 overflow-hidden">
        <Sidebar collapsed={isSidebarCollapsed} />
        
        <main className="flex-1 overflow-auto p-6">
          <div className="max-w-7xl mx-auto">
            {children}
          </div>
        </main>
      </div>

      {/* 下部 - 页脚 */}
      <Footer />
    </div>
  )
}
