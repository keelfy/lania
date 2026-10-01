import {
  getProduct,
  getProductCatalog,
  getProducts,
  getPrimarySeason,
  getSeasons,
  getSeasonScreenshots,
  getTopPlaytimeProfiles,
} from '@/lib/api-endpoints'
import { publicApiFetcher, publicApiFetcherFor } from '@/lib/public-api'
import { cacheLife } from 'next/cache'

// Public data every visitor sees the same, read without cookies and cached on the server. A failed read throws,
// so a failure is never cached. Callers decide what to show instead.

export async function getCachedSeasons() {
  'use cache'
  cacheLife('minutes')
  return getSeasons(publicApiFetcher)
}

export async function getCachedPrimarySeason() {
  'use cache'
  cacheLife('minutes')
  return getPrimarySeason(publicApiFetcher)
}

export async function getCachedSeasonScreenshots(
  seasonId: string,
  locale: string,
) {
  'use cache'
  cacheLife('minutes')
  return getSeasonScreenshots(publicApiFetcherFor(locale), seasonId)
}

export async function getCachedTopPlaytimeProfiles(
  locale: string,
  seasonId?: string,
) {
  'use cache'
  cacheLife('minutes')
  return getTopPlaytimeProfiles(
    publicApiFetcherFor(locale),
    undefined,
    seasonId,
  )
}

export async function getCachedProductCatalog(
  category: string | undefined,
  cursor: string | undefined,
  locale: string,
  currency: string,
) {
  'use cache'
  cacheLife('minutes')
  return getProductCatalog(
    publicApiFetcherFor(locale),
    category,
    cursor,
    locale,
    currency,
  )
}

export async function getCachedProducts(
  category: string | undefined,
  locale: string,
  currency: string,
) {
  'use cache'
  cacheLife('minutes')
  return getProducts(publicApiFetcherFor(locale), category, locale, currency)
}

export async function getCachedProduct(
  id: string,
  locale: string,
  currency: string,
) {
  'use cache'
  cacheLife('minutes')
  return getProduct(publicApiFetcherFor(locale), id, locale, currency)
}
