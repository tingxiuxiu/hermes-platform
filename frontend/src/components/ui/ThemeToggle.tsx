import { Sun, Moon } from 'lucide-react'
import { Button } from './button'
import { useThemeStore } from '@/stores/themeStore'

interface ThemeToggleProps {
  className?: string
}

export function ThemeToggle({ className }: ThemeToggleProps) {
  const { theme, toggleTheme } = useThemeStore()
  const isDark = theme === 'dark'

  return (
    <Button
      onClick={toggleTheme}
      variant="ghost"
      size="sm"
      className={`flex items-center justify-center ${className}`}
    >
      {isDark ? (
        <Sun className="h-4 w-4" />
      ) : (
        <Moon className="h-4 w-4" />
      )}
    </Button>
  )
}
