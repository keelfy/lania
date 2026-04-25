'use client'

import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import { CheckIcon, LinkIcon } from 'lucide-react'
import { useTranslations } from 'next-intl'
import React from 'react'

type Props = React.ComponentProps<typeof Button> & {
  locale: string
}

export default function CopyIPButton({ className, ...props }: Props) {
  const [copied, setCopied] = React.useState(false)
  const t = useTranslations('landing')

  const copyLink = React.useCallback(() => {
    if (typeof navigator !== 'undefined' && navigator.clipboard) {
      navigator.clipboard.writeText(`play.lania.gg`)
      setCopied(true)
      setTimeout(() => {
        setCopied(false)
      }, 2000)
    }
  }, [])

  return (
    <Button
      variant="outline"
      size="lg"
      className={cn('transition-colors', className)}
      {...props}
      onClick={copyLink}
      disabled={
        !process.env.NEXT_PUBLIC_ACTIVE_SEASON_ID ||
        process.env.NEXT_PUBLIC_ACTIVE_SEASON_ID.length === 0 ||
        process.env.NEXT_PUBLIC_PREREGISTRATION === 'true'
      }
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
      {t('copyIP')}
    </Button>
  )
}
