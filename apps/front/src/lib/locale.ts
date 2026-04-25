export enum Locale {
  Russian = 'ru',
  English = 'en',
}

export const LOCALES = [Locale.Russian, Locale.English]
export const DEFAULT_LOCALE = Locale.English

export const LOCALE_COOKIE = 'locale'

export const LOCALE_NAMES = {
  [Locale.Russian]: 'Русский',
  [Locale.English]: 'English',
}
