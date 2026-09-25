'use client'

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog'
import { Button } from '@/components/ui/button'
import { rekeyPremiumProfiles } from '@/lib/api-endpoints'
import { clientApiFetcher } from '@/lib/client'
import { errorToast } from '@/lib/toasts'
import { useTranslations } from 'next-intl'
import { useRouter } from 'next/navigation'
import React from 'react'
import { toast } from 'sonner'

// Moves profiles of licensed nicknames to the Mojang UUID NavAuth gives them in game.
export default function RekeyPremiumButton() {
  const t = useTranslations('admin.profiles.rekey')
  const router = useRouter()
  const [isPending, startTransition] = React.useTransition()

  const handleRekey = () => {
    startTransition(async () => {
      try {
        const report = await rekeyPremiumProfiles(clientApiFetcher)
        toast.success(
          t('done', {
            rekeyed: report.rekeyed.length,
            skipped: report.skipped.length,
            failed: report.failed.length,
            unchecked: report.unchecked,
          }),
        )
        router.refresh()
      } catch (error) {
        errorToast(t('failed'), error)
      }
    })
  }

  return (
    <AlertDialog>
      <AlertDialogTrigger asChild>
        <Button variant="outline" disabled={isPending}>
          {t('button')}
        </Button>
      </AlertDialogTrigger>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>{t('confirmTitle')}</AlertDialogTitle>
          <AlertDialogDescription>
            {t('confirmDescription')}
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>{t('cancel')}</AlertDialogCancel>
          <AlertDialogAction onClick={handleRekey}>
            {t('confirm')}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
