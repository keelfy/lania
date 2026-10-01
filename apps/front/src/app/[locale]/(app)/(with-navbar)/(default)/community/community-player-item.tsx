'use client'

import McUsername from '@/components/ui/mc-username'
import NamePrefixes from '@/components/ui/name-prefixes'
import PlayerFace from '@/components/ui/player-face'
import { formatPlaytime } from '@/lib/playtime'
import {
  PROFILE_ROLE_COLORS,
  PROFILE_STATUS_COLORS,
} from '@/lib/profile-colors'
import { formatTimeAgo } from '@/lib/time-ago'
import { PublicProfile } from '@/models/profile'
import { communityProfileHref } from './community-href'
import { cn } from '@/lib/utils'
import { BanIcon, ClockIcon } from 'lucide-react'
import { useTranslations } from 'next-intl'
import PlayerProfileLink from '@/components/ui/player-profile-link'

type Props = {
  profile: PublicProfile
  locale: string
  season?: string
}

export default function CommunityPlayerItem({
  profile,
  locale,
  season,
}: Props) {
  const t = useTranslations('playerCard')
  const tCommunity = useTranslations('community')
  const playtime = formatPlaytime(profile.playtime)
  const lastSeen = profile.lastSeenAt
    ? formatTimeAgo(profile.lastSeenAt)
    : undefined
  const hasBeenOnline = profile.isOnline || lastSeen !== undefined
  // A banned player is kicked from the server, so the ban replaces the online status.
  const banned = profile.isBanned
  return (
    <PlayerProfileLink
      href={communityProfileHref(locale, profile.username, season)}
      accent={profile.cosmetics.name.colors?.colors[0]}
      aria-label={profile.username}
      className="flex min-w-0 items-start gap-3 rounded-xl border p-4"
    >
      <div className="relative shrink-0">
        <PlayerFace
          player={profile}
          className={cn('size-12 rounded-lg', banned && 'opacity-60 grayscale')}
        />
        {banned ? (
          <span
            className="bg-background absolute -right-1.5 -bottom-1.5 rounded-full p-0.5"
            title={t('statuses.banned')}
          >
            <BanIcon
              className="size-4"
              strokeWidth={2.5}
              style={{ color: PROFILE_STATUS_COLORS.banned }}
            />
          </span>
        ) : (
          profile.isOnline && (
            <span
              className="ring-background absolute -right-1 -bottom-1 size-3.5 rounded-full ring-3"
              style={{ backgroundColor: PROFILE_STATUS_COLORS.online }}
            />
          )
        )}
      </div>
      <div className="flex min-w-0 flex-1 flex-col gap-1.5">
        <div
          className="flex min-w-0 items-center gap-1.5"
          title={profile.username}
        >
          <NamePrefixes cosmetics={profile.cosmetics.name} />
          <McUsername
            username={profile.username}
            colors={profile.cosmetics.name.colors?.colors}
            className="min-w-0 truncate text-lg"
          />
        </div>
        <div className="text-muted-foreground flex flex-wrap items-center gap-x-2 gap-y-1 text-sm">
          <span className="flex items-start gap-1.5">
            <ClockIcon aria-hidden className="mt-0.5 size-3.5 shrink-0" />
            {profile.playtime === 0
              ? tCommunity('notPlayed')
              : `${playtime.value.toLocaleString(locale)} ${t(`playtime.${playtime.unit}`)}`}
          </span>
        </div>
        {(banned || hasBeenOnline || profile.role !== 'player') && (
          <div className="text-muted-foreground flex flex-wrap items-center gap-x-2 gap-y-1 text-xs">
            {banned ? (
              <span
                className="font-medium"
                style={{ color: PROFILE_STATUS_COLORS.banned }}
              >
                {t('statuses.banned')}
              </span>
            ) : profile.isOnline ? (
              <span style={{ color: PROFILE_STATUS_COLORS.online }}>
                {t('statuses.online')}
              </span>
            ) : lastSeen ? (
              <span suppressHydrationWarning>
                {lastSeen.unit === 'now'
                  ? t('timeAgo.now')
                  : tCommunity('lastSeen', {
                      time: t('timeAgo.ago', {
                        value: lastSeen.value,
                        unit: t(`timeAgo.${lastSeen.unit}`),
                      }),
                    })}
              </span>
            ) : null}
            {profile.role !== 'player' && (
              <span
                className="shrink-0 rounded-sm border px-1.5 text-xs font-medium"
                style={{
                  color: PROFILE_ROLE_COLORS[profile.role],
                  borderColor: PROFILE_ROLE_COLORS[profile.role],
                }}
              >
                {t(`roles.${profile.role}`)}
              </span>
            )}
          </div>
        )}
      </div>
    </PlayerProfileLink>
  )
}
