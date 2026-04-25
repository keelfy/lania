import { LOCALES } from '@/lib/locale'
import { createNavigation } from 'next-intl/navigation'
import { routing } from './routing'

// Lightweight wrappers around Next.js' navigation
// APIs that consider the routing configuration
export const { Link, redirect, usePathname, useRouter, getPathname } =
  createNavigation(routing)

export const pathnameHasLocale = (pathname: string) => {
  return LOCALES.some(
    (locale) => pathname.startsWith(`/${locale}/`) || pathname === `/${locale}`,
  )
}

export const pathnameStartsWith = (
  pathname: string,
  expectedPathname: string,
) => {
  const hasLocale = pathnameHasLocale(pathname)
  if (hasLocale) {
    const [, , ...segments] = pathname.split('/')
    const cleanPathname = segments.join('/')
    return cleanPathname.startsWith(`/${expectedPathname}`)
  }
  return pathname.startsWith(expectedPathname)
}
