'use client'

import { capeUrl, skinTexture, STEVE_SKIN_URL } from '@/lib/skin'
import { initializeViewer } from '@/lib/skin-viewer'
import { ProfileDetails } from '@/models/profile'
import { PauseIcon, PlayIcon, RotateCcwIcon } from 'lucide-react'
import { useTranslations } from 'next-intl'
import { useTheme } from 'next-themes'
import { useEffect, useRef, useState } from 'react'
import { Button } from './button'

type Props = {
  profile?: ProfileDetails
  colors?: string[]
  interactive: boolean
}

export default function PlayerSkin({ profile, colors, interactive }: Props) {
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const viewerRef = useRef<ReturnType<typeof initializeViewer>>(null)
  const [rotating, setRotating] = useState(false)
  const { resolvedTheme } = useTheme()
  const t = useTranslations('playerCard.skin')
  const skin = skinTexture(profile)
  const cape = capeUrl(profile)

  useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas || !profile?.id) return
    const viewer = initializeViewer(
      skin?.url ?? STEVE_SKIN_URL,
      cape,
      skin?.slim ?? false,
      resolvedTheme === 'dark' ? 'dark' : 'light',
      { canvas, transparent: interactive },
    )
    const reducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)')
    viewer.autoRotate = !interactive && !reducedMotion.matches
    const stopRotation = () => {
      if (reducedMotion.matches) {
        viewer.autoRotate = false
        setRotating(false)
      }
    }
    reducedMotion.addEventListener('change', stopRotation)
    viewerRef.current = viewer
    return () => {
      reducedMotion.removeEventListener('change', stopRotation)
      viewerRef.current = null
      viewer.dispose()
    }
  }, [profile?.id, skin?.url, skin?.slim, cape, resolvedTheme, interactive])

  // Keep the control state and viewer in sync after a skin or theme change.
  useEffect(() => {
    if (interactive && viewerRef.current) {
      viewerRef.current.autoRotate = rotating
    }
  }, [
    rotating,
    interactive,
    profile?.id,
    skin?.url,
    skin?.slim,
    cape,
    resolvedTheme,
  ])

  const first = colors?.[0] ?? 'var(--primary)'
  const last = colors?.[colors.length - 1] ?? first
  return (
    <div
      className={
        interactive
          ? 'relative overflow-hidden rounded-lg border'
          : 'self-center'
      }
      style={
        interactive
          ? {
              backgroundImage: `radial-gradient(ellipse at 50% 90%, color-mix(in srgb, ${first} 20%, transparent), transparent 65%), radial-gradient(ellipse at 50% 10%, color-mix(in srgb, ${last} 10%, transparent), transparent 70%)`,
            }
          : undefined
      }
    >
      <canvas
        ref={canvasRef}
        width={300}
        height={200}
        aria-label={t('label')}
        className="mx-auto block h-auto w-[300px] max-w-full cursor-grab active:cursor-grabbing"
      />
      {interactive && (
        <div className="flex items-center justify-between gap-2 px-3 pb-2">
          <p className="text-muted-foreground text-xs">{t('drag')}</p>
          <div className="flex shrink-0 gap-1">
            <Button
              variant="ghost"
              size="icon"
              className="size-8"
              aria-label={t(rotating ? 'pause' : 'rotate')}
              title={t(rotating ? 'pause' : 'rotate')}
              aria-pressed={rotating}
              disabled={!profile}
              onClick={() => setRotating((value) => !value)}
            >
              {rotating ? <PauseIcon /> : <PlayIcon />}
            </Button>
            <Button
              variant="ghost"
              size="icon"
              className="size-8"
              aria-label={t('reset')}
              title={t('reset')}
              disabled={!profile}
              onClick={() => {
                const viewer = viewerRef.current
                if (!viewer) return
                setRotating(false)
                viewer.playerObject.rotation.y = 0
                viewer.resetCameraPose()
              }}
            >
              <RotateCcwIcon />
            </Button>
          </div>
        </div>
      )}
    </div>
  )
}
