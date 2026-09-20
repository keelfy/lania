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
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { releaseProfileOwner, transferProfileOwner } from '@/lib/api-endpoints'
import { clientApiFetcher } from '@/lib/client'
import { errorToast } from '@/lib/toasts'
import { AdminUser } from '@/models/admin'
import { useTranslations } from 'next-intl'
import Link from 'next/link'
import { useRouter } from 'next/navigation'
import React from 'react'
import { toast } from 'sonner'

type Props = {
  profileId: string
  owner: AdminUser | undefined
  locale: string
}

export default function OwnerCard({ profileId, owner, locale }: Props) {
  const t = useTranslations('admin.profiles.owner')
  const router = useRouter()
  const [email, setEmail] = React.useState('')
  const [isPending, startTransition] = React.useTransition()

  const handleTransfer = (event: React.FormEvent) => {
    event.preventDefault()
    if (isPending) return

    startTransition(async () => {
      try {
        const profile = await transferProfileOwner(
          clientApiFetcher,
          profileId,
          email.trim(),
        )
        toast.success(
          t('transferred', { email: profile.owner?.email ?? email.trim() }),
        )
        setEmail('')
        router.refresh()
      } catch (error) {
        errorToast(t('transferFailed'), error)
      }
    })
  }

  const handleRelease = () => {
    startTransition(async () => {
      try {
        await releaseProfileOwner(clientApiFetcher, profileId)
        toast.success(t('released'))
        router.refresh()
      } catch (error) {
        errorToast(t('releaseFailed'), error)
      }
    })
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-lg">{t('title')}</CardTitle>
        <CardDescription>{t('description')}</CardDescription>
      </CardHeader>
      <CardContent className="flex flex-col gap-4">
        <div className="flex flex-wrap items-center justify-between gap-2">
          {owner ? (
            <Link
              href={`/${locale}/admin/users/${owner.id}`}
              className="font-medium hover:underline"
            >
              {owner.email || owner.id}
            </Link>
          ) : (
            <p className="text-muted-foreground">{t('noOwner')}</p>
          )}
          {owner && (
            <AlertDialog>
              <AlertDialogTrigger asChild>
                <Button variant="outline" size="sm" disabled={isPending}>
                  {t('release')}
                </Button>
              </AlertDialogTrigger>
              <AlertDialogContent>
                <AlertDialogHeader>
                  <AlertDialogTitle>
                    {t('releaseConfirmTitle')}
                  </AlertDialogTitle>
                  <AlertDialogDescription>
                    {t('releaseConfirmDescription')}
                  </AlertDialogDescription>
                </AlertDialogHeader>
                <AlertDialogFooter>
                  <AlertDialogCancel>{t('cancel')}</AlertDialogCancel>
                  <AlertDialogAction onClick={handleRelease}>
                    {t('release')}
                  </AlertDialogAction>
                </AlertDialogFooter>
              </AlertDialogContent>
            </AlertDialog>
          )}
        </div>
        <form onSubmit={handleTransfer} className="flex flex-wrap gap-2">
          <Input
            type="email"
            required
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder={t('emailPlaceholder')}
            aria-label={t('emailPlaceholder')}
            className="flex-1 sm:min-w-64"
          />
          <Button type="submit" disabled={isPending || email.trim() === ''}>
            {t('transfer')}
          </Button>
        </form>
      </CardContent>
    </Card>
  )
}
