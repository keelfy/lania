export enum Currency {
  RUB = 'RUB',
  USD = 'USD',
  EUR = 'EUR',
  BRL = 'BRL',
  TRY = 'TRY',
  PLN = 'PLN',
}

export const DEFAULT_CURRENCY = Currency.RUB

export const SUPPORTED_CURRENCIES = [
  Currency.RUB,
  Currency.USD,
  Currency.EUR,
  Currency.BRL,
  Currency.TRY,
  Currency.PLN,
]

export const CURRENCY_SYMBOLS = {
  [Currency.RUB]: '₽',
  [Currency.USD]: '$',
  [Currency.EUR]: '€',
  [Currency.BRL]: 'R$',
  [Currency.TRY]: '₺',
  [Currency.PLN]: 'zł',
}

export const CURRENCY_COOKIE = 'currency'
