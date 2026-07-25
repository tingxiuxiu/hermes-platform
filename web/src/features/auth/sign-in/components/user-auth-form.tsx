import { useState } from 'react'
import { z } from 'zod'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { Link, useNavigate } from '@tanstack/react-router'
import { Loader2, LogIn } from 'lucide-react'
import { toast } from 'sonner'
import { useAuthStore } from '@/stores/auth-store'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { PasswordInput } from '@/components/password-input'
import { loginApi } from '@/features/auth/api/auth-api'

// ─── Validation Schema ────────────────────────────────────────────────────────

const formSchema = z.object({
  username: z
    .string()
    .min(1, '请输入用户名。')
    .min(3, '用户名至少需要 3 个字符。'),
  password: z.string().min(1, '请输入密码。').min(6, '密码至少需要 6 个字符。'),
})

type FormValues = z.infer<typeof formSchema>

// ─── Component Props ──────────────────────────────────────────────────────────

interface UserAuthFormProps extends React.HTMLAttributes<HTMLFormElement> {
  redirectTo?: string
}

// ─── Component ────────────────────────────────────────────────────────────────

export function UserAuthForm({
  className,
  redirectTo,
  ...props
}: UserAuthFormProps) {
  const [isLoading, setIsLoading] = useState(false)
  const navigate = useNavigate()
  const { auth } = useAuthStore()

  const form = useForm<FormValues>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      username: '',
      password: '',
    },
  })

  async function onSubmit(values: FormValues) {
    setIsLoading(true)
    try {
      const response = await loginApi({
        username: values.username,
        password: values.password,
      })

      if (!response.success) {
        toast.error(response.message || '登录失败，请重试。')
        return
      }

      const { access_token, user } = response.data

      // Persist token and user in store (and cookies)
      auth.setAccessToken(access_token)
      if (user) {
        auth.setUser(user)
      }

      toast.success(`欢迎回来，${user?.username ?? values.username}！`)

      const targetPath = redirectTo || '/'
      navigate({ to: targetPath, replace: true })
    } catch (err: unknown) {
      // The axios interceptor already maps `detail` → error.message
      const message =
        err instanceof Error ? err.message : '登录时发生错误，请重试。'
      toast.error(message)
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <Form {...form}>
      <form
        id='sign-in-form'
        onSubmit={form.handleSubmit(onSubmit)}
        className={cn('grid gap-3', className)}
        {...props}
      >
        {/* Username */}
        <FormField
          control={form.control}
          name='username'
          render={({ field }) => (
            <FormItem>
              <FormLabel>用户名或邮箱</FormLabel>
              <FormControl>
                <Input
                  id='sign-in-username'
                  placeholder='请输入用户名或邮箱'
                  autoComplete='username'
                  {...field}
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        {/* Password */}
        <FormField
          control={form.control}
          name='password'
          render={({ field }) => (
            <FormItem className='relative'>
              <FormLabel>密码</FormLabel>
              <FormControl>
                <PasswordInput
                  id='sign-in-password'
                  placeholder='••••••••'
                  autoComplete='current-password'
                  {...field}
                />
              </FormControl>
              <FormMessage />
              <Link
                to='/forgot-password'
                className='absolute inset-e-0 -top-0.5 text-sm font-medium text-muted-foreground hover:opacity-75'
              >
                忘记密码？
              </Link>
            </FormItem>
          )}
        />

        {/* Submit */}
        <Button id='sign-in-submit' className='mt-2' disabled={isLoading}>
          {isLoading ? <Loader2 className='animate-spin' /> : <LogIn />}登 录
        </Button>
      </form>
    </Form>
  )
}
