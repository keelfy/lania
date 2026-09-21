'use client'

import PlayerCard from '@/components/ui/player-card'
import { Profile } from '@/models/profile'
import { useSelectedProfile } from './use-selected-profile'

type Props = {
  profiles: Profile[]
  locale: string
  className?: string
}

export default function ProfilePlayerCard({
  profiles,
  locale,
  className,
}: Props) {
  const profile = useSelectedProfile(profiles)
  return (
    <PlayerCard
      profileId={profile?.id}
      username={profile?.username}
      nameCosmetics={profile?.cosmetics.name}
      locale={locale}
      className={className}
    />
  )
}
