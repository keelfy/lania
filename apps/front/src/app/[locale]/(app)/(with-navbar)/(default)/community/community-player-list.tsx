'use client'

import { PublicProfile } from '@/models/profile'
import CommunityPlayerItem from './community-player-item'
import React from 'react'

type Props = React.ComponentProps<'div'> & {
  profiles: PublicProfile[]
  locale: string
}

export default function CommunityPlayerList({
  profiles,
  locale,
  ...props
}: Props) {
  return (
    <div
      className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3"
      {...props}
    >
      {profiles.map((profile) => (
        <CommunityPlayerItem
          key={profile.id}
          profile={profile}
          locale={locale}
        />
      ))}
    </div>
  )
}
