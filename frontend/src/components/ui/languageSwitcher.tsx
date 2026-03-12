import { useTranslation } from 'react-i18next'
import { Button } from './button'

interface LanguageSwitcherProps {
  className?: string
}

export function LanguageSwitcher({ className }: LanguageSwitcherProps) {
  const { i18n } = useTranslation()
  const currentLang = i18n.language

  const handleLanguageChange = () => {
    const newLang = currentLang === 'en' ? 'zh' : 'en'
    i18n.changeLanguage(newLang)
  }

  return (
    <Button
      onClick={handleLanguageChange}
      variant="ghost"
      size="sm"
      className={`flex items-center gap-2 ${className}`}
    >
      {currentLang === 'en' ? '中文' : 'English'}
    </Button>
  )
}