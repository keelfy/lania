'use client'

import { useEffect, useRef } from 'react'
import { cn } from '@/lib/utils'
import styles from './landing-motion.module.css'

type LiquidGlassCardProps = React.PropsWithChildren<React.ComponentProps<'div'>>

export default function LiquidGlassCard({
  children,
  className,
  ...props
}: LiquidGlassCardProps) {
  const cardRef = useRef<HTMLDivElement>(null)
  useEffect(() => {
    const card = cardRef.current
    if (!card) return
    const motion = window.matchMedia('(prefers-reduced-motion: reduce)')
    const pointer = window.matchMedia('(hover: hover) and (pointer: fine)')
    let frame = 0
    const reset = () => {
      cancelAnimationFrame(frame)
      card.style.removeProperty('--tilt-x')
      card.style.removeProperty('--tilt-y')
      card.style.removeProperty('--glint-opacity')
    }
    const move = (event: PointerEvent) => {
      if (motion.matches || !pointer.matches || event.pointerType === 'touch')
        return
      const bounds = card.getBoundingClientRect()
      const x = Math.max(
        0,
        Math.min(1, (event.clientX - bounds.left) / bounds.width),
      )
      const y = Math.max(
        0,
        Math.min(1, (event.clientY - bounds.top) / bounds.height),
      )
      cancelAnimationFrame(frame)
      frame = requestAnimationFrame(() => {
        card.style.setProperty('--glint-x', `${x * 100}%`)
        card.style.setProperty('--glint-y', `${y * 100}%`)
        card.style.setProperty('--tilt-x', `${(0.5 - y) * 4}deg`)
        card.style.setProperty('--tilt-y', `${(x - 0.5) * 4}deg`)
        card.style.setProperty('--glint-opacity', '1')
      })
    }
    card.addEventListener('pointermove', move, { passive: true })
    card.addEventListener('pointerleave', reset)
    motion.addEventListener('change', reset)
    pointer.addEventListener('change', reset)
    window.addEventListener('blur', reset)
    return () => {
      reset()
      card.removeEventListener('pointermove', move)
      card.removeEventListener('pointerleave', reset)
      motion.removeEventListener('change', reset)
      pointer.removeEventListener('change', reset)
      window.removeEventListener('blur', reset)
    }
  }, [])
  return (
    <div
      ref={cardRef}
      className={cn(
        'px-6 py-4 backdrop-blur-xs sm:max-w-xl sm:items-start sm:rounded-2xl',
        styles.glass,
        className,
      )}
      {...props}
    >
      {children}
    </div>
  )
}
