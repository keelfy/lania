'use client'

import { useEffect, useRef, useState } from 'react'
import { initializeViewer } from '@/lib/skin-viewer'
import type { SkinViewer } from 'skinview3d'

export default function CustomizationSkinPreview({
  src,
  slim,
  label,
  fallback,
}: {
  src: string
  slim: boolean
  label: string
  fallback: string
}) {
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const [failed, setFailed] = useState(false)

  useEffect(() => {
    let viewer: SkinViewer | undefined
    const frame = requestAnimationFrame(() => {
      if (!canvasRef.current) return
      try {
        viewer = initializeViewer(src, '', slim, 'dark', {
          canvas: canvasRef.current,
          width: 280,
          height: 360,
          transparent: true,
        })
        viewer.autoRotate = false
        viewer.playerObject.rotation.y = -0.35
      } catch {
        setFailed(true)
      }
    })
    return () => {
      cancelAnimationFrame(frame)
      viewer?.dispose()
    }
  }, [src, slim])

  return failed ? (
    <p className="flex h-[360px] items-center px-6 text-center text-sm text-white/70">
      {fallback}
    </p>
  ) : (
    <canvas
      ref={canvasRef}
      role="img"
      aria-label={label}
      className="max-w-full touch-pan-y"
      style={{ width: 280, height: 360 }}
    />
  )
}
