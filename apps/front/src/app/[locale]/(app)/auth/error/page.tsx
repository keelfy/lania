'use client'

import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import LaniaLogo from '@/components/ui/lania-logo'
import LoadingSpinner from '@/components/ui/loading-spinner'
import { Link } from '@/i18n/navigation'
import ory from '@/lib/ory'
import { FlowError } from '@ory/client-fetch'
import { useTranslations } from 'next-intl'
import { useSearchParams } from 'next/navigation'
import React from 'react'

export default function ErrorPage() {
  const t = useTranslations('auth.error')
  const id = useSearchParams().get('id') ?? ''
  const [flowError, setFlowError] = React.useState<FlowError>()
  const [loading, setLoading] = React.useState(Boolean(id))

  React.useEffect(() => {
    if (!id) return
    let cancelled = false
    ory
      .getFlowError({ id })
      .then((error) => !cancelled && setFlowError(error))
      // An expired error is not worth more than the general text.
      .catch(() => undefined)
      .finally(() => !cancelled && setLoading(false))
    return () => {
      cancelled = true
    }
  }, [id])

  return (
    <Card className="mx-auto max-w-sm flex-1">
      <CardHeader>
        <CardTitle className="flex items-center justify-between gap-2">
          {t('title')}
          <Link href="/" className="hover:opacity-80">
            <LaniaLogo />
          </Link>
        </CardTitle>
        <CardDescription>{t('description')}</CardDescription>
      </CardHeader>
      <CardContent className="flex flex-col gap-4">
        {loading ? (
          <div className="flex min-h-12 items-center justify-center">
            <LoadingSpinner />
          </div>
        ) : (
          flowError && (
            <details className="text-muted-foreground text-xs">
              <summary className="cursor-pointer">{t('details')}</summary>
              <pre className="mt-2 overflow-auto whitespace-pre-wrap">
                {JSON.stringify(flowError.error, null, 2)}
              </pre>
            </details>
          )
        )}
        <Button asChild>
          <Link href="/auth/login">{t('retry')}</Link>
        </Button>
      </CardContent>
    </Card>
  )
}
