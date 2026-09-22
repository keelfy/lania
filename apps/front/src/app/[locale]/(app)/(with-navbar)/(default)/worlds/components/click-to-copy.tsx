'use client'

import { Card } from '@/components/ui/card'
import { cn } from '@/lib/utils'
import React from 'react'

// Whether the address was just copied. Exposed via context so leaf icons
// (e.g. CopyStateIcon) can react without every server component in
// between needing to be a client component too.
export const CopiedAddressContext = React.createContext(false)

type Props = {
  copyText: string
  copyLabel: string
  className?: string
  children: React.ReactNode
}

// Makes the whole server Card clickable to copy the connection address
// (MOTD included), instead of relying on a small, easy-to-miss button.
// Renders the real `Card` primitive with interactive affordances layered
// on top, so it stays visually identical to every other card on the site.
export default function ClickToCopy({
  copyText,
  copyLabel,
  className,
  children,
}: Props) {
  const [copied, setCopied] = React.useState(false)
  const resetTimeout = React.useRef<ReturnType<typeof setTimeout> | undefined>(
    undefined,
  )

  const onCopy = React.useCallback(() => {
    if (!copyText || typeof navigator === 'undefined' || !navigator.clipboard) {
      return
    }

    navigator.clipboard
      .writeText(copyText)
      .then(() => {
        setCopied(true)
        clearTimeout(resetTimeout.current)
        resetTimeout.current = setTimeout(() => setCopied(false), 2000)
      })
      .catch((error) => {
        console.error('Failed to copy server address', error)
      })
  }, [copyText])

  const onKeyDown = React.useCallback(
    (event: React.KeyboardEvent) => {
      if (event.key !== 'Enter' && event.key !== ' ') return
      event.preventDefault()
      onCopy()
    },
    [onCopy],
  )

  React.useEffect(() => () => clearTimeout(resetTimeout.current), [])

  return (
    <CopiedAddressContext.Provider value={copied}>
      <Card
        role="button"
        tabIndex={0}
        onClick={onCopy}
        onKeyDown={onKeyDown}
        aria-label={`${copyLabel}: ${copyText}`}
        className={cn(
          'hover:bg-accent/40 focus-visible:ring-ring cursor-pointer text-start transition-colors focus-visible:ring-2 focus-visible:outline-none',
          className,
        )}
      >
        {children}
      </Card>
    </CopiedAddressContext.Provider>
  )
}
