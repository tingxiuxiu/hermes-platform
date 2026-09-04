import { Link } from '@tanstack/react-router'
import { type ColumnDef } from '@tanstack/react-table'
import { ExternalLink, ListTree } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { DataTableColumnHeader } from '@/components/data-table'
import { LongText } from '@/components/long-text'
import type { JobItem } from '../data/schema'
import { formatDateTime, statusColor, stringifyParams } from '../data/utils'

export const jobsColumns: ColumnDef<JobItem>[] = [
  {
    accessorKey: 'id',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Job ID' />
    ),
    cell: ({ row }) => <div className='font-medium'>{row.original.id}</div>,
    meta: { className: 'w-24' },
  },
  {
    accessorKey: 'job_name',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Job Name' />
    ),
    cell: ({ row }) => (
      <LongText className='max-w-64 font-medium'>
        {row.original.job_name}
      </LongText>
    ),
    enableHiding: false,
    meta: { className: 'min-w-64' },
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
    accessorKey: 'last_build_status',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Last Build' />
    ),
    cell: ({ row }) => {
      const status = row.original.last_build_status
      if (!status) return <span className='text-muted-foreground'>--</span>
      return (
        <Badge
          variant='outline'
          className={cn('capitalize', statusColor(status))}
        >
          {status}
        </Badge>
      )
    },
    filterFn: (row, id, value) => value.includes(row.getValue(id)),
    enableSorting: false,
    meta: { className: 'w-36' },
  },
  {
    accessorKey: 'last_build_number',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Build No.' />
    ),
    cell: ({ row }) => row.original.last_build_number ?? '--',
    meta: { className: 'w-28' },
  },
  {
    accessorKey: 'pipeline_params',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Params' />
    ),
    cell: ({ row }) => (
      <LongText className='max-w-48 text-xs text-muted-foreground'>
        {stringifyParams(row.original.pipeline_params)}
      </LongText>
    ),
    enableSorting: false,
    meta: { className: 'min-w-48' },
  },
  {
    accessorKey: 'sync_at',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Synced At' />
    ),
    cell: ({ row }) => formatDateTime(row.original.sync_at),
    meta: { className: 'w-48' },
  },
  {
    id: 'actions',
    header: 'Action',
    cell: ({ row }) => (
      <div className='flex flex-wrap justify-end gap-2'>
        <Button variant='outline' size='sm' asChild>
          {row.original.last_build_uid ? (
            <Link
              to='/automation/executions/$buildUid'
              params={{ buildUid: row.original.last_build_uid }}
            >
              <ListTree className='size-4' />
              Session
            </Link>
          ) : (
            <Link
              to='/automation/executions'
              search={{
                page: 1,
                pageSize: 10,
                jobName: row.original.job_name,
                status: [],
              }}
            >
              <ListTree className='size-4' />
              Executions
            </Link>
          )}
        </Button>
        <Button variant='ghost' size='sm' asChild>
          <a href={row.original.job_url} target='_blank' rel='noreferrer'>
            <ExternalLink className='size-4' />
            Jenkins
          </a>
        </Button>
      </div>
    ),
    enableSorting: false,
    enableHiding: false,
    meta: { className: 'w-56 text-right' },
  },
]
