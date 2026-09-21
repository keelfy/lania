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
import { RefreshCwIcon } from 'lucide-react'
import { useTranslations } from 'next-intl'
import React from 'react'
import { toast } from 'sonner'

type Props = {
  profileId: string
  asAdmin?: boolean
}

// Runs the resync. Admins can resync any profile, and there is no cooldown for them.
function useResync({ profileId, asAdmin }: Props) {
  const t = useTranslations('profileResync')
  const [isPending, startTransition] = React.useTransition()

  const resync = () => {
    startTransition(async () => {
      try {
        await (asAdmin ? resyncProfileAsAdmin : resyncProfile)(
          clientApiFetcher,
          profileId,
        )
        toast.success(t('success'))
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
