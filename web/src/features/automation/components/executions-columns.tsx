import { Link } from '@tanstack/react-router'
import { type ColumnDef } from '@tanstack/react-table'
import { ExternalLink, PanelsTopLeft } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { DataTableColumnHeader } from '@/components/data-table'
import { LongText } from '@/components/long-text'
import type { ExecutionItem } from '../data/schema'
import { formatDateTime, formatDuration, shortUid, statusColor } from '../data/utils'

export function getExecutionColumns(): ColumnDef<ExecutionItem>[] {
  return [
    {
      accessorKey: 'id',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title='ID' />
      ),
      cell: ({ row }) => <div className='font-medium'>{row.original.id}</div>,
      meta: { className: 'w-20' },
    },
    {
      accessorKey: 'job_name',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title='Job Name' />
      ),
      cell: ({ row }) => (
        <div className='space-y-1'>
          <LongText className='max-w-72 font-medium'>
            {row.original.job_name}
          </LongText>
          <div className='text-xs text-muted-foreground'>
            {shortUid(row.original.build_uid)}
          </div>
        </div>
      ),
      enableHiding: false,
      meta: { className: 'min-w-72' },
    },
    {
      accessorKey: 'status',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title='Status' />
      ),
      cell: ({ row }) => (
        <Badge
          variant='outline'
          className={cn('capitalize', statusColor(row.original.status))}
        >
          {row.original.status}
        </Badge>
      ),
      filterFn: (row, id, value) => value.includes(row.getValue(id)),
      enableSorting: false,
      meta: { className: 'w-32' },
    },
    {
      accessorKey: 'start_time',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title='Started At' />
      ),
      cell: ({ row }) => formatDateTime(row.original.start_time),
      meta: { className: 'w-48' },
    },
    {
      accessorKey: 'duration',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title='Duration' />
      ),
      cell: ({ row }) => formatDuration(row.original.duration),
      meta: { className: 'w-28' },
    },
    {
      accessorKey: 'pass_count',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title='Passed' />
      ),
      cell: ({ row }) => row.original.pass_count ?? '--',
      meta: { className: 'w-24' },
    },
    {
      accessorKey: 'failure_count',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title='Failed' />
      ),
      cell: ({ row }) => row.original.failure_count ?? '--',
      meta: { className: 'w-24' },
    },
    {
      accessorKey: 'skipped_count',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title='Skipped' />
      ),
      cell: ({ row }) => row.original.skipped_count ?? '--',
      meta: { className: 'w-24' },
    },
    {
      id: 'actions',
      header: 'Action',
      cell: ({ row }) => (
        <div className='flex flex-wrap justify-end gap-2'>
          <Button size='sm' asChild>
            <Link
              to='/automation/executions/$buildUid'
              params={{ buildUid: row.original.build_uid }}
            >
              <PanelsTopLeft className='size-4' />
              Session
            </Link>
          </Button>
          {row.original.job_url ? (
            <Button variant='ghost' size='sm' asChild>
              <a href={row.original.job_url} target='_blank' rel='noreferrer'>
                <ExternalLink className='size-4' />
                Job
              </a>
            </Button>
          ) : null}
        </div>
      ),
      enableSorting: false,
      enableHiding: false,
      meta: { className: 'w-44 text-right' },
    },
  ]
}
