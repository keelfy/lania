'use client'

import McUsername from '@/components/ui/mc-username'
import NamePrefixes from '@/components/ui/name-prefixes'
import PlayerFace from '@/components/ui/player-face'
import { formatPlaytime } from '@/lib/playtime'
import {
  PROFILE_ROLE_COLORS,
  PROFILE_STATUS_COLORS,
} from '@/lib/profile-colors'
import { PublicProfile } from '@/models/profile'
import { ClockIcon } from 'lucide-react'
import { useTranslations } from 'next-intl'
import { useTimeAgo } from 'next-timeago'
import Link from 'next/link'

type Props = {
  profile: PublicProfile
  locale: string
}

export default function CommunityPlayerItem({ profile, locale }: Props) {
  const t = useTranslations('playerCard')
  const { TimeAgo } = useTimeAgo()
  const playtime = formatPlaytime(profile.playtime)
  return (
    <Link
      href={`/${locale}/community/${encodeURIComponent(profile.username)}`}
      className="hover:bg-accent/50 hover:border-foreground/20 flex items-center gap-4 rounded-lg border p-4 transition-colors"
    >
      <div className="relative shrink-0">
        <PlayerFace player={profile} className="size-12 rounded-sm" />
        {profile.isOnline && (
          <span
            className="ring-background absolute -right-1 -bottom-1 size-3.5 rounded-full ring-3"
            style={{ backgroundColor: PROFILE_STATUS_COLORS.online }}
          />
        )}
      </div>
      <div className="flex min-w-0 flex-1 flex-col gap-1.5">
        <div className="flex min-w-0 items-center gap-1.5">
          <NamePrefixes cosmetics={profile.cosmetics.name} />
          <McUsername
            username={profile.username}
            colors={profile.cosmetics.name.colors.colors}
            className="truncate text-xl"
          />
        </div>
        <div className="text-muted-foreground flex items-center gap-2 text-sm">
          <span className="flex shrink-0 items-center gap-1">
            <ClockIcon className="size-3.5" />
            {playtime.value} {t(`playtime.${playtime.unit}`)}
          </span>
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
            profile.lastSeenAt && (
              <span className="truncate" suppressHydrationWarning>
                <TimeAgo date={profile.lastSeenAt} locale={locale} />
              </span>
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
