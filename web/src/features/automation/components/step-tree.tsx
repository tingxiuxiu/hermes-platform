import { useState } from 'react'
import { ChevronDown } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Badge } from '@/components/ui/badge'
import type { LiveStepNode } from '../data/schema'
import { formatDateTime, formatDuration, statusColor } from '../data/utils'

type StepTreeProps = {
  nodes: LiveStepNode[]
}

export function StepTree({ nodes }: StepTreeProps) {
  if (!nodes.length) {
    return (
      <div className='py-10 text-center text-sm text-muted-foreground'>
        还没有步骤上报。
      </div>
    )
  }

  return (
    <div className='space-y-3'>
      {nodes.map((node) => (
        <StepTreeNodeView key={node.step_path} node={node} depth={0} />
      ))}
    </div>
  )
}

function StepTreeNodeView({
  node,
  depth,
}: {
  node: LiveStepNode
  depth: number
}) {
  const [collapsed, setCollapsed] = useState(false)
  const hasChildren = node.children.length > 0
  const running = node.status === 'running'

  return (
    <div className={cn(depth > 0 && 'ml-4 border-l border-[#e0e0e0] pl-4')}>
      <div
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
        <div className='mt-3 space-y-3'>
          {node.children.map((child) => (
            <StepTreeNodeView
              key={child.step_path}
              node={child}
              depth={depth + 1}
            />
          ))}
        </div>
      )}
    </div>
  )
}
