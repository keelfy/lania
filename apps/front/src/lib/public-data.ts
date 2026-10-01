import { getPrimarySeason } from '@/lib/api-endpoints'
import { publicApiFetcher } from '@/lib/public-api'
import { cacheLife } from 'next/cache'

// A failed read throws, so a failure is never cached. Callers decide what to show instead.
export async function getCachedPrimarySeason() {
  'use cache'
  cacheLife('minutes')
  return getPrimarySeason(publicApiFetcher)
}
