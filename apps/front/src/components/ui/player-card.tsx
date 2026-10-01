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
import { hasOwnFace, SKIN_CHANGED_EVENT } from '@/lib/skin'
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
import { useTimeAgo } from 'next-timeago'
import React from 'react'
import McUsername from './mc-username'
import NamePrefixes from './name-prefixes'
import PlayerFace from './player-face'
import VerifiedBadge from './verified-badge'
import PlayerSkin from './player-skin'

type PlayerCardProps = React.ComponentProps<typeof Card> & {
  profileId: string | undefined
  username?: string
  nameCosmetics?: NameCosmetics
  locale?: string
  interactiveSkin?: boolean
}

export default function PlayerCard({
  profileId,
  username,
  nameCosmetics,
  locale,
  className,
  interactiveSkin = false,
  ...props
}: PlayerCardProps) {
  const t = useTranslations('playerCard')
  const { TimeAgo } = useTimeAgo()

  const [loadedProfile, setProfile] = React.useState<ProfileDetails>()
  const [reloads, setReloads] = React.useState(0)

  const profile = loadedProfile?.id === profileId ? loadedProfile : undefined

  React.useEffect(() => {
    let active = true
    if (profileId) {
      getProfileDetails(clientApiFetcher, profileId)
        .then((result) => {
          if (active) setProfile(result)
        })
        .catch((error) => {
          if (!active) return
          errorToast(t('noProfile'), error)
          setProfile(undefined)
        })
    }
    return () => {
      active = false
    }
  }, [profileId, reloads, t])

  React.useEffect(() => {
    const reload = (event: Event) => {
      if ((event as CustomEvent<string>).detail === profileId) {
        setReloads((count) => count + 1)
      }
    }
    window.addEventListener(SKIN_CHANGED_EVENT, reload)
    return () => window.removeEventListener(SKIN_CHANGED_EVENT, reload)
  }, [profileId])

  const status = React.useMemo(
    () => (profile?.isOnline ? 'online' : 'offline'),
    [profile?.isOnline],
  )

  const playtime = React.useMemo(
    () => formatPlaytime(profile?.playtime ?? 0),
    [profile?.playtime],
  )

  const displayName = username ?? profile?.username
  const [copiedUsername, setCopiedUsername] = React.useState<string>()
  const copied = copiedUsername === displayName && !!displayName

  React.useEffect(() => {
    if (!copiedUsername) return
    const timeout = setTimeout(() => setCopiedUsername(undefined), 2000)
    return () => clearTimeout(timeout)
  }, [copiedUsername])

  const copyUsername = () => {
    if (!displayName) return
    navigator.clipboard
      .writeText(displayName)
      .then(() => setCopiedUsername(displayName))
      .catch((error) => errorToast(t('copyFailed'), error))
  }

  return (
    <Card
      className={cn('h-fit w-full sm:max-w-sm sm:min-w-sm', className)}
      {...props}
    >
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          {hasOwnFace(profile) &&
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
              aria-label={t(copied ? 'copiedUsername' : 'copyUsername')}
              title={t(copied ? 'copiedUsername' : 'copyUsername')}
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
          <PlayerSkin
            profile={profile}
            colors={
              nameCosmetics?.colors?.colors ??
              profile?.cosmetics.name.colors?.colors
            }
            interactive={interactiveSkin}
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
