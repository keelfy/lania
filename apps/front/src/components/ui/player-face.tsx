import { Profile } from '@/models/profile'
import Image from 'next/image'
import React from 'react'

type PlayerFaceProps = Partial<React.ComponentProps<typeof Image>> & {
  player?: Profile
}

export default function PlayerFace({ player, ...props }: PlayerFaceProps) {
  const src = player?.mojangUuid
    ? `https://crafatar.com/avatars/${player.mojangUuid}?size=64`
    : '/images/steve_face.jpg'

  return (
    <Image
      width={64}
      height={64}
      {...props}
      src={src}
      alt={player?.username ?? 'Steve'}
      unoptimized
    />
  )
}
