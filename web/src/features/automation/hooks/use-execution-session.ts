import { useEffect, useMemo, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import {
  applyLiveEvent,
  followLiveCaseUid,
  fromGoSteps,
  isTerminalExecution,
  mergeSessionItems,
  mergeSessionTrees,
  shouldLoadPersistedTree,
  sortCasesForSession,
  treesFromSnapshot,
  type SessionTrees,
} from '../data/live-tree'
import type { CaseItem, ExecutionStatus, LiveEvent } from '../data/schema'
import { getItemDetail, getLiveSnapshot, subscribeLiveEvents } from '../api/automation-api'

type SessionState = {
  items: CaseItem[]
  trees: SessionTrees
  executionStatus: ExecutionStatus
}

export function useExecutionSession(buildUid: string, urlCase?: string) {
  const liveQuery = useQuery({
    queryKey: ['automation', 'live', buildUid],
    queryFn: () => getLiveSnapshot(buildUid),
  })

  const [state, setState] = useState<SessionState>({
    items: [],
    trees: {},
    executionStatus: 'running',
  })
  const [selectedCaseUid, setSelectedCaseUid] = useState<string | null>(null)
  const [followLive, setFollowLive] = useState(!urlCase)
  const [hydrated, setHydrated] = useState(false)

  useEffect(() => {
    setState({ items: [], trees: {}, executionStatus: 'running' })
    setSelectedCaseUid(null)
    setFollowLive(!urlCase)
    setHydrated(false)
  }, [buildUid, urlCase])

  useEffect(() => {
    const snapshot = liveQuery.data
    if (!snapshot) return
    setState((current) => ({
      items: mergeSessionItems(current.items, snapshot.items),
      trees: mergeSessionTrees(current.trees, treesFromSnapshot(snapshot)),
      executionStatus: snapshot.execution.status,
    }))
    setHydrated(true)
  }, [liveQuery.data, urlCase])

  useEffect(() => {
    if (!hydrated) return
    setSelectedCaseUid((current) =>
      followLiveCaseUid(state.items, current, followLive, urlCase)
    )
  }, [hydrated, state.items, followLive, urlCase])

  const selected = useMemo(
    () => state.items.find((item) => item.case_uid === selectedCaseUid) ?? null,
    [state.items, selectedCaseUid]
  )

  const itemQuery = useQuery({
    queryKey: ['automation', 'item', selectedCaseUid],
    queryFn: () => getItemDetail(selectedCaseUid!),
    enabled:
      Boolean(selectedCaseUid) &&
      shouldLoadPersistedTree(
        selected?.status,
        Boolean(selectedCaseUid && state.trees[selectedCaseUid])
      ),
  })

  useEffect(() => {
    const detail = itemQuery.data
    if (!detail || detail.case_uid !== selectedCaseUid) return
    setState((current) => ({
      ...current,
      trees: {
        ...current.trees,
        [detail.case_uid]: fromGoSteps(detail.steps),
      },
    }))
  }, [itemQuery.data, selectedCaseUid])

  const refetchLive = liveQuery.refetch

  useEffect(() => {
    if (!hydrated || isTerminalExecution(state.executionStatus)) return

    const abort = new AbortController()
    let cancelled = false
    let delay = 1000

    const connect = async () => {
      while (!cancelled && !abort.signal.aborted) {
        try {
          await subscribeLiveEvents(
            buildUid,
            (event: LiveEvent) => {
              setState((current) => {
                const next = applyLiveEvent(
                  current.items,
                  current.trees,
                  current.executionStatus,
                  event
                )
                return {
                  items: next.items,
                  trees: next.trees,
                  executionStatus: next.executionStatus as ExecutionStatus,
                }
              })
              if (
                event.type === 'execution.updated' &&
                isTerminalExecution(event.status)
              ) {
                abort.abort()
                void refetchLive()
              }
            },
            abort.signal
          )
          delay = 1000
        } catch {
          if (cancelled || abort.signal.aborted) return
          delay = Math.min(delay * 2, 15_000)
        }
        if (cancelled || abort.signal.aborted) return
        await refetchLive()
        await sleep(delay, abort.signal)
      }
    }

    void connect()
    return () => {
      cancelled = true
      abort.abort()
    }
  }, [buildUid, hydrated, refetchLive, state.executionStatus])

  return {
    snapshot: liveQuery.data,
    isLoading: liveQuery.isLoading,
    error: liveQuery.error,
    items: sortCasesForSession(state.items),
    trees: state.trees,
    executionStatus: state.executionStatus,
    selectedCaseUid,
    selected,
    refetch: liveQuery.refetch,
    selectCase: (caseUid: string) => {
      const latestRunning = sortCasesForSession(state.items).find(
        (row) => row.status === 'running'
      )
      setFollowLive(caseUid === latestRunning?.case_uid)
      setSelectedCaseUid(caseUid)
    },
  }
}

function sleep(ms: number, signal: AbortSignal) {
  return new Promise<void>((resolve) => {
    const timer = setTimeout(resolve, ms)
    signal.addEventListener('abort', () => {
      clearTimeout(timer)
      resolve()
    })
  })
}
