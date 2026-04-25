import { DEFAULT_LOCALE, LOCALE_COOKIE, LOCALES } from '@/lib/locale'
import { defineRouting } from 'next-intl/routing'

export const routing = defineRouting({
  locales: LOCALES,
  defaultLocale: DEFAULT_LOCALE,
  localeCookie: {
    name: LOCALE_COOKIE,
    maxAge: 60 * 60 * 24 * 365,
  },
})
