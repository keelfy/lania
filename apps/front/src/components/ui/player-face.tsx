import { crafatarFaceUrl, STEVE_FACE_URL } from '@/lib/skin'
import { cn } from '@/lib/utils'
import { Profile } from '@/models/profile'
import Image from 'next/image'
import React from 'react'

type PlayerFaceProps = Partial<React.ComponentProps<typeof Image>> & {
  player?: Pick<Profile, 'username' | 'mojangUuid' | 'skin'>
}

export default function PlayerFace({ player, ...props }: PlayerFaceProps) {
  const textureUrl = player?.skin?.textureUrl
  if (textureUrl) {
    return (
      <TextureFace
        textureUrl={textureUrl}
        alt={player.username}
        className={props.className}
      />
    )
  }

  return (
    <Image
      width={64}
      height={64}
      {...props}
      src={crafatarFaceUrl(player) ?? STEVE_FACE_URL}
      alt={player?.username ?? 'Steve'}
      unoptimized
    />
  )
}

// A texture is 8 faces wide, so it is drawn 8 times wider than the face and moved by one face for every 8 pixels.
// That holds for 64x64 and legacy 64x32 textures alike.
const FACE_LAYERS = [
  { left: '-100%', top: '-100%' },
  // The hat over the face.
  { left: '-500%', top: '-100%' },
]

// Cuts the face out of the skin texture, so a skin chosen in game needs no renderer.
function TextureFace({
  textureUrl,
  alt,
  className,
}: {
  textureUrl: string
  alt: string
  className?: string
}) {
  return (
    <span
      role="img"
      aria-label={alt}
      className={cn(
        'relative inline-block size-16 shrink-0 overflow-hidden align-middle',
        className,
      )}
    >
      {FACE_LAYERS.map((position) => (
        <Image
          key={position.left}
          src={textureUrl}
          width={512}
          height={512}
          alt=""
          aria-hidden
          unoptimized
          className="absolute h-auto w-[800%] max-w-none [image-rendering:pixelated]"
          style={position}
        />
      ))}
    </span>
  )
}
