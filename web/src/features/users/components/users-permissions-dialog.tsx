import { useState } from 'react'
import { Check, ShieldCheck } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Checkbox } from '@/components/ui/checkbox'
import { type User } from '../data/schema'
import { useUsers } from './users-provider'

type UsersPermissionsDialogProps = {
  currentRow: User
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function UsersPermissionsDialog({
  currentRow,
  open,
  onOpenChange,
}: UsersPermissionsDialogProps) {
  const { roles, permissionTree, updateRolePermissions } = useUsers()
  const [roleId, setRoleId] = useState(currentRow.roles[0]?.id ?? 0)
  const selectedRole = roles.find((role) => role.id === roleId)
  const selectedPermissionIds = new Set(
    selectedRole?.permissions.map((permission) => permission.id) ?? []
  )

  const togglePermission = async (permissionId: number, checked: boolean) => {
    if (!selectedRole) return
    const nextPermissionIds = new Set(selectedPermissionIds)
    if (checked) nextPermissionIds.add(permissionId)
    else nextPermissionIds.delete(permissionId)
    try {
      await updateRolePermissions(selectedRole.id, [...nextPermissionIds])
      toast.success(`Permissions updated for ${selectedRole.name}`)
    } catch {
      toast.error('Unable to update permissions')
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='sm:max-w-lg'>
        <DialogHeader className='text-start'>
          <DialogTitle className='flex items-center gap-2'>
            <ShieldCheck />
            Configure role permissions
          </DialogTitle>
          <DialogDescription>
            Configure permissions for the roles assigned to {currentRow.username}.
          </DialogDescription>
        </DialogHeader>
        <div className='flex max-h-[60vh] flex-col gap-5 overflow-y-auto py-2'>
          <div className='flex flex-wrap gap-2'>
            {currentRow.roles.map((role) => (
              <Button
                key={role.id}
                size='sm'
                variant={role.id === roleId ? 'default' : 'outline'}
                onClick={() => setRoleId(role.id)}
              >
                {role.name}
              </Button>
            ))}
          </div>
          {permissionTree.map((permission) => (
            <label
              key={permission.id}
              className='flex items-center gap-3 rounded-md border p-3 text-sm hover:bg-muted/50'
            >
              <Checkbox
                checked={selectedPermissionIds.has(permission.id)}
                onCheckedChange={(checked) =>
                  void togglePermission(permission.id, checked === true)
                }
              />
              <span>{permission.name}</span>
              <span className='ms-auto text-xs text-muted-foreground'>
                {permission.code}
              </span>
            </label>
          ))}
        </div>
        <DialogFooter>
          <Button variant='outline' onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button onClick={() => onOpenChange(false)}>
            <Check data-icon='inline-start' />
            Done
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
