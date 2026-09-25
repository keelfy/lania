'use client'

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import { getProfileDetails } from '@/lib/api-endpoints'
import { clientApiFetcher } from '@/lib/client'
import { formatPlaytime } from '@/lib/playtime'
import {
  PROFILE_ROLE_COLORS,
  PROFILE_STATUS_COLORS,
} from '@/lib/profile-colors'
import { initializeViewer } from '@/lib/skin-viewer'
import { errorToast } from '@/lib/toasts'
import { cn } from '@/lib/utils'
import { NameCosmetics, ProfileDetails } from '@/models/profile'
import {
  CheckIcon,
  CircleUserRoundIcon,
  ClockIcon,
  CopyIcon,
} from 'lucide-react'
import { useTranslations } from 'next-intl'
import { useTheme } from 'next-themes'
import { useTimeAgo } from 'next-timeago'
import React from 'react'
import McUsername from './mc-username'
import NamePrefixes from './name-prefixes'
import PlayerFace from './player-face'
import VerifiedBadge from './verified-badge'

type PlayerCardProps = React.ComponentProps<typeof Card> & {
  profileId: string | undefined
  username?: string
  nameCosmetics?: NameCosmetics
  locale?: string
}

export default function PlayerCard({
  profileId,
  username,
  nameCosmetics,
  locale,
  className,
  ...props
}: PlayerCardProps) {
  const t = useTranslations('playerCard')
  const { resolvedTheme } = useTheme()
  const { TimeAgo } = useTimeAgo()

  const [profile, setProfile] = React.useState<ProfileDetails>()

  React.useEffect(() => {
    if (profileId) {
      getProfileDetails(clientApiFetcher, profileId)
        .then(setProfile)
        .catch((error) => {
          errorToast(t('noProfile'), error)
          setProfile(undefined)
        })
    }
  }, [profileId])

  React.useEffect(() => {
    const skinUrl = profile?.mojangUuid
      ? `https://crafatar-pub.neodium.fr/skins/${profile.mojangUuid}`
      : '/images/steve_skin.png'
    const capeUrl = profile?.mojangUuid
      ? `https://crafatar-pub.neodium.fr/capes/${profile.mojangUuid}`
      : ''

    if (typeof window !== 'undefined' && profile?.id !== undefined) {
      initializeViewer(
        skinUrl,
        capeUrl,
        profile?.isSlimModel ?? false,
        resolvedTheme === 'dark' ? 'dark' : 'light',
      )
    }
  }, [profile?.id, profile?.mojangUuid, profile?.isSlimModel, resolvedTheme])

  const status = React.useMemo(
    () => (profile?.isOnline ? 'online' : 'offline'),
    [profile?.isOnline],
  )

  const playtime = React.useMemo(
    () => formatPlaytime(profile?.playtime ?? 0),
    [profile?.playtime],
  )

  const displayName = username ?? profile?.username
  const [copied, setCopied] = React.useState(false)

  React.useEffect(() => {
    if (!copied) return
    const timeout = setTimeout(() => setCopied(false), 2000)
    return () => clearTimeout(timeout)
  }, [copied])

  const copyUsername = () => {
    if (!displayName) return
    navigator.clipboard
      .writeText(displayName)
      .then(() => setCopied(true))
      .catch((error) => errorToast(t('copyFailed'), error))
  }

  return (
    <Card
      className={cn('h-fit w-full sm:max-w-sm sm:min-w-sm', className)}
      {...props}
    >
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          {profile?.mojangUuid &&
            !(nameCosmetics ?? profile.cosmetics.name).glythPrefix && (
              <PlayerFace player={profile} className="size-5" />
            )}
          <NamePrefixes cosmetics={nameCosmetics ?? profile?.cosmetics.name} />
          <McUsername
            username={displayName ?? 'Steve'}
            className="text-xl"
            colors={
              nameCosmetics?.colors?.colors ??
              profile?.cosmetics.name.colors?.colors
            }
          />
          {profile?.verified && <VerifiedBadge />}
          {displayName && (
            <Button
              variant="ghost"
              size="icon"
              className="size-6"
              onClick={copyUsername}
              aria-label={t('copyUsername')}
              title={t('copyUsername')}
            >
              {copied ? (
                <CheckIcon className="size-4" />
              ) : (
                <CopyIcon className="size-4" />
              )}
            </Button>
          )}
        </CardTitle>
        <CardDescription
          className="font-medium"
          style={{ color: PROFILE_ROLE_COLORS[profile?.role ?? 'player'] }}
        >
          {t(`roles.${profile?.role ?? 'player'}`)}
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="flex flex-col gap-2">
          <canvas
            id="skin_container"
            className="bg-card h-[200px] max-w-[300px] self-center"
          />
          <Separator className="mb-4" />
          <div className="mb-0 flex flex-nowrap items-center gap-2">
            <CircleUserRoundIcon className="size-4" />
            <p>
              <span className="text-muted-foreground">{t('status')}:</span>
              &nbsp;
              <span
                className="font-medium"
                style={{ color: PROFILE_STATUS_COLORS[status] }}
              >
                {t(`statuses.${status}`)}
              </span>
            </p>
          </div>
          <div className="mb-4 flex flex-nowrap items-center gap-2">
            <ClockIcon className="size-4" />
            <p>
              <span className="text-muted-foreground">
                {t('playtime.title')}:
              </span>
              &nbsp;
              <span className="font-medium">{playtime.value}</span>
              &nbsp;
              {t(`playtime.${playtime.unit}`)}
            </p>
          </div>

          <p>
            <span className="text-muted-foreground">
              {t('lastSeenAt.title')}:
            </span>
            &nbsp;
            {profile?.isOnline ? (
              <span>{t('lastSeenAt.now')}</span>
            ) : profile?.lastSeenAt ? (
              <span suppressHydrationWarning>
                <TimeAgo date={profile.lastSeenAt} locale={locale} />
              </span>
            ) : (
              <>&mdash;</>
            )}
          </p>
          <p className="mb-0">
            <span className="text-muted-foreground">
              {t('firstSeenAt.title')}:
            </span>
            &nbsp;
            {profile?.isOnline && !profile.firstSeenAt ? (
              <span>{t('firstSeenAt.now')}</span>
            ) : profile?.firstSeenAt ? (
              <span suppressHydrationWarning>
                <TimeAgo date={profile.firstSeenAt} locale={locale} />
              </span>
            ) : (
              <>&mdash;</>
            )}
          </p>
        </div>
      </CardContent>
    </Card>
  )
}
