import { hasLocale } from 'next-intl'
import { routing } from './routing'

// Next calls generateMetadata without params to render a 404 for a URL that matches no route,
// because the app has several root layouts under the [locale] segment.
export async function getMetadataLocale(
  params: Promise<{ locale?: string }> | undefined,
) {
  const locale = (await params)?.locale
  return hasLocale(routing.locales, locale) ? locale : routing.defaultLocale
}
