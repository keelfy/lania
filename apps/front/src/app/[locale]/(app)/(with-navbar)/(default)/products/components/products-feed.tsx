'use client'

import { Button } from '@/components/ui/button'
import { useCursorFeed } from '@/lib/use-cursor-feed'
import { LoaderCircleIcon } from 'lucide-react'
import { useTranslations } from 'next-intl'
import type { ReactNode } from 'react'
import { loadMoreProducts } from '../load-more-products'

// The first page comes from the server render, the next ones are fetched by cursor when the end of the grid scrolls into view.
export default function ProductsFeed({
  category,
  initialCards,
  initialCursor,
}: {
  category: string
  initialCards: ReactNode
  initialCursor?: string
}) {
  const t = useTranslations('products')
  const { pages, cursor, failed, pending, sentinel, loadMore } = useCursorFeed(
    initialCursor,
    async (requested) => {
      const page = await loadMoreProducts(category, requested)
      return { items: page.cards, nextCursor: page.nextCursor }
    },
  )

  return (
    <>
      <div className="grid w-full grid-cols-1 items-stretch gap-4 md:grid-cols-2 lg:grid-cols-3">
        {initialCards}
        {pages.map((page) => (
          <div key={page.key} className="contents">
            {page.items}
          </div>
        ))}
      </div>
      {cursor && (
        <div ref={sentinel} className="flex flex-col items-center gap-2">
          {failed && (
            <p role="alert" className="text-destructive text-sm">
              {t('loadMoreFailed')}
            </p>
          )}
          <Button variant="outline" onClick={loadMore} disabled={pending}>
            {pending && <LoaderCircleIcon className="animate-spin" />}
            {t('loadMore')}
          </Button>
        </div>
      )}
    </>
  )
}
