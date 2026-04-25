'use client'

import CurrencySelect from '@/components/ui/currency-select'
import { useBasket } from '@/context/basket'
import { Currency } from '@/lib/currency'
import { useTranslations } from 'next-intl'
import React from 'react'

type Props = React.ComponentProps<'h2'> & {
  currency: Currency
}

export default function BasketTitle({ currency, ...props }: Props) {
  const t = useTranslations()
  const { items } = useBasket()
  return (
    <div className="flex w-full items-center justify-between gap-2">
      <h2 {...props}>
        {t('basket.title')}&nbsp;
        <span className="text-muted-foreground">({items.length})</span>
      </h2>
      <CurrencySelect currency={currency} className="hidden lg:flex" />
    </div>
  )
}
