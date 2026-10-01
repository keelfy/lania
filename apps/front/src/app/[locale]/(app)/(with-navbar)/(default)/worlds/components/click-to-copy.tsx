'use client'

import { cn } from '@/lib/utils'
import React from 'react'

// Whether the address was just copied. Exposed via context so leaf icons
// (e.g. CopyStateIcon) can react without every server component in
// between needing to be a client component too.
export const CopiedAddressContext = React.createContext(false)

type Props = {
  copyText: string
  copyLabel: string
  copiedLabel: string
  errorLabel: string
  className?: string
  children: React.ReactNode
}

// Keeps clipboard feedback local to the connection button.
export default function ClickToCopy({
  copyText,
  copyLabel,
  copiedLabel,
  errorLabel,
  className,
  children,
}: Props) {
  const [copied, setCopied] = React.useState(false)
  const [failed, setFailed] = React.useState(false)
  const resetTimeout = React.useRef<ReturnType<typeof setTimeout> | undefined>(
    undefined,
  )

  const onCopy = React.useCallback(() => {
    if (!copyText || typeof navigator === 'undefined' || !navigator.clipboard) {
      setFailed(true)
      return
    }

    setFailed(false)
    navigator.clipboard
      .writeText(copyText)
      .then(() => {
        setCopied(true)
        clearTimeout(resetTimeout.current)
        resetTimeout.current = setTimeout(() => setCopied(false), 2000)
      })
      .catch((error) => {
        setFailed(true)
        console.error('Failed to copy server address', error)
      })
  }, [copyText])

  React.useEffect(() => () => clearTimeout(resetTimeout.current), [])

  return (
    <CopiedAddressContext.Provider value={copied}>
      <button
        type="button"
        onClick={onCopy}
        disabled={!copyText}
        aria-label={`${copyLabel}: ${copyText}`}
        className={cn(
          'group flex w-full flex-col items-start justify-between gap-2 rounded-lg border border-teal-500/20 bg-teal-500/5 px-4 py-3 text-start transition-colors hover:border-teal-400/50 hover:bg-teal-500/10 focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-teal-300 active:bg-teal-500/20 disabled:opacity-50 sm:flex-row sm:items-center sm:gap-4',
          className,
        )}
      >
        <span className="text-foreground min-w-0 font-mono text-sm break-all sm:text-base">
          {copyText}
        </span>
        <span className="flex shrink-0 items-center gap-2 text-sm text-teal-300">
          {children}
          <span className="grid" aria-live="polite" aria-atomic="true">
            <span
              className="invisible col-start-1 row-start-1"
              aria-hidden="true"
            >
              {copyLabel}
            </span>
            <span
              className="invisible col-start-1 row-start-1"
              aria-hidden="true"
            >
              {copiedLabel}
            </span>
            <span className="col-start-1 row-start-1">
              {copied ? copiedLabel : copyLabel}
            </span>
          </span>
        </span>
      </button>
      {failed && (
        <p role="alert" className="text-destructive mt-2 text-sm">
          {errorLabel}
        </p>
      )}
    </CopiedAddressContext.Provider>
  )
}
