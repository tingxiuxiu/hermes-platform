import { useState } from 'react'
import { Bell, Check } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/lib/utils'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'
import { Button } from '@/components/ui/button'

interface Notification {
  id: string
  title: string
  time: string
  read: boolean
  type: 'system' | 'task' | 'user'
}

const mockNotifications: Notification[] = [
  { id: '1', title: '测试任务完成', time: '2分钟前', read: false, type: 'task' },
  { id: '2', title: '新用户注册', time: '15分钟前', read: false, type: 'user' },
  { id: '3', title: '系统更新', time: '1小时前', read: true, type: 'system' },
  { id: '4', title: '测试任务失败', time: '2小时前', read: true, type: 'task' },
  { id: '5', title: '权限变更通知', time: '3小时前', read: true, type: 'system' },
]

export function NotificationBell() {
  const { t } = useTranslation()
  const [notifications, setNotifications] = useState<Notification[]>(mockNotifications)
  const [open, setOpen] = useState(false)

  const unreadCount = notifications.filter((n) => !n.read).length

  const markAsRead = (id: string) => {
    setNotifications((prev) =>
      prev.map((n) => (n.id === id ? { ...n, read: true } : n))
    )
  }

  const markAllAsRead = () => {
    setNotifications((prev) => prev.map((n) => ({ ...n, read: true })))
  }

  const getTypeColor = (type: Notification['type']) => {
    switch (type) {
      case 'task':
        return 'bg-blue-500'
      case 'user':
        return 'bg-green-500'
      case 'system':
        return 'bg-orange-500'
      default:
        return 'bg-slate-500'
    }
  }

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <button
          className={cn(
            'relative p-2 rounded-full hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors',
            'focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2'
          )}
        >
          <Bell className="w-5 h-5 text-slate-600 dark:text-slate-300" />
          {unreadCount > 0 && (
            <span className="absolute top-0 right-0 w-2 h-2 bg-red-500 rounded-full animate-pulse" />
          )}
        </button>
      </PopoverTrigger>
      <PopoverContent align="end" className="w-80 p-0">
        <div className="flex items-center justify-between p-4 border-b border-slate-200 dark:border-slate-700">
          <h3 className="font-semibold text-slate-900 dark:text-slate-100">
            {t('notification.title', '系统通知')}
          </h3>
          {unreadCount > 0 && (
            <Button
              variant="ghost"
              size="sm"
              onClick={markAllAsRead}
              className="h-auto py-1 px-2 text-xs text-blue-600 dark:text-blue-400"
            >
              <Check className="w-3 h-3 mr-1" />
              {t('notification.markAllRead', '全部已读')}
            </Button>
          )}
        </div>
        <div className="max-h-80 overflow-y-auto">
          {notifications.length === 0 ? (
            <div className="p-8 text-center text-slate-500 dark:text-slate-400">
              {t('notification.empty', '暂无通知')}
            </div>
          ) : (
            notifications.map((notification) => (
              <div
                key={notification.id}
                onClick={() => markAsRead(notification.id)}
                className={cn(
                  'flex items-start gap-3 p-4 cursor-pointer transition-colors',
                  'hover:bg-slate-50 dark:hover:bg-slate-800/50',
                  !notification.read && 'bg-blue-50/50 dark:bg-blue-900/10'
                )}
              >
                <div className={cn('w-2 h-2 mt-2 rounded-full shrink-0', getTypeColor(notification.type))} />
                <div className="flex-1 min-w-0">
                  <p
                    className={cn(
                      'text-sm truncate',
                      notification.read
                        ? 'text-slate-600 dark:text-slate-400'
                        : 'text-slate-900 dark:text-slate-100 font-medium'
                    )}
                  >
                    {notification.title}
                  </p>
                  <p className="text-xs text-slate-500 dark:text-slate-500 mt-0.5">
                    {notification.time}
                  </p>
                </div>
                {!notification.read && (
                  <div className="w-2 h-2 bg-blue-500 rounded-full shrink-0 mt-2" />
                )}
              </div>
            ))
          )}
        </div>
      </PopoverContent>
    </Popover>
  )
}
