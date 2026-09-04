import z from 'zod'
import { createFileRoute } from '@tanstack/react-router'
import { AutomationJobsPage } from '@/features/automation'

const automationJobsSearchSchema = z.object({
  page: z.number().optional().catch(1),
  pageSize: z.number().optional().catch(10),
  jobName: z.string().optional().catch(''),
  status: z
    .array(z.enum(['active', 'inactive']))
    .optional()
    .catch([]),
  lastBuildStatus: z
    .array(z.enum(['running', 'completed', 'failed', 'aborted', 'calculated']))
    .optional()
    .catch([]),
})

function AutomationJobsRouteComponent() {
  const search = Route.useSearch()
  const navigate = Route.useNavigate()

  return <AutomationJobsPage search={search} navigate={navigate} />
}

export const Route = createFileRoute('/_authenticated/automation/jobs/')({
  validateSearch: automationJobsSearchSchema,
  component: AutomationJobsRouteComponent,
})
