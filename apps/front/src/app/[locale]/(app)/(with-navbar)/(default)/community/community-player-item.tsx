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
import Link from 'next/link'

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
  const playtime = formatPlaytime(profile.playtime)
  const lastSeen = profile.lastSeenAt
    ? formatTimeAgo(profile.lastSeenAt)
    : undefined
  const hasBeenOnline = profile.isOnline || lastSeen !== undefined
  // A banned player is kicked from the server, so the ban replaces the online status.
  const banned = profile.isBanned
  return (
    <Link
      href={communityProfileHref(locale, profile.username, season)}
      className="hover:bg-accent/50 hover:border-foreground/20 flex items-center gap-4 rounded-lg border p-4 transition-colors"
    >
      <div className="relative shrink-0">
        <PlayerFace
          player={profile}
          className={cn('size-12 rounded-sm', banned && 'opacity-60 grayscale')}
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
        <div className="flex min-w-0 items-center gap-1.5">
          <NamePrefixes cosmetics={profile.cosmetics.name} />
          <McUsername
            username={profile.username}
            colors={profile.cosmetics.name.colors?.colors}
            className="truncate text-xl"
          />
        </div>
        <div className="text-muted-foreground flex items-center gap-2 text-sm">
          <span className="flex shrink-0 items-center gap-1">
            <ClockIcon className="size-3.5" />
            {playtime.value} {t(`playtime.${playtime.unit}`)}
          </span>
          {banned ? (
            <>
              <span aria-hidden className="shrink-0">
                ·
              </span>
              <span
                className="truncate font-medium"
                style={{ color: PROFILE_STATUS_COLORS.banned }}
              >
                {t('statuses.banned')}
              </span>
            </>
          ) : (
            hasBeenOnline && (
              <>
                <span aria-hidden className="shrink-0">
                  ·
                </span>
                {profile.isOnline ? (
                  <span
                    className="truncate"
                    style={{ color: PROFILE_STATUS_COLORS.online }}
                  >
                    {t('statuses.online')}
                  </span>
                ) : (
                  <span className="truncate" suppressHydrationWarning>
                    {lastSeen!.unit === 'now'
                      ? t('timeAgo.now')
                      : t('timeAgo.ago', {
                          value: lastSeen!.value,
                          unit: t(`timeAgo.${lastSeen!.unit}`),
                        })}
                  </span>
                )}
              </>
            )
          )}
          {profile.role !== 'player' && (
            <span
              className="ml-auto shrink-0 rounded-sm border px-1.5 text-xs font-medium"
              style={{
                color: PROFILE_ROLE_COLORS[profile.role],
                borderColor: PROFILE_ROLE_COLORS[profile.role],
              }}
            >
              {t(`roles.${profile.role}`)}
            </span>
          )}
        </div>
      </div>
    </Link>
  )
}
