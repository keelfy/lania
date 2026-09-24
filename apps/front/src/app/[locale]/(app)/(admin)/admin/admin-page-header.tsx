import { Button } from '@/components/ui/button'
import Link from 'next/link'
import React from 'react'

type Props = {
  title?: React.ReactNode
  description?: React.ReactNode
  actions?: React.ReactNode
  backHref?: string
  backLabel?: React.ReactNode
}

// Replaces the old AdminShell: the section nav now lives in the sidebar, so
// this only renders the per-page title/back-link row.
export default function AdminPageHeader({
  title,
  description,
  actions,
  backHref,
  backLabel,
}: Props) {
  return (
    <div className="flex flex-col gap-2">
      {backHref && (
        <Button asChild variant="link" className="h-auto w-fit p-0">
          <Link href={backHref}>{backLabel}</Link>
        </Button>
      )}
      {(title || actions) && (
        <div className="flex flex-wrap items-center justify-between gap-2">
          <div>
            {title && (
              <h1 className="text-2xl font-bold tracking-tight">{title}</h1>
            )}
            {description && (
              <p className="text-muted-foreground text-sm">{description}</p>
            )}
          </div>
          {actions}
        </div>
      )}
    </div>
  )
}
