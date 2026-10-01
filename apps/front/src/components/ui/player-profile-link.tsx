'use client'

import Link from 'next/link'
import type { ComponentProps, CSSProperties, PointerEvent } from 'react'
import { cn } from '@/lib/utils'
import styles from './player-profile-link.module.css'

type Props = ComponentProps<typeof Link> & { accent?: string }

// The light follows the cursor without updating React state or moving the content.
export default function PlayerProfileLink({
  accent,
  className,
  style,
  ...props
}: Props) {
  function moveLight(event: PointerEvent<HTMLAnchorElement>) {
    if (
      event.pointerType !== 'mouse' ||
      window.matchMedia('(prefers-reduced-motion: reduce)').matches
    )
      return
    const node = event.currentTarget
    const bounds = node.getBoundingClientRect()
    node.style.setProperty('--light-x', `${event.clientX - bounds.left}px`)
    node.style.setProperty('--light-y', `${event.clientY - bounds.top}px`)
  }

  return (
    <Link
      {...props}
      className={cn(
        styles.card,
        'focus-visible:ring-ring focus-visible:ring-offset-background outline-none focus-visible:ring-2 focus-visible:ring-offset-2',
        className,
      )}
      style={
        {
          ...style,
          '--player-accent': accent ?? 'var(--primary)',
        } as CSSProperties
      }
      onPointerMove={moveLight}
    />
  )
}
