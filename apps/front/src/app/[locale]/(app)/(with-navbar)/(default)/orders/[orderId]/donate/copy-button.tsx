'use client'

import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import { CheckIcon, LinkIcon } from 'lucide-react'
import React from 'react'

type Props = React.ComponentProps<typeof Button> & {
  copyText: string
}

export default function CopyButton({ className, copyText, ...props }: Props) {
  const [copied, setCopied] = React.useState(false)

  const onCopy = React.useCallback(() => {
    if (typeof navigator !== 'undefined' && navigator.clipboard) {
      navigator.clipboard.writeText(copyText)
      setCopied(true)
      setTimeout(() => {
        setCopied(false)
      }, 2000)
    }
  }, [])

  return (
    <Button
      variant="outline"
      className={cn('transition-colors', className)}
      {...props}
      onClick={onCopy}
    >
      <div className="relative h-4 w-4">
        <CheckIcon
          className={cn(
            'absolute transition-all duration-200',
            copied
              ? 'scale-100 text-teal-500 opacity-100'
              : 'scale-75 opacity-0',
          )}
        />
        <LinkIcon
          className={cn(
            'absolute transition-all duration-200',
            copied ? 'scale-75 opacity-0' : 'scale-100 opacity-100',
          )}
        />
      </div>
    </Button>
  )
}
