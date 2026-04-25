'use client'

import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Currency, CURRENCY_SYMBOLS, DEFAULT_CURRENCY } from '@/lib/currency'
import { OrderAmount } from '@/models/order'
import { useTranslations } from 'next-intl'
import React from 'react'
import CopyButton from './copy-button'

type Props = {
  defaultCurrency?: Currency
  availableCurrencyCodes: Currency[]
  amounts: OrderAmount[]
}

export default function DonateAmountForm({
  defaultCurrency,
  availableCurrencyCodes,
  amounts,
}: Props) {
  const t = useTranslations()
  const [currency, setCurrency] = React.useState(
    defaultCurrency ?? DEFAULT_CURRENCY,
  )
  const amount = React.useMemo(() => {
    return amounts.find((amount) => amount.currency === currency)?.amount
  }, [amounts, currency])
  return (
    <>
      <li className="mt-2">
        <div className="flex items-center">
          <p className="text-lg font-semibold">Валюта доната:</p>
          <Select
            defaultValue={currency}
            onValueChange={(value) => setCurrency(value as Currency)}
          >
            <SelectTrigger className="ml-2 h-8 py-0" size="sm">
              <SelectValue placeholder="Валюта" />
            </SelectTrigger>
            <SelectContent>
              {availableCurrencyCodes.map((currencyCode) => {
                return (
                  <SelectItem value={currencyCode} key={currencyCode}>
                    <p>
                      <span className="font-mono font-semibold">
                        {currencyCode}
                      </span>
                      &nbsp;-&nbsp;{CURRENCY_SYMBOLS[currencyCode]}&nbsp;
                      <span className="text-muted-foreground">
                        ({t(`currencies.names.${currencyCode}`)})
                      </span>
                    </p>
                  </SelectItem>
                )
              })}
            </SelectContent>
          </Select>
        </div>
      </li>
      <li className="mt-3">
        <div className="flex items-center">
          <p className="text-lg font-semibold">Сумма доната:</p>
          <p className="bg-muted ml-2 h-8 rounded-md rounded-r-none border px-2 py-1 font-mono font-semibold">
            {amount?.toString() ?? ''}
          </p>
          <CopyButton
            copyText={amount?.toString() ?? ''}
            className="h-8 rounded-l-none border-l-0 p-0"
            size="icon"
          />
        </div>
      </li>
    </>
  )
}
