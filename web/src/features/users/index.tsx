import { getRouteApi } from '@tanstack/react-router'
import { useQuery } from '@tanstack/react-query'
import { ConfigDrawer } from '@/components/config-drawer'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import { UsersDialogs } from './components/users-dialogs'
import { UsersPrimaryButtons } from './components/users-primary-buttons'
import { UsersProvider } from './components/users-provider'
import { UsersTable } from './components/users-table'
import { getUsers } from './api/users-api'

const route = getRouteApi('/_authenticated/users/')

export function Users() {
  return (
    <UsersProvider>
      <UsersContent />
    </UsersProvider>
  )
}

function UsersContent() {
  const search = route.useSearch()
  const navigate = route.useNavigate()
  const { data, isLoading } = useQuery({
    queryKey: ['users', search],
    queryFn: () =>
      getUsers({
        page: search.page ?? 1,
        pageSize: search.pageSize ?? 10,
        username: search.username || undefined,
        status: search.status?.length === 1 ? search.status[0] : undefined,
        roleId: search.role?.length === 1 ? Number(search.role[0]) : undefined,
      }),
  })

  return (
    <>
      <Header fixed>
        <Search className='me-auto' />
        <ThemeSwitch />
        <ConfigDrawer />
        <ProfileDropdown />
      </Header>

      <Main className='flex flex-1 flex-col gap-4 sm:gap-6'>
        <div className='flex flex-wrap items-end justify-between gap-2'>
          <div>
            <h2 className='text-2xl font-bold tracking-tight'>User List</h2>
            <p className='text-muted-foreground'>
              Manage your users and their roles here.
            </p>
          </div>
          <UsersPrimaryButtons />
        </div>
        <UsersTable
          data={data?.items ?? []}
          total={data?.total ?? 0}
          loading={isLoading}
          search={search}
          navigate={navigate}
        />
      </Main>

      <UsersDialogs />
    </>
  )
}
