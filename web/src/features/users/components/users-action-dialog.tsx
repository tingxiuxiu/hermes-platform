import { useEffect } from 'react'
import { z } from 'zod'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { toast } from 'sonner'
import { createUser, updateUser, type CreateUserPayload } from '../api/users-api'
import { type User } from '../data/schema'
import { useUsers } from './users-provider'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
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
import { Checkbox } from '@/components/ui/checkbox'

const formSchema = z.object({
  username: z.string().min(3, 'Username must be at least 3 characters.'),
  email: z.string().email('Enter a valid email.').or(z.literal('')),
  password: z.string().optional(),
  roleIds: z.array(z.number()).min(1, 'Select at least one role.'),
})
type UserForm = z.infer<typeof formSchema>

type UserActionDialogProps = {
  currentRow?: User
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function UsersActionDialog({
  currentRow,
  open,
  onOpenChange,
}: UserActionDialogProps) {
  const isEdit = !!currentRow
  const { roleOptions, updateUserRoles } = useUsers()
  const queryClient = useQueryClient()
  const form = useForm<UserForm>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      username: currentRow?.username ?? '',
      email: currentRow?.email ?? '',
      password: '',
      roleIds: currentRow?.roles.map((role) => role.id) ?? [],
    },
  })

  useEffect(() => {
    if (open) {
      form.reset({
        username: currentRow?.username ?? '',
        email: currentRow?.email ?? '',
        password: '',
        roleIds: currentRow?.roles.map((role) => role.id) ?? [],
      })
    }
  }, [currentRow, form, open])

  const mutation = useMutation({
    mutationFn: async (values: UserForm) => {
      if (isEdit && currentRow) {
        await updateUser(currentRow.id, {
          username: values.username,
          email: values.email || undefined,
        })
        await updateUserRoles(currentRow.id, values.roleIds)
        return
      }
      const payload: CreateUserPayload = {
        username: values.username,
        password: values.password ?? '',
        email: values.email || undefined,
        role_ids: values.roleIds,
      }
      await createUser(payload)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] })
      toast.success(isEdit ? 'User updated' : 'User created')
      onOpenChange(false)
    },
    onError: (error: Error) => toast.error(error.message),
  })

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='sm:max-w-lg'>
        <DialogHeader className='text-start'>
          <DialogTitle>{isEdit ? 'Edit User' : 'Add New User'}</DialogTitle>
          <DialogDescription>
            {isEdit ? 'Update the backend user record.' : 'Create a backend user record.'}
          </DialogDescription>
        </DialogHeader>
        <Form {...form}>
          <form onSubmit={form.handleSubmit((values) => mutation.mutate(values))} className='flex flex-col gap-4'>
            <FormField control={form.control} name='username' render={({ field }) => (
              <FormItem>
                <FormLabel>Username</FormLabel>
                <FormControl><Input {...field} autoComplete='off' /></FormControl>
                <FormMessage />
              </FormItem>
            )} />
            <FormField control={form.control} name='email' render={({ field }) => (
              <FormItem>
                <FormLabel>Email</FormLabel>
                <FormControl><Input {...field} type='email' /></FormControl>
                <FormMessage />
              </FormItem>
            )} />
            {!isEdit && (
              <FormField control={form.control} name='password' render={({ field }) => (
                <FormItem>
                  <FormLabel>Password</FormLabel>
                  <FormControl><PasswordInput {...field} /></FormControl>
                  <FormMessage />
                </FormItem>
              )} />
            )}
            <FormField control={form.control} name='roleIds' render={({ field }) => (
              <FormItem>
                <FormLabel>Roles</FormLabel>
                <div className='grid gap-2 sm:grid-cols-2'>
                  {roleOptions.map((role) => (
                    <label key={role.id} className='flex items-center gap-2 rounded-md border p-2 text-sm'>
                      <Checkbox
                        checked={field.value.includes(role.id)}
                        onCheckedChange={(checked) => field.onChange(
                          checked ? [...field.value, role.id] : field.value.filter((id) => id !== role.id)
                        )}
                      />
                      {role.name}
                    </label>
                  ))}
                </div>
                <FormMessage />
              </FormItem>
            )} />
            <DialogFooter>
              <Button type='button' variant='outline' onClick={() => onOpenChange(false)}>Cancel</Button>
              <Button type='submit' disabled={mutation.isPending}>
                {mutation.isPending ? 'Saving...' : 'Save changes'}
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  )
}
