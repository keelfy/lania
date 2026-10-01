'use client'

import { useEffect, useRef, useState } from 'react'
import { initializeViewer } from '@/lib/skin-viewer'
import { FunctionAnimation, type SkinViewer } from 'skinview3d'

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
    const motion = window.matchMedia('(prefers-reduced-motion: reduce)')
    const finePointer = window.matchMedia('(hover: hover) and (pointer: fine)')
    const pointer = { x: 0, y: 0 }
    const canvas = canvasRef.current
    const move = (event: PointerEvent) => {
      if (
        !canvas ||
        motion.matches ||
        !finePointer.matches ||
        event.buttons ||
        event.pointerType === 'touch'
      )
        return
      const bounds = canvas.getBoundingClientRect()
      pointer.x = ((event.clientX - bounds.left) / bounds.width - 0.5) * 0.7
      pointer.y = ((event.clientY - bounds.top) / bounds.height - 0.5) * 0.35
    }
    const reset = () => {
      pointer.x = 0
      pointer.y = 0
    }
    const syncMotion = () => {
      reset()
      if (!viewer) return
      if (motion.matches) {
        viewer.animation = null
        viewer.playerObject.skin.head.rotation.set(0, 0, 0)
        viewer.playerObject.skin.rightArm.rotation.set(0, 0, 0)
      } else {
        viewer.animation = new FunctionAnimation((player, progress, delta) => {
          const blend = 1 - Math.exp(-delta * 10)
          player.skin.head.rotation.y +=
            (pointer.x - player.skin.head.rotation.y) * blend
          player.skin.head.rotation.x +=
            (pointer.y - player.skin.head.rotation.x) * blend
          const wave = progress < 1.2 ? Math.sin((Math.PI * progress) / 1.2) : 0
          player.skin.rightArm.rotation.z =
            -wave * (1.3 + Math.sin(progress * 18) * 0.15)
        })
      }
    }
    const syncVisibility = () => {
      if (viewer) viewer.renderPaused = document.hidden
    }
    canvas?.addEventListener('pointermove', move, { passive: true })
    canvas?.addEventListener('pointerleave', reset)
    canvas?.addEventListener('pointerdown', reset)
    motion.addEventListener('change', syncMotion)
    document.addEventListener('visibilitychange', syncVisibility)
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
        syncMotion()
        syncVisibility()
      } catch {
        setFailed(true)
      }
    })
    return () => {
      cancelAnimationFrame(frame)
      canvas?.removeEventListener('pointermove', move)
      canvas?.removeEventListener('pointerleave', reset)
      canvas?.removeEventListener('pointerdown', reset)
      motion.removeEventListener('change', syncMotion)
      document.removeEventListener('visibilitychange', syncVisibility)
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
