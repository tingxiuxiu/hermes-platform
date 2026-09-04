import z from 'zod'
import { createFileRoute } from '@tanstack/react-router'
import { AutomationExecutionsPage } from '@/features/automation'

const automationExecutionsSearchSchema = z.object({
  page: z.number().optional().catch(1),
  pageSize: z.number().optional().catch(10),
  jobId: z.number().optional(),
  jobName: z.string().optional().catch(''),
  status: z
    .array(z.enum(['running', 'completed', 'failed', 'aborted', 'calculated']))
    .optional()
    .catch([]),
})

function AutomationExecutionsRouteComponent() {
  const search = Route.useSearch()
  const navigate = Route.useNavigate()

  return <AutomationExecutionsPage search={search} navigate={navigate} />
}

export const Route = createFileRoute('/_authenticated/automation/executions/')({
  validateSearch: automationExecutionsSearchSchema,
  component: AutomationExecutionsRouteComponent,
})
