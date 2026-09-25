'use client'

import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
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
import { Separator } from '@/components/ui/separator'
import VerifiedBadge from '@/components/ui/verified-badge'
import {
  confirmProfileVerification,
  startProfileVerification,
} from '@/lib/api-endpoints'
import { clientApiFetcher } from '@/lib/client'
import { errorToast } from '@/lib/toasts'
import { cn } from '@/lib/utils'
import { Profile } from '@/models/profile'
import {
  BadgeCheckIcon,
  ExternalLinkIcon,
  MinusIcon,
  PlusIcon,
} from 'lucide-react'
import { useTranslations } from 'next-intl'
import { useRouter } from 'next/navigation'
import React from 'react'
import { toast } from 'sonner'
import { useSelectedProfile } from './use-selected-profile'

type Props = {
  profiles: Profile[]
  // Where the player joins to get the code.
  serverAddress?: string
  className?: string
}

const CODE_LENGTH = 6
const BUY_URL =
  'https://www.minecraft.net/store/minecraft-java-bedrock-edition-pc'

// The license of the selected profile: the optional checkmark for a licensed nickname, and what playing without
// a license means for a free one.
export default function ProfileLicenseCard({
  profiles,
  serverAddress,
  className,
}: Props) {
  const profile = useSelectedProfile(profiles)
  if (!profile) return null
  return profile.mojangUuid ? (
    <ProfileVerificationCard
      key={profile.id}
      profile={profile}
      serverAddress={serverAddress}
      className={className}
    />
  ) : (
    <UnlicensedCard className={className} />
  )
}

function UnlicensedCard({ className }: { className?: string }) {
  const t = useTranslations('profiles.verification.unlicensed')
  return (
    <Card className={cn('gap-4', className)}>
      <CardHeader>
        <CardTitle className="text-lg">{t('title')}</CardTitle>
        <CardDescription>{t('description')}</CardDescription>
      </CardHeader>
      <CardContent className="flex flex-col gap-3 text-sm">
        <ul className="flex flex-col gap-2">
          <li className="flex gap-2">
            <PlusIcon className="mt-0.5 size-4 shrink-0 text-green-500" />
            {t('pros.access')}
          </li>
          {(['nickname', 'security', 'badge'] as const).map((con) => (
            <li key={con} className="flex gap-2">
              <MinusIcon className="text-destructive mt-0.5 size-4 shrink-0" />
              {t(`cons.${con}`)}
            </li>
          ))}
        </ul>
        <Separator />
        <p className="text-muted-foreground">{t('advice')}</p>
        <Button variant="outline" size="sm" asChild>
          <a href={BUY_URL} target="_blank" rel="noopener noreferrer">
            {t('buy')}
            <ExternalLinkIcon className="size-4" />
          </a>
        </Button>
      </CardContent>
    </Card>
  )
}

// The optional license checkmark. Only a profile keyed to its Mojang UUID can get it: the proxy shows the code
// to the licensed player, and the owner types it here.
function ProfileVerificationCard({
  profile,
  serverAddress,
  className,
}: {
  profile: Profile
  serverAddress?: string
  className?: string
}) {
  const t = useTranslations('profiles.verification')
  const router = useRouter()
  const [open, setOpen] = React.useState(false)
  const [code, setCode] = React.useState('')
  const [isStarting, startStarting] = React.useTransition()
  const [isConfirming, startConfirming] = React.useTransition()
  // Until the admin rekeys the profile to the Mojang UUID, the proxy cannot match the licensed login to it.
  const rekeyed = profile.mojangUuid === profile.mcUuid
  // Verification is what the site expects from the owner here, so the card stands out until it is done.
  const awaits = rekeyed && !profile.verified

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
    <Card
      className={cn(
        awaits && 'border-primary/60 ring-primary/20 ring-2',
        className,
      )}
    >
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
      </CardHeader>
      {!profile.verified && (
        <CardContent>
          <Button
            variant={awaits ? 'default' : 'outline'}
            size="sm"
            className="w-full"
            onClick={start}
            disabled={!rekeyed || isStarting}
          >
            <BadgeCheckIcon className="size-4" />
            {t('start')}
          </Button>
        </CardContent>
      )}
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
