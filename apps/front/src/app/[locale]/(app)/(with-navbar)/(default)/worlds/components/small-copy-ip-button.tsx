'use client'

import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import { CheckIcon, CopyIcon, LinkIcon } from 'lucide-react'
import React from 'react'

type Props = React.ComponentProps<typeof Button> & {
  copyText: string
}

export default function SmallCopyIpButton({ className, copyText, ...props }: Props) {
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
      variant='link'
      className={cn('transition-colors p-0 items-center flex gap-1 h-fit group', className)}
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
        <CopyIcon
          width={12}
          height={12}
          className={cn(
            'absolute transition-all duration-200 size-3 opacity-0 group-hover:opacity-100',
            copied ? 'scale-75 opacity-0 group-hover:opacity-0' : 'scale-100 group-hover:opacity-100',
          )}
        />
      </div>
      <p className="text-muted-foreground mr-2 font-mono text-xs">
        {copyText}
      </p>
    </Button>
  )
}
