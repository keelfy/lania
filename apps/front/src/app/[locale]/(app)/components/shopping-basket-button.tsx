'use client'

import { Button } from '@/components/ui/button'
import { useBasket } from '@/context/basket'
import { ShoppingBasketIcon } from 'lucide-react'
import { useTranslations } from 'next-intl'
import { useRouter } from 'next/navigation'
import React from 'react'

type Props = React.ComponentProps<typeof Button>

export default function ShoppingBasketButton({ ...props }: Props) {
  const router = useRouter()
  const { items } = useBasket()
  const isEmpty = items.length === 0
  const feedback = React.useRef<HTMLDivElement>(null)
  const previousCount = React.useRef(items.length)

  React.useEffect(() => {
    const increased = items.length > previousCount.current
    previousCount.current = items.length
    if (
      !increased ||
      window.matchMedia('(prefers-reduced-motion: reduce)').matches
    )
      return
    const animation = feedback.current?.animate(
      [
        { transform: 'scale(1)' },
        { transform: 'scale(1.16)' },
        { transform: 'scale(1)' },
      ],
      { duration: 280, easing: 'ease-out' },
    )
    return () => animation?.cancel()
  }, [items.length])
  const t = useTranslations('basket')

  return (
    <Button
      variant="secondary"
      size={isEmpty ? 'icon' : 'default'}
      onClick={() => router.push('/basket')}
      {...props}
    >
      <span className="sr-only">{t('title')}</span>
      <div ref={feedback} className="flex items-center gap-2">
        <ShoppingBasketIcon aria-hidden="true" className="size-4" />
        <span className="sr-only" aria-live="polite" aria-atomic="true">
          {t('itemsCount', { count: items.length })}
        </span>
        {!isEmpty && (
          <p aria-hidden="true" className="text-sm tabular-nums">
            ({items.length})
          </p>
        )}
      </div>
    </Button>
  )
}
