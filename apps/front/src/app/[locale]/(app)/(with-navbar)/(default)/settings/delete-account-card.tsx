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
import { deleteAccount } from '@/lib/api-endpoints'
import { clientApiFetcher } from '@/lib/client'
import { errorToast } from '@/lib/toasts'
import { useLocale, useTranslations } from 'next-intl'
import { usePathname, useRouter } from 'next/navigation'
import React from 'react'

// The API answers with this message when the sign in is older than the Kratos privileged session.
const SESSION_REFRESH_REQUIRED = 'session_refresh_required'

export default function DeleteAccountCard() {
  const t = useTranslations('accountSettings.deleteAccount')
  const locale = useLocale()
  const router = useRouter()
  const pathname = usePathname()
  const [confirmation, setConfirmation] = React.useState('')
  const [isPending, startTransition] = React.useTransition()
  const confirmed =
    confirmation.trim().toLowerCase() === t('confirmWord').toLowerCase()

  const handleDelete = () => {
    startTransition(async () => {
      try {
        await deleteAccount(clientApiFetcher)
        // The session is gone with the identity, so the whole app is loaded again.
        // eslint-disable-next-line @next/next/no-location-assign-relative-destination
        window.location.href = `/${locale}`
      } catch (error) {
        if (
          error instanceof Error &&
          error.message.trim() === SESSION_REFRESH_REQUIRED
        ) {
          router.push(
            `/${locale}/auth/login?refresh=true&goto=${encodeURIComponent(pathname)}`,
          )
          return
        }
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
          <li>{t('consequences.username')}</li>
          <li>{t('consequences.access')}</li>
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
              <AlertDialogTitle>{t('confirmTitle')}</AlertDialogTitle>
              <AlertDialogDescription>
                {t('confirmDescription', { word: t('confirmWord') })}
              </AlertDialogDescription>
            </AlertDialogHeader>
            <Input
              value={confirmation}
              onChange={(event) => setConfirmation(event.target.value)}
              placeholder={t('confirmWord')}
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
                onClick={handleDelete}
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
