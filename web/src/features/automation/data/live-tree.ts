import type {
  CaseItem,
  GoStepRow,
  ItemUpdatedEvent,
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

export function mergeSessionTrees(
  live: SessionTrees,
  snapshot: SessionTrees
): SessionTrees {
  return {
    ...live,
    ...snapshot,
  }
}

export function mergeSessionItems(live: CaseItem[], snapshot: CaseItem[]): CaseItem[] {
  const byUid = new Map(live.map((item) => [item.case_uid, item]))
  for (const item of snapshot) {
    const current = byUid.get(item.case_uid)
    if (!current) {
      byUid.set(item.case_uid, item)
      continue
    }
    if (current.status !== 'running' && item.status === 'running') {
      continue
    }
    byUid.set(item.case_uid, {
      ...current,
      ...item,
      case_name: item.case_name || current.case_name,
      case_key: item.case_key || current.case_key,
    })
  }
  return [...byUid.values()]
}

export function isTerminalCase(status: string | null | undefined): boolean {
  return Boolean(status) && status !== 'running'
}

export function shouldLoadPersistedTree(
  status: string | null | undefined,
  hasTree: boolean
): boolean {
  if (!status) return false
  if (isTerminalCase(status)) return true
  return !hasTree
}

export function applyLiveEvent(
  items: CaseItem[],
  trees: SessionTrees,
  executionStatus: string,
  event: LiveEvent
): { items: CaseItem[]; trees: SessionTrees; executionStatus: string } {
  if (event.type === 'step.upserted') {
    const existing = items.find((item) => item.case_uid === event.case_uid)
    if (existing && isTerminalCase(existing.status)) {
      return { items, trees, executionStatus }
    }
    const nextTrees = { ...trees }
    nextTrees[event.case_uid] = upsertStepTree(
      nextTrees[event.case_uid] ?? [],
      event
    )
    const nextItems = upsertRunningItem(items, event)
    return { items: nextItems, trees: nextTrees, executionStatus }
  }

  if (event.type === 'item.updated') {
    return {
      items: upsertItemFromEvent(items, event),
      trees,
      executionStatus,
    }
  }

  if (event.type === 'execution.updated') {
    return { items, trees, executionStatus: event.status }
  }

  return { items, trees, executionStatus }
}

function upsertItemFromEvent(items: CaseItem[], event: ItemUpdatedEvent): CaseItem[] {
  const index = items.findIndex((item) => item.case_uid === event.case_uid)
  const current = index >= 0 ? items[index] : null
  const nextItem: CaseItem = {
    id: current?.id ?? 0,
    case_uid: event.case_uid,
    case_key: event.case_key || current?.case_key || '',
    case_name: event.case_name || current?.case_name || 'Running case',
    attempt_number: event.attempt_number ?? current?.attempt_number ?? 1,
    status: event.status,
    start_time: event.start_time ?? current?.start_time ?? null,
    end_time: event.end_time ?? current?.end_time ?? null,
    duration: event.duration ?? current?.duration ?? null,
  }
  if (index < 0) return [...items, nextItem]
  const next = [...items]
  next[index] = { ...current!, ...nextItem }
  return next
}

export function followLiveCaseUid(
  items: CaseItem[],
  current: string | null,
  followLive: boolean,
  urlCase?: string | null
): string | null {
  if (urlCase && !followLive && items.some((item) => item.case_uid === urlCase)) {
    return urlCase
  }
  const running = sortCasesForSession(items).find((item) => item.status === 'running')
  if (followLive) {
    return running?.case_uid ?? current ?? items[0]?.case_uid ?? null
  }
  if (current && items.some((item) => item.case_uid === current)) {
    return current
  }
  return running?.case_uid ?? items[0]?.case_uid ?? null
}

export function activeStepPath(nodes: LiveStepNode[]): string | null {
  let last: string | null = null
  let running: string | null = null
  const walk = (list: LiveStepNode[]) => {
    for (const node of list) {
      last = node.step_path
      if (node.status === 'running') running = node.step_path
      walk(node.children)
    }
  }
  walk(nodes)
  return running ?? last
}

function upsertRunningItem(items: CaseItem[], event: StepDelta): CaseItem[] {
  const caseName = event.case_name?.trim()
  const caseKey = event.case_key?.trim()
  const index = items.findIndex((item) => item.case_uid === event.case_uid)
  if (index < 0) {
    return [
      ...items,
      {
        id: 0,
        case_uid: event.case_uid,
        case_key: caseKey || '',
        case_name: caseName || 'Running case',
        attempt_number: 1,
        status: 'running',
        start_time: null,
        end_time: null,
        duration: null,
      },
    ]
  }

  const current = items[index]
  const nextName =
    caseName && (!current.case_name || current.case_name === 'Running case')
      ? caseName
      : current.case_name
  const nextKey = caseKey && !current.case_key ? caseKey : current.case_key
  if (nextName === current.case_name && nextKey === current.case_key) {
    return items
  }
  const next = [...items]
  next[index] = { ...current, case_name: nextName, case_key: nextKey }
  return next
}

export function isTerminalExecution(status: string): boolean {
  return status !== 'running'
}
