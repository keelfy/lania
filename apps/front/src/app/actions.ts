'use server'

import { Currency, CURRENCY_COOKIE } from '@/lib/currency'
import { Locale, LOCALE_COOKIE, LOCALES } from '@/lib/locale'
import { cookies } from 'next/headers'

export async function setLanguage(lang: Locale, pathname: string = '/') {
  const pathnameHasLocale = LOCALES.some(
    (locale) => pathname.startsWith(`/${locale}/`) || pathname === `/${locale}`,
  )

  if (pathnameHasLocale) {
    const [, , ...segments] = pathname.split('/')
    pathname = `/${lang}/${segments.join('/')}`
  }

  const cookieStore = await cookies()
  if (cookieStore.get(LOCALE_COOKIE)?.value === lang) {
    return { redirectTo: pathname }
  }

  // if (cookieStore.get(COOKIE_CONSENT_COOKIE)?.value !== 'true') {
  //   return { redirectTo: pathname }
  // }

  cookieStore.set(LOCALE_COOKIE, lang)
  return { redirectTo: pathname }
}

export async function setCurrency(currency: Currency) {
  const cookieStore = await cookies()
  // if (cookieStore.get(COOKIE_CONSENT_COOKIE)?.value === 'true') {
  cookieStore.set(CURRENCY_COOKIE, currency)
  // }
}
