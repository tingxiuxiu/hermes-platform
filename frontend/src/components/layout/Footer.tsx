import { cn } from '@/lib/utils'
import { useTranslation } from 'react-i18next'

interface FooterProps {
  className?: string
}

const version = '0.0.0'

export function Footer({ className }: FooterProps) {
  const { t } = useTranslation()

  return (
    <footer
      className={cn(
        'h-12 flex items-center justify-center text-sm text-slate-500 dark:text-slate-400 bg-white dark:bg-slate-900 border-t border-slate-200 dark:border-slate-800',
        className
      )}
    >
      <div className="flex flex-col sm:flex-row items-center gap-1 sm:gap-4 px-4">
        <span>{t('footer.copyright')}</span>
        <span className="hidden sm:inline">|</span>
        <span>{t('footer.version')} {version}</span>
      </div>
    </footer>
  )
}
