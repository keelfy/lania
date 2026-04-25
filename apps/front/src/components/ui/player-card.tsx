'use client'

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
import { getProfileDetails } from '@/lib/api-endpoints'
import { clientApiFetcher } from '@/lib/client'
import { initializeViewer } from '@/lib/skin-viewer'
import { errorToast } from '@/lib/toasts'
import { cn } from '@/lib/utils'
import { NameCosmetics, ProfileDetails } from '@/models/profile'
import { CircleUserRoundIcon, ClockIcon } from 'lucide-react'
import { useTranslations } from 'next-intl'
import { useTheme } from 'next-themes'
import { useTimeAgo } from 'next-timeago'
import React from 'react'
import McUsername from './mc-username'
import PlayerFace from './player-face'

type PlayerCardProps = React.ComponentProps<typeof Card> & {
  profileId: string | undefined
  username?: string
  nameCosmetics?: NameCosmetics
  locale?: string
}

export const PROFILE_STATUS_COLORS = {
  online: '#22c55e', // green-500
  offline: '#eab308', // yellow-500
  banned: '#ef4444', // red-500
}

export const PROFILE_ROLE_COLORS = {
  admin: '#e7000b', // destructive
  player: '#71717b', // muted-foreground
  mod: '#10b981', // emerald-500
  owner: '#ef4444', // red-500
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
      ? `https://crafatar.com/skins/${profile.mojangUuid}`
      : '/images/steve_skin.png'
    const capeUrl = profile?.mojangUuid
      ? `https://crafatar.com/capes/${profile.mojangUuid}`
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

  const [ptSeconds, ptMinutes, ptHours] = React.useMemo(() => {
    const ptSeconds = (profile?.playtime ?? 0) / 1000
    const ptMinutes = ptSeconds / 60
    const ptHours = ptMinutes / 60
    return [ptSeconds, ptMinutes, ptHours]
  }, [profile?.playtime])

  return (
    <Card
      className={cn('h-fit w-full sm:max-w-sm sm:min-w-sm', className)}
      {...props}
    >
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          {profile?.mojangUuid && (
            <PlayerFace player={profile} className="size-5" />
          )}
          <McUsername
            username={username ?? profile?.username ?? 'Steve'}
            className="text-xl"
            colors={
              nameCosmetics?.colors.colors ??
              profile?.cosmetics.name.colors.colors
            }
          />
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
              <span className="font-medium">
                {ptSeconds > 60
                  ? ptMinutes > 60
                    ? Math.floor(ptHours)
                    : Math.floor(ptMinutes)
                  : Math.floor(ptSeconds)}
              </span>
              &nbsp;
              {ptSeconds > 60
                ? ptMinutes > 60
                  ? t('playtime.hours')
                  : t('playtime.minutes')
                : t('playtime.seconds')}
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
