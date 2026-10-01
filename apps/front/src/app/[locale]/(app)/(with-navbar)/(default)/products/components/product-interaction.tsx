'use client'

import type { CSSProperties, ComponentProps, PointerEvent } from 'react'
import { cn } from '@/lib/utils'
import { useImageAccent } from './use-image-accent'
import styles from './product-interaction.module.css'

type Props = ComponentProps<'div'> & {
  accent?: string
  // The accent is the average color of this image when no accent is given.
  accentImage?: { src: string; unoptimized?: boolean }
}

export default function ProductInteraction({
  accent,
  accentImage,
  className,
  style,
  ...props
}: Props) {
  const imageAccent = useImageAccent(
    accent ? undefined : accentImage?.src,
    accentImage?.unoptimized,
  )
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
      style={
        {
          ...style,
          '--product-accent': accent ?? imageAccent ?? '#80cfc3',
        } as CSSProperties
      }
      onPointerMove={onPointerMove}
    />
  )
}
