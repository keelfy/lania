'use client'

import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { resyncProfile, resyncProfileAsAdmin } from '@/lib/api-endpoints'
import { clientApiFetcher } from '@/lib/client'
import { errorToast } from '@/lib/toasts'
import { ProfileResync } from '@/models/profile'
import { CheckIcon, RefreshCwIcon, XIcon } from 'lucide-react'
import { useTranslations } from 'next-intl'
import React from 'react'
import { toast } from 'sonner'

type Props = {
  profileId: string
  asAdmin?: boolean
}

// Time to read a report with failures. A success goes away as fast as any other one.
const FAILED_TOAST_DURATION = 15_000

// Tells for every season whether its server got everything, and what failed when it did not.
export function ResyncReport({ resync }: { resync: ProfileResync }) {
  const t = useTranslations('profileResync')

  return (
    <ul className="flex flex-col gap-1">
      {resync.seasons.map((season) => (
        <li key={season.seasonId}>
          <div className="flex items-center gap-1.5 font-medium">
            {season.ok ? (
              <CheckIcon className="size-4 shrink-0 text-green-500" />
            ) : (
              <XIcon className="size-4 shrink-0 text-red-500" />
            )}
            {season.seasonName}
          </div>
          {season.parts
            .filter((part) => !part.ok)
            .map((part) => (
              <div key={part.part} className="ml-5.5 text-xs opacity-80">
                {t(`parts.${part.part}`)}
                {part.error ? `: ${part.error}` : null}
              </div>
            ))}
        </li>
      ))}
    </ul>
  )
}

// Shows the report as a success, a warning when only some seasons failed, or an error when all of them did.
export function showResync(resync: ProfileResync, t: (key: string) => string) {
  if (resync.seasons.length === 0) {
    toast.success(t('success'), { description: t('noServers') })
    return
  }

  const description = <ResyncReport resync={resync} />
  if (resync.ok) {
    toast.success(t('success'), { description })
  } else if (resync.seasons.some((season) => season.ok)) {
    toast.warning(t('partial'), {
      description,
      duration: FAILED_TOAST_DURATION,
    })
  } else {
    toast.error(t('failed'), { description, duration: FAILED_TOAST_DURATION })
  }
}

// Runs the resync. Admins can resync any profile, and there is no cooldown for them.
function useResync({ profileId, asAdmin }: Props) {
  const t = useTranslations('profileResync')
  const [isPending, startTransition] = React.useTransition()

  const resync = () => {
    startTransition(async () => {
      try {
        const resync = await (asAdmin ? resyncProfileAsAdmin : resyncProfile)(
          clientApiFetcher,
          profileId,
        )
        showResync(resync, t)
      } catch (error) {
        errorToast(t('error'), error)
      }
    })
  }

  return { resync, isPending }
}

// A small button for a place where the card is too big.
export function ProfileResyncButton(props: Props) {
  const t = useTranslations('profileResync')
  const { resync, isPending } = useResync(props)

  return (
    <Button
      variant="outline"
      size="icon"
      title={t('description')}
      aria-label={t('button')}
      onClick={resync}
      disabled={isPending}
    >
      <RefreshCwIcon className={isPending ? 'animate-spin' : undefined} />
    </Button>
  )
}

export default function ProfileResyncCard(props: Props) {
  const t = useTranslations('profileResync')
  const { resync, isPending } = useResync(props)

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-lg">{t('title')}</CardTitle>
        <CardDescription>{t('description')}</CardDescription>
      </CardHeader>
      <CardContent>
        <Button
          variant="outline"
          className="w-full sm:w-fit"
          onClick={resync}
          disabled={isPending}
        >
          <RefreshCwIcon className={isPending ? 'animate-spin' : undefined} />
          {t('button')}
        </Button>
      </CardContent>
    </Card>
  )
}
