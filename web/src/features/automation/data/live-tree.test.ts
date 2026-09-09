import { describe, expect, it } from 'vitest'
import {
  mergeSessionTrees,
  treesFromSnapshot,
  type SessionTrees,
} from './live-tree'
import type { LiveSnapshot, LiveStepNode } from './schema'

function step(path: string, name: string): LiveStepNode {
  return {
    step_path: path,
    step_name: name,
    status: 'passed',
    start_time: null,
    end_time: null,
    duration: null,
    children: [],
  }
}

function snapshotWithCurrent(
  caseUid: string,
  stepName: string
): LiveSnapshot {
  return {
    execution: {
      id: 1,
      build_uid: 'build-1',
      job_name: 'job',
      job_url: null,
      project_name: 'p',
      software_name: 's',
      software_version: '1',
      labels: null,
      status: 'running',
      start_time: '2026-01-01T00:00:00Z',
      end_time: null,
      duration: null,
      planned_cases_count: null,
      pass_count: null,
      failure_count: null,
      skipped_count: null,
      pass_rate: null,
      last_heartbeat_at: null,
    },
    items: [],
    current_item: {
      case_uid: caseUid,
      case_key: 'k',
      case_name: 'case',
      status: 'running',
      steps: [
        {
          step_path: '1',
          step_name: stepName,
          status: 'passed',
        },
      ],
    },
  }
}

describe('mergeSessionTrees', () => {
  it('lets persisted snapshot trees overwrite live trees for the same case', () => {
    const live: SessionTrees = {
      'case-a': [step('1', 'stale live step')],
      'case-b': [step('1', 'live only')],
    }
    const snapshot = treesFromSnapshot(
      snapshotWithCurrent('case-a', 'persisted step')
    )

    const merged = mergeSessionTrees(live, snapshot)

    expect(merged['case-a'][0].step_name).toBe('persisted step')
    expect(merged['case-b'][0].step_name).toBe('live only')
  })

  it('keeps live trees when the snapshot has no current item', () => {
    const live: SessionTrees = {
      'case-a': [step('1', 'live step')],
    }

    expect(mergeSessionTrees(live, treesFromSnapshot({
      ...snapshotWithCurrent('case-a', 'unused'),
      current_item: null,
    }))).toEqual(live)
  })
})
