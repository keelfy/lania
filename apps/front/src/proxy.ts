import createMiddleware from 'next-intl/middleware'
import { NextRequest, NextResponse } from 'next/server'
import { routing } from './i18n/routing'
import { isCurrentSessionActive } from './lib/get-current-session'
import { DEFAULT_LOCALE, LOCALE_COOKIE, LOCALES } from './lib/locale'

const privateRoutes = ['profiles']

const handleI18nRouting = createMiddleware(routing)

export async function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl
  const pathnameHasLocale = LOCALES.some(
    (locale) => pathname.startsWith(`/${locale}/`) || pathname === `/${locale}`,
  )

  if (pathnameHasLocale) {
    // auth and wiki are public; root locale index is public
    const [, currentLocale, ...segments] = pathname.split('/')
    const first = segments[0] || ''

    // update locale in cookies if it's not the default locale
    if (
      currentLocale !== DEFAULT_LOCALE &&
      request.cookies.has(LOCALE_COOKIE) &&
      request.cookies.get(LOCALE_COOKIE)?.value !== currentLocale
    ) {
      request.cookies.set(LOCALE_COOKIE, currentLocale)
    }

    const isPrivateUnderLocale = privateRoutes.includes(first)
    if (isPrivateUnderLocale) {
      const ok = await isCurrentSessionActive()
      if (!ok) {
        const originalPath = `${pathname}${request.nextUrl.search}`
        request.nextUrl.pathname = `/${currentLocale}/auth/login`
        request.nextUrl.searchParams.set('goto', originalPath)
        return NextResponse.redirect(request.nextUrl)
      }
    }
  }

  return handleI18nRouting(request)
}

export const config = {
  matcher: ['/((?!api|static|.*\\..*|_next).*)'],
}
