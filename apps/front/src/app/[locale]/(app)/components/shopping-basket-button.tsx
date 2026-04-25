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
  const isEmpty = React.useMemo(() => items.length === 0, [items.length])
  const t = useTranslations('basket')

  return (
    <Button
      variant="secondary"
      size={isEmpty ? 'icon' : 'default'}
      onClick={() => router.push('/basket')}
      {...props}
    >
      <span className="sr-only">{t('title')}</span>
      <div className="flex items-center gap-2">
        <ShoppingBasketIcon className="size-4" />
        {!isEmpty && <p className="text-sm">({items.length})</p>}
      </div>
    </Button>
  )
}
