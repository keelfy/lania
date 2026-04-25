'use client'

import { PublicProfile } from '@/models/profile'
import CommunityPlayerItem from './community-player-item'
import CommunityPlayerItemTrigger from './community-player-item-trigger'
import React from 'react'
import { Button } from '@/components/ui/button'

type Props = React.ComponentProps<'div'> & {
  profiles: PublicProfile[]
}

export default function CommunityPlayerList({ profiles, ...props }: Props) {
  return (
    <div
      className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3"
      {...props}
    >
      {profiles.map((profile) => (
        <CommunityPlayerItem key={profile.id} profile={profile}>
          <Button
            className="flex h-fit flex-col items-start gap-1 border p-4"
            variant="outline"
          >
            <CommunityPlayerItemTrigger profile={profile} />
          </Button>
        </CommunityPlayerItem>
      ))}
    </div>
  )
}
