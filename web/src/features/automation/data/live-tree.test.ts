import { describe, expect, it } from 'vitest'
import {
  applyLiveEvent,
  fromGoSteps,
  parentStepPath,
  pickDefaultCaseUid,
  sortCasesForSession,
  upsertStepTree,
} from './live-tree'
import type { CaseItem, LiveSnapshot } from './schema'

const sampleTree = fromGoSteps([
  {
    step_path: '0',
    step_name: '登录',
    status: 'passed',
    sub_steps: [
      {
        step_path: '0.0',
        step_name: '输入账号',
        status: 'running',
        sub_steps: [],
      },
    ],
  },
])

describe('fromGoSteps', () => {
  it('maps nested sub_steps into children', () => {
    expect(sampleTree).toEqual([
      {
        step_path: '0',
        step_name: '登录',
        status: 'passed',
        start_time: null,
        end_time: null,
        duration: null,
        children: [
          {
            step_path: '0.0',
            step_name: '输入账号',
            status: 'running',
            start_time: null,
            end_time: null,
            duration: null,
            children: [],
          },
        ],
      },
    ])
  })
})

describe('upsertStepTree', () => {
  it('updates an existing node without dropping siblings', () => {
    const next = upsertStepTree(sampleTree, {
      step_path: '0.0',
      step_name: '输入账号',
      status: 'passed',
    })
    expect(next[0]?.children[0]?.status).toBe('passed')
    expect(next[0]?.step_name).toBe('登录')
  })

  it('inserts a child under its parent path', () => {
    const next = upsertStepTree(sampleTree, {
      step_path: '0.1',
      step_name: '点登录',
      status: 'running',
    })
    expect(next[0]?.children.map((child) => child.step_path)).toEqual([
      '0.0',
      '0.1',
    ])
  })

  it('adds a new root when the path has no parent', () => {
    const next = upsertStepTree([], {
      step_path: '0',
      step_name: '打开页面',
      status: 'running',
    })
    expect(next[0]?.step_path).toBe('0')
    expect(next[0]?.status).toBe('running')
  })
})

describe('parentStepPath', () => {
  it('strips the last segment', () => {
    expect(parentStepPath('0.1.2')).toBe('0.1')
    expect(parentStepPath('0')).toBeNull()
  })
})

describe('pickDefaultCaseUid', () => {
  const snapshot: LiveSnapshot = {
    execution: {
      id: 1,
      build_uid: 'b',
      job_name: 'job',
      job_url: null,
      project_name: 'p',
      software_name: 's',
      software_version: '1',
      labels: null,
      status: 'running',
      start_time: '2026-09-04T12:00:00Z',
      end_time: null,
      duration: null,
      planned_cases_count: 2,
      pass_count: 1,
      failure_count: 0,
      skipped_count: 0,
      pass_rate: null,
      last_heartbeat_at: null,
    },
    items: [
      caseItem('done', 'passed', '2026-09-04T12:01:00Z'),
      caseItem('run', 'running', null),
    ],
    current_item: {
      case_uid: 'run',
      case_key: 'k',
      case_name: 'n',
      status: 'running',
      steps: [],
    },
  }

  it('prefers the URL case, then the running current_item', () => {
    expect(pickDefaultCaseUid(snapshot, 'done')).toBe('done')
    expect(pickDefaultCaseUid(snapshot, null)).toBe('run')
  })
})

describe('sortCasesForSession', () => {
  it('pins running cases to the top', () => {
    const items: CaseItem[] = [
      caseItem('a', 'passed', '2026-09-04T12:02:00Z'),
      caseItem('b', 'running', null),
    ]
    expect(sortCasesForSession(items).map((item) => item.case_uid)).toEqual([
      'b',
      'a',
    ])
  })
})

describe('applyLiveEvent', () => {
  it('upserts a step onto the matching case tree', () => {
    const result = applyLiveEvent(
      [caseItem('run', 'running', null)],
      { run: sampleTree },
      'running',
      {
        type: 'step.upserted',
        build_uid: 'b',
        case_uid: 'run',
        step_path: '0.0',
        step_name: '输入账号',
        status: 'passed',
      }
    )
    expect(result.trees.run?.[0]?.children[0]?.status).toBe('passed')
  })

  it('marks execution terminal on execution.updated', () => {
    const result = applyLiveEvent([], {}, 'running', {
      type: 'execution.updated',
      build_uid: 'b',
      status: 'aborted',
    })
    expect(result.executionStatus).toBe('aborted')
  })
})

function caseItem(
  uid: string,
  status: CaseItem['status'],
  endTime: string | null
): CaseItem {
  return {
    id: 1,
    case_uid: uid,
    case_key: uid,
    case_name: uid,
    attempt_number: 1,
    status,
    start_time: '2026-09-04T12:00:00Z',
    end_time: endTime,
    duration: null,
  }
}
