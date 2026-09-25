'use client'

import { cn } from '@/lib/utils'
import { BadgeCheckIcon } from 'lucide-react'
import { useTranslations } from 'next-intl'
import { Tooltip, TooltipContent, TooltipTrigger } from './tooltip'

type Props = {
  className?: string
}

// The checkmark of a profile whose owner proved the licensed account in game.
export default function VerifiedBadge({ className }: Props) {
  const t = useTranslations('playerCard')

  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <span
          className="inline-flex shrink-0 align-middle"
          aria-label={t('verified')}
        >
          <BadgeCheckIcon
            className={cn('size-5 fill-sky-500 text-white', className)}
          />
        </span>
      </TooltipTrigger>
      <TooltipContent>{t('verified')}</TooltipContent>
    </Tooltip>
  )
}
