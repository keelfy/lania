'use client'

import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import { CheckIcon, LinkIcon } from 'lucide-react'
import { Season } from '@/models/season'
import { useTranslations } from 'next-intl'
import React from 'react'

type Props = React.ComponentProps<typeof Button> & {
  season?: Season
}

function serverAddress(season?: Season) {
  return season?.publicAddress
}

export default function CopyIPButton({ className, season, ...props }: Props) {
  const [copied, setCopied] = React.useState(false)
  const t = useTranslations('landing')
  const address = serverAddress(season)

  const copyLink = React.useCallback(() => {
    if (address && typeof navigator !== 'undefined' && navigator.clipboard) {
      navigator.clipboard.writeText(address)
      setCopied(true)
      setTimeout(() => {
        setCopied(false)
      }, 2000)
    }
  }, [address])

  return (
    <Button
      variant="outline"
      size="lg"
      className={cn('transition-colors', className)}
      {...props}
      onClick={copyLink}
      disabled={
        !season?.isActive || !address || season.preregistration
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
