'use client'

import { capeUrl, skinTexture, STEVE_SKIN_URL } from '@/lib/skin'
import { initializeViewer } from '@/lib/skin-viewer'
import { cn } from '@/lib/utils'
import { ProfileDetails } from '@/models/profile'
import React from 'react'

type Props = React.ComponentProps<'div'> & {
  profile: Pick<ProfileDetails, 'mojangUuid' | 'isSlimModel' | 'skin'>
  // The name colors of the player. They light the stage.
  colors: string[]
}

const STAGE_WIDTH = 300
const STAGE_HEIGHT = 400

// Any CSS color works here, so the alpha is mixed in instead of appended to a hex code.
function glow(color: string, percent: number) {
  return `color-mix(in srgb, ${color} ${percent}%, transparent)`
}

// The skin of the player on a stage lit in the colors of the player name.
export default function ProfileSkinStage({
  profile,
  colors,
  className,
  ...props
}: Props) {
  const canvasRef = React.useRef<HTMLCanvasElement>(null)
  const skin = skinTexture(profile)
  const cape = capeUrl(profile)

  React.useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas) return

    const viewer = initializeViewer(
      skin?.url ?? STEVE_SKIN_URL,
      cape,
      skin?.slim ?? false,
      'dark',
      {
        canvas,
        width: STAGE_WIDTH,
        height: STAGE_HEIGHT,
        transparent: true,
      },
    )
    return () => viewer.dispose()
  }, [skin?.url, skin?.slim, cape])

  const first = colors[0] ?? 'var(--primary)'
  const last = colors[colors.length - 1] ?? first

  return (
    <div
      className={cn(
        'bg-card relative flex items-end justify-center overflow-hidden rounded-xl border',
        className,
      )}
      style={{
        backgroundImage: `radial-gradient(70% 45% at 50% 82%, ${glow(first, 28)}, transparent), radial-gradient(60% 40% at 50% 18%, ${glow(last, 16)}, transparent)`,
      }}
      {...props}
    >
      <div
        aria-hidden
        className="bg-foreground/15 absolute bottom-6 h-4 w-40 rounded-[100%] blur-md"
      />
      <canvas
        ref={canvasRef}
        className="relative"
        style={{ width: STAGE_WIDTH, height: STAGE_HEIGHT }}
      />
    </div>
  )
}
