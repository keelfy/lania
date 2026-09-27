'use client'

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
import SeasonSelect from '@/components/ui/season-select'
import { setProfilePassword } from '@/lib/api-endpoints'
import { clientApiFetcher } from '@/lib/client'
import { errorToast } from '@/lib/toasts'
import { Profile } from '@/models/profile'
import { Season } from '@/models/season'
import { KeyRoundIcon } from 'lucide-react'
import { useTranslations } from 'next-intl'
import React from 'react'
import { toast } from 'sonner'

type Props = {
  profile: Profile
  // Running seasons the profile has access to.
  seasons: Season[]
}

// The default password rules of the login plugin, which the API checks too.
const PASSWORD_PATTERN = /^(?=.*\p{Ll})(?=.*\p{Nd})\S{5,32}$/u

const PLAYER_NOT_REGISTERED = 'player_not_registered'

// Replaces the in-game password of an unlicensed profile: for a forgotten password, or when someone registered
// in game under the nickname before the owner did.
export default function ProfilePasswordCard({ profile, seasons }: Props) {
  const t = useTranslations('profiles.password')
  const [open, setOpen] = React.useState(false)
  const [seasonId, setSeasonId] = React.useState<string>()
  const [password, setPassword] = React.useState('')
  const [isSaving, startSaving] = React.useTransition()

  const selectedSeasonId =
    seasonId ?? (seasons.find((season) => season.isPrimary) ?? seasons[0])?.id
  const valid = PASSWORD_PATTERN.test(password) && !!selectedSeasonId

  const openDialog = () => {
    setPassword('')
    setOpen(true)
  }

  const submit = (event: React.FormEvent) => {
    event.preventDefault()
    if (!valid || !selectedSeasonId) return
    startSaving(async () => {
      try {
        await setProfilePassword(
          clientApiFetcher,
          profile.id,
          selectedSeasonId,
          password,
        )
        setOpen(false)
        toast.success(t('success'), { description: t('successHint') })
      } catch (error) {
        if (
          error instanceof Error &&
          error.message.trim() === PLAYER_NOT_REGISTERED
        ) {
          toast.error(t('notRegistered'), {
            description: t('notRegisteredHint'),
          })
          return
        }
        errorToast(t('failed'), error)
      }
    })
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-lg">{t('title')}</CardTitle>
        <CardDescription>
          {t(seasons.length > 0 ? 'description' : 'noSeasons')}
        </CardDescription>
        <CardAction>
          <Button
            variant="outline"
            size="sm"
            onClick={openDialog}
            disabled={seasons.length === 0}
          >
            <KeyRoundIcon className="size-4" />
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
            <DialogDescription>{t('dialogDescription')}</DialogDescription>
          </DialogHeader>
          <ul className="text-muted-foreground list-disc space-y-1 pl-4 text-sm">
            <li>{t('cases.forgotten')}</li>
            <li>{t('cases.taken')}</li>
          </ul>
          <form
            id="profile-password-form"
            onSubmit={submit}
            className="space-y-3"
          >
            {seasons.length > 1 && (
              <SeasonSelect
                seasons={seasons}
                selectedSeasonId={selectedSeasonId}
                onSelectSeasonId={setSeasonId}
                className="w-full"
                aria-label={t('season')}
              />
            )}
            <div>
              <Input
                type="password"
                value={password}
                onChange={(event) => setPassword(event.target.value)}
                maxLength={32}
                autoComplete="new-password"
                aria-label={t('label')}
                placeholder={t('label')}
                aria-invalid={password.length > 0 && !valid}
              />
              <p className="text-muted-foreground mt-2 text-xs">{t('hint')}</p>
            </div>
          </form>
          <DialogFooter>
            <Button
              type="submit"
              form="profile-password-form"
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
