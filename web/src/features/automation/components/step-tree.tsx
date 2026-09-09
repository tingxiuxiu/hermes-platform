import { useEffect, useRef, useState } from 'react'
import { ChevronDown } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Badge } from '@/components/ui/badge'
import type { LiveStepNode } from '../data/schema'
import { activeStepPath } from '../data/live-tree'
import { formatDateTime, formatDuration, statusColor } from '../data/utils'

type StepTreeProps = {
  nodes: LiveStepNode[]
}

export function StepTree({ nodes }: StepTreeProps) {
  const scrollerRef = useRef<HTMLDivElement>(null)
  const followRef = useRef(true)
  const activePath = activeStepPath(nodes)

  useEffect(() => {
    if (!followRef.current || !activePath) return
    const root = scrollerRef.current
    if (!root) return
    const target = root.querySelector(
      `[data-step-path="${CSS.escape(activePath)}"]`
    )
    target?.scrollIntoView({ block: 'nearest', behavior: 'smooth' })
  }, [nodes, activePath])

  if (!nodes.length) {
    return (
      <div className='py-10 text-center text-sm text-muted-foreground'>
        还没有步骤上报。
      </div>
    )
  }

  return (
    <div
      ref={scrollerRef}
      className='min-h-0 flex-1 overflow-y-auto'
      onScroll={(event) => {
        const root = event.currentTarget
        const gap = root.scrollHeight - root.scrollTop - root.clientHeight
        followRef.current = gap < 96
      }}
    >
      <div className='flex flex-col gap-3 p-4 pt-0'>
        {nodes.map((node) => (
          <StepTreeNodeView
            key={node.step_path}
            node={node}
            depth={0}
            activePath={activePath}
          />
        ))}
      </div>
    </div>
  )
}

function StepTreeNodeView({
  node,
  depth,
  activePath,
}: {
  node: LiveStepNode
  depth: number
  activePath: string | null
}) {
  const [collapsed, setCollapsed] = useState(false)
  const hasChildren = node.children.length > 0
  const running = node.status === 'running'
  const active = node.step_path === activePath

  return (
    <div className={cn(depth > 0 && 'ml-4 border-l border-[#e0e0e0] pl-4')}>
      <div
        data-step-path={node.step_path}
        data-step-active={active ? 'true' : undefined}
        className={cn(
          'rounded-[18px] border bg-white p-4 dark:bg-background',
          running ? 'border-[#0066cc]' : 'border-[#e0e0e0]'
        )}
      >
        <div className='flex flex-wrap items-start justify-between gap-3'>
          <div className='min-w-0 space-y-1'>
            <div className='flex items-center gap-2'>
              {hasChildren ? (
                <button
                  type='button'
                  onClick={() => setCollapsed((current) => !current)}
                  className='inline-flex size-6 shrink-0 items-center justify-center rounded-md border border-[#e0e0e0] hover:bg-muted/50'
                  aria-label={collapsed ? 'Expand step' : 'Collapse step'}
                >
                  <ChevronDown
                    className={cn(
                      'size-4 transition-transform',
                      collapsed && '-rotate-90'
                    )}
                  />
                </button>
              ) : (
                <span className='inline-block size-6 shrink-0' />
              )}
              <span className='text-sm font-semibold tracking-tight'>
                {node.step_path}
              </span>
              <Badge
                variant='outline'
                className={cn('capitalize', statusColor(node.status))}
              >
                {node.status}
              </Badge>
            </div>
            <div className='text-[17px] font-medium tracking-tight'>
              {node.step_name}
            </div>
          </div>
          <div className='space-y-1 text-right text-xs text-muted-foreground'>
            <div>{formatDateTime(node.start_time)}</div>
            <div>{formatDuration(node.duration)}</div>
          </div>
        </div>
      </div>

      {hasChildren && !collapsed && (
        <div className='mt-3 flex flex-col gap-3'>
          {node.children.map((child) => (
            <StepTreeNodeView
              key={child.step_path}
              node={child}
              depth={depth + 1}
              activePath={activePath}
            />
          ))}
        </div>
      )}
    </div>
  )
}
