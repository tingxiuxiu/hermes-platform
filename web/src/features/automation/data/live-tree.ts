import type {
  CaseItem,
  GoStepRow,
  LiveEvent,
  LiveSnapshot,
  LiveStepNode,
  StepDelta,
} from './schema'

export function fromGoSteps(steps: GoStepRow[] | null | undefined): LiveStepNode[] {
  if (!steps?.length) return []
  return steps.map(fromGoStep)
}

function fromGoStep(step: GoStepRow): LiveStepNode {
  return {
    step_path: step.step_path,
    step_name: step.step_name,
    status: step.status,
    start_time: step.start_time ?? null,
    end_time: step.end_time ?? null,
    duration: step.duration ?? null,
    children: fromGoSteps(step.sub_steps),
  }
}

export function parentStepPath(path: string): string | null {
  const index = path.lastIndexOf('.')
  if (index < 0) return null
  return path.slice(0, index)
}

export function upsertStepTree(
  roots: LiveStepNode[],
  delta: Pick<StepDelta, 'step_path' | 'step_name' | 'status'>
): LiveStepNode[] {
  const next = cloneTree(roots)
  if (updateNode(next, delta)) return next

  const node: LiveStepNode = {
    step_path: delta.step_path,
    step_name: delta.step_name,
    status: delta.status,
    start_time: null,
    end_time: delta.status === 'running' ? null : null,
    duration: null,
    children: [],
  }

  const parentPath = parentStepPath(delta.step_path)
  if (!parentPath) {
    next.push(node)
    return next
  }
  if (insertUnder(next, parentPath, node)) return next
  next.push(node)
  return next
}

function cloneTree(nodes: LiveStepNode[]): LiveStepNode[] {
  return nodes.map((node) => ({
    ...node,
    children: cloneTree(node.children),
  }))
}

function updateNode(
  nodes: LiveStepNode[],
  delta: Pick<StepDelta, 'step_path' | 'step_name' | 'status'>
): boolean {
  for (const node of nodes) {
    if (node.step_path === delta.step_path) {
      node.step_name = delta.step_name
      node.status = delta.status
      return true
    }
    if (updateNode(node.children, delta)) return true
  }
  return false
}

function insertUnder(
  nodes: LiveStepNode[],
  parentPath: string,
  child: LiveStepNode
): boolean {
  for (const node of nodes) {
    if (node.step_path === parentPath) {
      node.children.push(child)
      return true
    }
    if (insertUnder(node.children, parentPath, child)) return true
  }
  return false
}

export function sortCasesForSession(items: CaseItem[]): CaseItem[] {
  return [...items].sort((left, right) => {
    const leftRunning = left.status === 'running' ? 0 : 1
    const rightRunning = right.status === 'running' ? 0 : 1
    if (leftRunning !== rightRunning) return leftRunning - rightRunning
    return (right.start_time ?? '').localeCompare(left.start_time ?? '')
  })
}

export function pickDefaultCaseUid(
  snapshot: LiveSnapshot,
  urlCase?: string | null
): string | null {
  if (urlCase && snapshot.items.some((item) => item.case_uid === urlCase)) {
    return urlCase
  }
  if (snapshot.current_item?.case_uid) return snapshot.current_item.case_uid

  const finished = snapshot.items.filter((item) => item.status !== 'running')
  finished.sort((left, right) =>
    (right.end_time ?? '').localeCompare(left.end_time ?? '')
  )
  return finished[0]?.case_uid ?? snapshot.items[0]?.case_uid ?? null
}

export type SessionTrees = Record<string, LiveStepNode[]>

export function treesFromSnapshot(snapshot: LiveSnapshot): SessionTrees {
  const trees: SessionTrees = {}
  if (snapshot.current_item) {
    trees[snapshot.current_item.case_uid] = fromGoSteps(
      snapshot.current_item.steps
    )
  }
  return trees
}

export function applyLiveEvent(
  items: CaseItem[],
  trees: SessionTrees,
  executionStatus: string,
  event: LiveEvent
): { items: CaseItem[]; trees: SessionTrees; executionStatus: string } {
  if (event.type === 'step.upserted') {
    const nextTrees = { ...trees }
    nextTrees[event.case_uid] = upsertStepTree(
      nextTrees[event.case_uid] ?? [],
      event
    )
    const nextItems = upsertRunningItem(items, event.case_uid)
    return { items: nextItems, trees: nextTrees, executionStatus }
  }

  if (event.type === 'item.updated') {
    const nextItems = items.map((item) =>
      item.case_uid === event.case_uid
        ? {
            ...item,
            status: event.status,
            end_time: event.end_time ?? item.end_time,
            duration: event.duration ?? item.duration,
          }
        : item
    )
    return { items: nextItems, trees, executionStatus }
  }

  if (event.type === 'execution.updated') {
    return { items, trees, executionStatus: event.status }
  }

  return { items, trees, executionStatus }
}

function upsertRunningItem(items: CaseItem[], caseUid: string): CaseItem[] {
  if (items.some((item) => item.case_uid === caseUid)) return items
  return [
    ...items,
    {
      id: 0,
      case_uid: caseUid,
      case_key: '',
      case_name: 'Running case',
      attempt_number: 1,
      status: 'running',
      start_time: null,
      end_time: null,
      duration: null,
    },
  ]
}

export function isTerminalExecution(status: string): boolean {
  return status !== 'running'
}
