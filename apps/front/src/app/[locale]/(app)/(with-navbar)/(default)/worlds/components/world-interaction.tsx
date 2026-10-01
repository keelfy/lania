'use client'

import type { PointerEvent, ReactNode } from 'react'
import { cn } from '@/lib/utils'
import styles from './world-interaction.module.css'

export default function WorldInteraction({
  children,
  className,
}: {
  children: ReactNode
  className?: string
}) {
  function onPointerMove(event: PointerEvent<HTMLDivElement>) {
    if (
      event.pointerType !== 'mouse' ||
      window.matchMedia('(prefers-reduced-motion: reduce)').matches
    )
      return
    const card = event.currentTarget
    const bounds = card.getBoundingClientRect()
    const x = (event.clientX - bounds.left) / bounds.width
    const y = (event.clientY - bounds.top) / bounds.height
    card.style.setProperty('--world-x', `${(x - 0.5) * 10}px`)
    card.style.setProperty('--world-y', `${(y - 0.5) * 8}px`)
    card.style.setProperty('--light-x', `${x * 100}%`)
    card.style.setProperty('--light-y', `${y * 100}%`)
  }

  function reset(event: PointerEvent<HTMLDivElement>) {
    event.currentTarget.style.removeProperty('--world-x')
    event.currentTarget.style.removeProperty('--world-y')
  }

  return (
    <div
      className={cn(
        'group bg-card relative overflow-hidden rounded-lg',
        styles.world,
        className,
      )}
      onPointerMove={onPointerMove}
      onPointerLeave={reset}
    >
      {children}
    </div>
  )
}
