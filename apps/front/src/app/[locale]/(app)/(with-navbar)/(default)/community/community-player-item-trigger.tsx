'use client'

import McUsername from '@/components/ui/mc-username'
import { PROFILE_ROLE_COLORS } from '@/components/ui/player-card'
import PlayerFace from '@/components/ui/player-face'
import { PublicProfile } from '@/models/profile'
import { useTranslations } from 'next-intl'
import PublicProfileStatusText from './public-profile-status-text'

type Props = {
  profile: PublicProfile
}

const CommunityPlayerItemTrigger = ({ profile }: Props) => {
  const t = useTranslations('playerCard')
  const status = profile.isOnline ? 'online' : 'offline'
  return (
    <>
      <div className="flex items-center gap-2">
        <PlayerFace player={profile} className="size-7" />
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
    </>
  )
}

export default CommunityPlayerItemTrigger
