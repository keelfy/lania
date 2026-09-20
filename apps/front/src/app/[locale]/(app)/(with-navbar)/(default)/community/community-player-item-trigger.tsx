'use client'

import McUsername from '@/components/ui/mc-username'
import NamePrefixes from '@/components/ui/name-prefixes'
import { PROFILE_ROLE_COLORS } from '@/lib/profile-colors'
import PlayerFace from '@/components/ui/player-face'
import { formatPlaytime } from '@/lib/playtime'
import { PublicProfile } from '@/models/profile'
import { ClockIcon } from 'lucide-react'
import { useTranslations } from 'next-intl'
import { useTimeAgo } from 'next-timeago'
import PublicProfileStatusText from './public-profile-status-text'

type Props = {
  profile: PublicProfile
  locale: string
}

const CommunityPlayerItemTrigger = ({ profile, locale }: Props) => {
  const t = useTranslations('playerCard')
  const { TimeAgo } = useTimeAgo()
  const playtime = formatPlaytime(profile.playtime)
  const status = profile.isOnline ? 'online' : 'offline'
  return (
    <>
      <div className="flex items-center gap-2">
        <PlayerFace player={profile} className="size-7" />
        <NamePrefixes cosmetics={profile.cosmetics.name} />
        <McUsername
          username={profile.username}
          colors={profile.cosmetics.name.colors.colors}
          className="text-2xl font-semibold"
        />
      </div>
      <div className="flex w-full items-center justify-between gap-2">
        <p
          className="text-muted-foreground text-md"
          style={{ color: PROFILE_ROLE_COLORS[profile.role] }}
        >
          {t(`roles.${profile.role}`)}
        </p>
        <PublicProfileStatusText status={status} />
      </div>
      <div className="text-muted-foreground flex w-full items-center justify-between gap-2 text-sm font-normal">
        <span className="flex items-center gap-1">
          <ClockIcon className="size-3.5" />
          {playtime.value} {t(`playtime.${playtime.unit}`)}
        </span>
        {!profile.isOnline && profile.lastSeenAt && (
          <span suppressHydrationWarning>
            <TimeAgo date={profile.lastSeenAt} locale={locale} />
          </span>
        )}
      </div>
    </>
  )
}

export default CommunityPlayerItemTrigger
