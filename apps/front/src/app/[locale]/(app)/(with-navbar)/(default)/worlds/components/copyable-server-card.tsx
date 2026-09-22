'use client'

import { cn } from '@/lib/utils'
import React from 'react'

// Whether the address was just copied. Exposed via context (not a
// render-prop) because a Server Component can pass plain React nodes
// across the client boundary, but never a function.
export const CopiedAddressContext = React.createContext(false)

type Props = {
  copyText: string
  copyLabel: string
  className?: string
  children: React.ReactNode
}

// Wraps the whole server card so clicking anywhere on it (MOTD included)
// copies the connection address, instead of relying on a small,
// easy-to-miss copy button.
export default function CopyableServerCard({
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

  React.useEffect(() => () => clearTimeout(resetTimeout.current), [])

  return (
    <button
      type="button"
      onClick={onCopy}
      aria-label={`${copyLabel}: ${copyText}`}
      className={cn(
        'group bg-card hover:bg-accent/40 focus-visible:ring-ring flex w-full flex-col items-center justify-between gap-2 rounded-md px-4 py-3 text-start shadow-md transition-colors focus-visible:ring-2 focus-visible:outline-none sm:flex-row',
        className,
      )}
    >
      <CopiedAddressContext.Provider value={copied}>
        {children}
      </CopiedAddressContext.Provider>
    </button>
  )
}
