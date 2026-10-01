'use client'

import { Button } from '@/components/ui/button'
import { LoaderCircleIcon } from 'lucide-react'
import { useTranslations } from 'next-intl'
import {
  useEffect,
  useRef,
  useState,
  useTransition,
  type ReactNode,
} from 'react'
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
  const [pages, setPages] = useState<{ key: string; cards: ReactNode }[]>([])
  const [cursor, setCursor] = useState(initialCursor)
  const [failed, setFailed] = useState(false)
  const [pending, startTransition] = useTransition()
  const sentinel = useRef<HTMLDivElement>(null)

  const loadMore = () => {
    if (!cursor || pending) return
    const requested = cursor
    startTransition(async () => {
      try {
        const page = await loadMoreProducts(category, requested)
        setPages((current) => [
          ...current,
          { key: requested, cards: page.cards },
        ])
        setCursor(page.nextCursor)
        setFailed(false)
      } catch (error) {
        console.error(error)
        setFailed(true)
      }
    })
  }
  const latestLoadMore = useRef(loadMore)
  useEffect(() => {
    latestLoadMore.current = loadMore
  })

  // A new observer reports the current state right away, so a sentinel that is still in view after a page loaded triggers the next one.
  useEffect(() => {
    const node = sentinel.current
    if (!node || !cursor || pending || failed) return
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries.some((entry) => entry.isIntersecting))
          latestLoadMore.current()
      },
      { rootMargin: '600px 0px' },
    )
    observer.observe(node)
    return () => observer.disconnect()
  }, [cursor, pending, failed])

  return (
    <>
      <div className="grid w-full grid-cols-1 items-stretch gap-4 md:grid-cols-2 lg:grid-cols-3">
        {initialCards}
        {pages.map((page) => (
          <div key={page.key} className="contents">
            {page.cards}
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
