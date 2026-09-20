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
  // Admins can resync any profile, and there is no cooldown for them.
  asAdmin?: boolean
}

export default function ProfileResyncCard({ profileId, asAdmin }: Props) {
  const t = useTranslations('profileResync')
  const [isPending, startTransition] = React.useTransition()

  const handleResync = () => {
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
          onClick={handleResync}
          disabled={isPending}
        >
          <RefreshCwIcon className={isPending ? 'animate-spin' : undefined} />
          {t('button')}
        </Button>
      </CardContent>
    </Card>
  )
}
