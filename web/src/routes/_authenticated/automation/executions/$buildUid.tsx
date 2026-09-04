import z from 'zod'
import { createFileRoute } from '@tanstack/react-router'
import { ExecutionSessionPage } from '@/features/automation/session-page'

const sessionSearchSchema = z.object({
  case: z.string().optional(),
})

function ExecutionSessionRoute() {
  const { buildUid } = Route.useParams()
  const { case: caseUid } = Route.useSearch()
  return <ExecutionSessionPage buildUid={buildUid} caseUid={caseUid} />
}

export const Route = createFileRoute(
  '/_authenticated/automation/executions/$buildUid'
)({
  validateSearch: sessionSearchSchema,
  component: ExecutionSessionRoute,
})
