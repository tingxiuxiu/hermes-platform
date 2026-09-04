import z from 'zod'
import { createFileRoute } from '@tanstack/react-router'
import { Dashboard } from '@/features/dashboard'

const dashboardSearchSchema = z.object({
  days: z.number().optional().catch(7),
  jobId: z.number().optional(),
  jobName: z.string().optional().catch(''),
})

function DashboardRouteComponent() {
  const search = Route.useSearch()
  const navigate = Route.useNavigate()

  return <Dashboard search={search} navigate={navigate} />
}

export const Route = createFileRoute('/_authenticated/')({
  validateSearch: dashboardSearchSchema,
  component: DashboardRouteComponent,
})
