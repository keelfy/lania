'use client'

import { initializeViewer } from '@/lib/skin-viewer'
import { cn } from '@/lib/utils'
import React from 'react'

type Props = React.ComponentProps<'div'> & {
  mojangUuid?: string
  isSlimModel: boolean
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
  mojangUuid,
  isSlimModel,
  colors,
  className,
  ...props
}: Props) {
  const canvasRef = React.useRef<HTMLCanvasElement>(null)

  React.useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas) return

    const skinUrl = mojangUuid
      ? `https://crafatar-pub.neodium.fr/skins/${mojangUuid}`
      : '/images/steve_skin.png'
    const capeUrl = mojangUuid
      ? `https://crafatar-pub.neodium.fr/capes/${mojangUuid}`
      : ''
    const viewer = initializeViewer(skinUrl, capeUrl, isSlimModel, 'dark', {
      canvas,
      width: STAGE_WIDTH,
      height: STAGE_HEIGHT,
      transparent: true,
    })
    return () => viewer.dispose()
  }, [mojangUuid, isSlimModel])

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
