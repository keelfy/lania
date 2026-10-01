'use server'

import { Currency, CURRENCY_COOKIE, DEFAULT_CURRENCY } from '@/lib/currency'
import { getCachedProductCatalog } from '@/lib/public-data'
import { getLocale } from 'next-intl/server'
import { cookies } from 'next/headers'
import type { ReactNode } from 'react'
import ProductCards from './components/product-cards'

// Renders the next catalog page on the server, so the cards stay server components.
export async function loadMoreProducts(
  category: string,
  cursor: string,
): Promise<{ cards: ReactNode; nextCursor?: string }> {
  const currency =
    ((await cookies()).get(CURRENCY_COOKIE)?.value as Currency) ??
    DEFAULT_CURRENCY
  const page = await getCachedProductCatalog(
    category === 'all' ? undefined : category,
    cursor,
    await getLocale(),
    currency,
  )
  return {
    cards: (
      <ProductCards
        items={page.content}
        currency={currency}
        previewOnly={false}
      />
    ),
    nextCursor: page.nextCursor,
  }
}
