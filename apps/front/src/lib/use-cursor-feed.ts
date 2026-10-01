'use client'

import { useEffect, useRef, useState, useTransition } from 'react'

export type FeedPage<T> = { items: T; nextCursor?: string }

// Pages of a cursor paginated list. The first one comes from the server render, the next ones are loaded
// when the sentinel scrolls into view, or by loadMore. Pages are kept apart, so they can be keyed by their cursor.
export function useCursorFeed<T>(
  initialCursor: string | undefined,
  load: (cursor: string) => Promise<FeedPage<T>>,
) {
  const [pages, setPages] = useState<{ key: string; items: T }[]>([])
  const [cursor, setCursor] = useState(initialCursor)
  const [failed, setFailed] = useState(false)
  const [pending, startTransition] = useTransition()
  const sentinel = useRef<HTMLDivElement>(null)

  const loadMore = () => {
    if (!cursor || pending) return
    const requested = cursor
    startTransition(async () => {
      try {
        const page = await load(requested)
        setPages((current) => [
          ...current,
          { key: requested, items: page.items },
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

  return { pages, cursor, failed, pending, sentinel, loadMore }
}
