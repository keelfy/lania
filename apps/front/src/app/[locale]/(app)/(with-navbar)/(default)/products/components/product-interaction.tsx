'use client'

import type { CSSProperties, ComponentProps, PointerEvent } from 'react'
import { cn } from '@/lib/utils'
import styles from './product-interaction.module.css'

type Props = ComponentProps<'div'> & { accent?: string }

export default function ProductInteraction({
  accent = '#80cfc3',
  className,
  style,
  ...props
}: Props) {
  function onPointerMove(event: PointerEvent<HTMLDivElement>) {
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
    <div
      {...props}
      className={cn(styles.card, className)}
      style={{ ...style, '--product-accent': accent } as CSSProperties}
      onPointerMove={onPointerMove}
    />
  )
}
