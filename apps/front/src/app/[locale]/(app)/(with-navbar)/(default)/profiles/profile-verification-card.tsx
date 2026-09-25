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
import VerifiedBadge from '@/components/ui/verified-badge'
import {
  confirmProfileVerification,
  startProfileVerification,
} from '@/lib/api-endpoints'
import { clientApiFetcher } from '@/lib/client'
import { errorToast } from '@/lib/toasts'
import { Profile } from '@/models/profile'
import { BadgeCheckIcon } from 'lucide-react'
import { useTranslations } from 'next-intl'
import { useRouter } from 'next/navigation'
import React from 'react'
import { toast } from 'sonner'

type Props = {
  profile: Profile
  // Where the player joins to get the code.
  serverAddress?: string
}

const CODE_LENGTH = 6

// The optional license checkmark. Only a profile keyed to its Mojang UUID can get it: the proxy shows the code
// to the licensed player, and the owner types it here.
export default function ProfileVerificationCard({
  profile,
  serverAddress,
}: Props) {
  const t = useTranslations('profiles.verification')
  const router = useRouter()
  const [open, setOpen] = React.useState(false)
  const [code, setCode] = React.useState('')
  const [isStarting, startStarting] = React.useTransition()
  const [isConfirming, startConfirming] = React.useTransition()
  // Until the admin rekeys the profile to the Mojang UUID, the proxy cannot match the licensed login to it.
  const rekeyed = profile.mojangUuid === profile.mcUuid

  const start = () => {
    startStarting(async () => {
      try {
        await startProfileVerification(clientApiFetcher, profile.id)
        setCode('')
        setOpen(true)
      } catch (error) {
        errorToast(t('startFailed'), error)
      }
    })
  }

  const confirm = (event: React.FormEvent) => {
    event.preventDefault()
    startConfirming(async () => {
      try {
        await confirmProfileVerification(clientApiFetcher, profile.id, code)
        setOpen(false)
        toast.success(t('success'))
        router.refresh()
      } catch (error) {
        errorToast(t('confirmFailed'), error)
      }
    })
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2 text-lg">
          {t('title')}
          {profile.verified && <VerifiedBadge />}
        </CardTitle>
        <CardDescription>
          {t(
            profile.verified
              ? 'verified'
              : rekeyed
                ? 'description'
                : 'notRekeyed',
          )}
        </CardDescription>
        {!profile.verified && (
          <CardAction>
            <Button
              variant="outline"
              size="sm"
              onClick={start}
              disabled={!rekeyed || isStarting}
            >
              <BadgeCheckIcon className="size-4" />
              {t('start')}
            </Button>
          </CardAction>
        )}
      </CardHeader>
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t('title')}</DialogTitle>
            <DialogDescription>{t('dialogDescription')}</DialogDescription>
          </DialogHeader>
          <ol className="list-decimal space-y-1 pl-5 text-sm">
            <li>
              {t.rich('steps.join', {
                username: profile.username,
                address: serverAddress ?? 'lania.network',
                b: (chunks) => <b>{chunks}</b>,
              })}
            </li>
            <li>{t('steps.kick')}</li>
            <li>{t('steps.type')}</li>
          </ol>
          <form onSubmit={confirm} className="flex gap-2">
            <Input
              value={code}
              onChange={(event) => setCode(event.target.value.toUpperCase())}
              maxLength={CODE_LENGTH}
              placeholder="ABC123"
              autoComplete="one-time-code"
              aria-label={t('code')}
              className="font-mono tracking-widest uppercase"
            />
            <Button
              type="submit"
              disabled={code.length !== CODE_LENGTH || isConfirming}
            >
              {t('confirm')}
            </Button>
          </form>
          <DialogFooter className="text-muted-foreground items-center text-xs sm:justify-between">
            <span>{t('expires')}</span>
            <Button
              variant="link"
              size="sm"
              className="h-auto p-0 text-xs"
              onClick={start}
              disabled={isStarting}
            >
              {t('restart')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </Card>
  )
}
