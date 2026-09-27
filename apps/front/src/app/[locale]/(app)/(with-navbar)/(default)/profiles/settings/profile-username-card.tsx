'use client'

import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardAction,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { changeProfileUsername } from '@/lib/api-endpoints'
import { clientApiFetcher } from '@/lib/client'
import { errorToast } from '@/lib/toasts'
import { Profile } from '@/models/profile'
import { PencilIcon, TriangleAlertIcon } from 'lucide-react'
import { useLocale, useTranslations } from 'next-intl'
import { useRouter } from 'next/navigation'
import React from 'react'
import { toast } from 'sonner'

type Props = {
  profile: Profile
}

// What Minecraft accepts as a nickname; the API checks the same.
const USERNAME_PATTERN = /^[a-zA-Z0-9_]{3,16}$/

// Changes the nickname of the profile. A licensed profile follows the name of its Minecraft account, which the site
// also picks up by itself. An unlicensed one becomes a new player in game, so the dialog says what stays behind.
export default function ProfileUsernameCard({ profile }: Props) {
  const t = useTranslations('profiles.username')
  const locale = useLocale()
  const router = useRouter()
  const [open, setOpen] = React.useState(false)
  const [username, setUsername] = React.useState('')
  const [isSaving, startSaving] = React.useTransition()

  const licensed = !!profile.mojangUuid && profile.mojangUuid === profile.mcUuid
  const availableAt = licensed ? undefined : profile.usernameChangeAvailableAt
  const availableDate =
    availableAt &&
    new Date(availableAt).toLocaleDateString(locale, {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      // The same on the server and in the browser, and the same as in the error of the API.
      timeZone: 'UTC',
    })
  const trimmed = username.trim()
  const valid = USERNAME_PATTERN.test(trimmed) && trimmed !== profile.username

  const openDialog = () => {
    setUsername('')
    setOpen(true)
  }

  const submit = (event: React.FormEvent) => {
    event.preventDefault()
    if (!valid) return
    startSaving(async () => {
      try {
        await changeProfileUsername(clientApiFetcher, profile.id, trimmed)
        setOpen(false)
        toast.success(t('success', { username: trimmed }))
        router.refresh()
      } catch (error) {
        errorToast(t('failed'), error)
      }
    })
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-lg">{t('title')}</CardTitle>
        <CardDescription>
          {availableDate
            ? t('availableAt', { date: availableDate })
            : t(licensed ? 'licensedDescription' : 'description')}
        </CardDescription>
        <CardAction>
          <Button
            variant="outline"
            size="sm"
            onClick={openDialog}
            disabled={!!availableDate}
          >
            <PencilIcon className="size-4" />
            {t('open')}
          </Button>
        </CardAction>
      </CardHeader>
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>
              {t('dialogTitle', { username: profile.username })}
            </DialogTitle>
            <DialogDescription>
              {t(licensed ? 'licensedDialogDescription' : 'dialogDescription')}
            </DialogDescription>
          </DialogHeader>
          {!licensed && (
            <Alert variant="destructive">
              <TriangleAlertIcon />
              <AlertTitle>{t('warning.title')}</AlertTitle>
              <AlertDescription>
                <ul className="list-disc space-y-1 pl-4">
                  <li>{t('warning.progress')}</li>
                  <li>{t('warning.password')}</li>
                  <li>
                    {t('warning.reserved', { username: profile.username })}
                  </li>
                  {!!profile.usernameChangeCooldownDays && (
                    <li>
                      {t('warning.cooldown', {
                        days: profile.usernameChangeCooldownDays,
                      })}
                    </li>
                  )}
                </ul>
              </AlertDescription>
            </Alert>
          )}
          <form id="profile-username-form" onSubmit={submit}>
            <Input
              value={username}
              onChange={(event) => setUsername(event.target.value)}
              maxLength={16}
              placeholder={profile.username}
              autoComplete="off"
              spellCheck={false}
              aria-label={t('label')}
              aria-invalid={trimmed.length > 0 && !valid}
              className="font-mono"
            />
            <p className="text-muted-foreground mt-2 text-xs">{t('hint')}</p>
          </form>
          <DialogFooter>
            <Button
              type="submit"
              form="profile-username-form"
              variant={licensed ? 'default' : 'destructive'}
              disabled={!valid || isSaving}
            >
              {t('submit')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </Card>
  )
}
