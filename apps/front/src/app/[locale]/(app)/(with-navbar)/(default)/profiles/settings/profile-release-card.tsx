'use client'

import {
  AlertDialog,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { releaseProfile } from '@/lib/api-endpoints'
import { clientApiFetcher } from '@/lib/client'
import { errorToast } from '@/lib/toasts'
import { Profile } from '@/models/profile'
import { useLocale, useTranslations } from 'next-intl'
import { useRouter } from 'next/navigation'
import React from 'react'
import { toast } from 'sonner'

type Props = {
  profile: Profile
}

// Gives the profile up, so anybody can take the nickname. The owner types the nickname to confirm.
export default function ProfileReleaseCard({ profile }: Props) {
  const t = useTranslations('profiles.release')
  const locale = useLocale()
  const router = useRouter()
  const [confirmation, setConfirmation] = React.useState('')
  const [isPending, startTransition] = React.useTransition()
  const confirmed =
    confirmation.trim().toLowerCase() === profile.username.toLowerCase()

  const handleRelease = () => {
    startTransition(async () => {
      try {
        await releaseProfile(clientApiFetcher, profile.id)
        toast.success(t('success', { username: profile.username }))
        // The profile is not the user's anymore, so its settings are gone too.
        router.push(`/${locale}/profiles`)
        router.refresh()
      } catch (error) {
        errorToast(t('error'), error)
      }
    })
  }

  return (
    <Card className="border-destructive/50">
      <CardHeader>
        <CardTitle className="text-destructive text-lg">{t('title')}</CardTitle>
        <CardDescription>{t('description')}</CardDescription>
      </CardHeader>
      <CardContent className="flex flex-col gap-4">
        <ul className="text-muted-foreground list-disc space-y-1 pl-5 text-sm">
          <li>{t('consequences.cosmetics')}</li>
          <li>{t('consequences.roles')}</li>
          <li>{t('consequences.stays')}</li>
          <li>{t('consequences.limit')}</li>
        </ul>
        <AlertDialog
          onOpenChange={(open) => {
            if (!open) setConfirmation('')
          }}
        >
          <AlertDialogTrigger asChild>
            <Button variant="destructive" className="self-start">
              {t('button')}
            </Button>
          </AlertDialogTrigger>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>
                {t('confirmTitle', { username: profile.username })}
              </AlertDialogTitle>
              <AlertDialogDescription>
                {t('confirmDescription', { username: profile.username })}
              </AlertDialogDescription>
            </AlertDialogHeader>
            <Input
              value={confirmation}
              onChange={(event) => setConfirmation(event.target.value)}
              placeholder={profile.username}
              autoComplete="off"
              disabled={isPending}
            />
            <AlertDialogFooter>
              <AlertDialogCancel disabled={isPending}>
                {t('cancel')}
              </AlertDialogCancel>
              {/* A plain button, so the dialog stays open while the request runs. */}
              <Button
                variant="destructive"
                disabled={!confirmed || isPending}
                onClick={handleRelease}
              >
                {t('confirm')}
              </Button>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      </CardContent>
    </Card>
  )
}
