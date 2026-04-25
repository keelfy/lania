'use client'

import { cn } from '@/lib/utils'
import { ArrowDownIcon } from 'lucide-react'
import { useTranslations } from 'next-intl'
import React from 'react'

type SeeMoreLandingButtonProps = React.ComponentProps<'button'> & {
  sectionId: string
}

export default function SeeMoreLandingButton({
  children,
  sectionId,
  className,
  ...props
}: SeeMoreLandingButtonProps) {
  const t = useTranslations('landing')

  const scrollToNext = () => {
    const next = document.getElementById(sectionId)
    if (next) next.scrollIntoView({ behavior: 'smooth', block: 'start' })
    else {
      console.error(`Section with id ${sectionId} not found`)
    }
  }

  return (
    <button
      onClick={scrollToNext}
      type="button"
      className={cn(
        'group focus-visible:ring-primary absolute left-1/2 flex w-fit -translate-x-1/2 flex-col items-center gap-6 py-2 shadow-lg focus:outline-none focus-visible:ring-2',
        className,
      )}
      aria-label={t('seeMore')}
      {...props}
    >
      <p className="text-primary/80 group-hover:text-primary text-lg font-medium transition-colors duration-300">
        {children}
      </p>
      <ArrowDownIcon className="text-primary group-hover:text-primary size-8 animate-bounce rounded-full bg-teal-900/30 p-1 transition-colors duration-300 group-hover:bg-teal-900/50" />
    </button>
  )
}
