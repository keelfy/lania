'use client'

import { cn } from '@/lib/utils'
import React from 'react'

const FADE_SIZE = '3rem'

type Props = React.ComponentProps<'div'>

// Scrolls its content horizontally and fades the edge that has more content behind it.
export default function ScrollFade({
  className,
  style,
  children,
  ...props
}: Props) {
  const ref = React.useRef<HTMLDivElement>(null)
  const [fade, setFade] = React.useState({ start: false, end: false })

  React.useEffect(() => {
    const element = ref.current
    if (!element) return

    const update = () => {
      const start = element.scrollLeft > 1
      const end =
        element.scrollLeft + element.clientWidth < element.scrollWidth - 1
      setFade((current) =>
        current.start === start && current.end === end
          ? current
          : { start, end },
      )
    }

    update()
    element.addEventListener('scroll', update, { passive: true })
    const observer = new ResizeObserver(update)
    observer.observe(element)
    if (element.firstElementChild) observer.observe(element.firstElementChild)
    return () => {
      element.removeEventListener('scroll', update)
      observer.disconnect()
    }
  }, [])

  const startSize = fade.start ? FADE_SIZE : '0px'
  const endSize = fade.end ? FADE_SIZE : '0px'
  const mask = `linear-gradient(to right, transparent, black ${startSize}, black calc(100% - ${endSize}), transparent)`

  return (
    <div
      ref={ref}
      className={cn('overflow-x-auto', className)}
      style={{ maskImage: mask, WebkitMaskImage: mask, ...style }}
      {...props}
    >
      {children}
    </div>
  )
}
