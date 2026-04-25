'use client'

import { setCurrency } from '@/app/actions'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Currency,
  CURRENCY_SYMBOLS,
  SUPPORTED_CURRENCIES,
} from '@/lib/currency'
import { useTranslations } from 'next-intl'
import { useRouter } from 'next/navigation'
import React from 'react'

type Props = React.ComponentProps<typeof SelectTrigger> & {
  currency: Currency
}

export default function CurrencySelect({ currency, ...props }: Props) {
  const router = useRouter()
  const t = useTranslations('currencies.names')
  const [isCurrencyChanging, startCurrencyChange] = React.useTransition()
  const onCurrencyChange = (value: Currency) => {
    startCurrencyChange(async () => {
      await setCurrency(value)
      router.refresh()
    })
  }
  return (
    <Select
      defaultValue={currency}
      disabled={isCurrencyChanging}
      onValueChange={onCurrencyChange}
    >
      <SelectTrigger disabled={isCurrencyChanging} {...props}>
        <SelectValue placeholder={t(currency)} />
      </SelectTrigger>
      <SelectContent>
        {SUPPORTED_CURRENCIES.map((currency) => (
          <SelectItem value={currency} key={currency}>
            <p>
              <span className="font-mono font-semibold">{currency}</span>
              &nbsp;-&nbsp;{CURRENCY_SYMBOLS[currency]}&nbsp;
              <span className="text-muted-foreground">({t(currency)})</span>
            </p>
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  )
}
