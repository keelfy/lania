import { cacheLife } from 'next/cache'

// The year is read inside a cache, so the footer can be part of a prerendered page. It refreshes daily.
export async function CurrentYear() {
  'use cache'
  cacheLife('days')
  return new Date().getFullYear()
}
