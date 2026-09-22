'use client'

import NotificationItem from '@/components/notification-item'
import { Button } from '@/components/ui/button'
import { getNotifications, markNotificationsRead } from '@/lib/api-endpoints'
import { clientApiFetcher } from '@/lib/client'
import { errorToast } from '@/lib/toasts'
import {
  announceNotificationsRead,
  onNotificationsRead,
} from '@/lib/use-notifications'
import { Notification } from '@/models/notification'
import { useTranslations } from 'next-intl'
import React from 'react'

const PAGE_SIZE = 20

type Props = {
  title: string
}

// NotificationFeed lists every notification, read or not, newest first, a page at a time.
export default function NotificationFeed({ title }: Props) {
  const t = useTranslations('notifications')
  const [notifications, setNotifications] = React.useState<Notification[]>([])
  const [unreadCount, setUnreadCount] = React.useState(0)
  const [hasMore, setHasMore] = React.useState(false)
  const [isLoading, setIsLoading] = React.useState(true)
  const [source] = React.useState(() => Symbol('feed'))

  // reload asks again for everything already shown, so a read elsewhere shows up here too.
  const reload = React.useCallback(
    async (limit: number) => {
      setIsLoading(true)
      try {
        const list = await getNotifications(clientApiFetcher, { limit })
        setNotifications(list.content)
        setUnreadCount(list.unreadCount)
        setHasMore(list.hasMore)
      } catch (error) {
        errorToast(t('loadFailed'), error)
      } finally {
        setIsLoading(false)
      }
    },
    [t],
  )

  const shown = notifications.length
  React.useEffect(() => {
    const first = setTimeout(() => void reload(PAGE_SIZE), 0)
    return () => clearTimeout(first)
  }, [reload])

  React.useEffect(
    () =>
      onNotificationsRead(
        source,
        () => void reload(Math.max(shown, PAGE_SIZE)),
      ),
    [reload, shown, source],
  )

  const loadMore = async () => {
    setIsLoading(true)
    try {
      const list = await getNotifications(clientApiFetcher, {
        offset: shown,
        limit: PAGE_SIZE,
      })
      // A notification that came in meanwhile shifts the offset, so a repeat is skipped.
      setNotifications((current) => {
        const known = new Set(current.map((n) => n.id))
        return [...current, ...list.content.filter((n) => !known.has(n.id))]
      })
      setUnreadCount(list.unreadCount)
      setHasMore(list.hasMore)
    } catch (error) {
      errorToast(t('loadFailed'), error)
    } finally {
      setIsLoading(false)
    }
  }

  // markRead stamps the notifications here right away and then asks the API. Empty ids reads all.
  const markRead = async (ids: string[] = []) => {
    const readAt = new Date().toISOString()
    setNotifications((current) =>
      current.map((n) =>
        !n.readAt && (ids.length === 0 || ids.includes(n.id))
          ? { ...n, readAt }
          : n,
      ),
    )
    setUnreadCount((count) =>
      ids.length === 0 ? 0 : Math.max(0, count - ids.length),
    )
    try {
      const list = await markNotificationsRead(clientApiFetcher, ids, {
        limit: 1,
      })
      setUnreadCount(list.unreadCount)
      announceNotificationsRead(source)
    } catch (error) {
      errorToast(t('markFailed'), error)
      void reload(Math.max(shown, PAGE_SIZE))
    }
  }

  return (
    <>
      <div className="flex flex-wrap items-end justify-between gap-2">
        <h1 className="text-4xl font-extrabold tracking-tight">
          {title}
          {unreadCount > 0 && (
            <span className="text-muted-foreground">&nbsp;({unreadCount})</span>
          )}
        </h1>
        {unreadCount > 0 && (
          <Button variant="outline" size="sm" onClick={() => void markRead()}>
            {t('markAllRead')}
          </Button>
        )}
      </div>

      {notifications.length === 0 ? (
        <p className="text-muted-foreground rounded-lg border p-10 text-center">
          {isLoading ? t('loading') : t('empty')}
        </p>
      ) : (
        <ul className="divide-y overflow-hidden rounded-lg border">
          {notifications.map((notification) => (
            <NotificationItem
              key={notification.id}
              notification={notification}
              markUnread
              onOpen={(n) => {
                if (!n.readAt) void markRead([n.id])
              }}
            />
          ))}
        </ul>
      )}

      {hasMore && (
        <Button
          variant="outline"
          className="self-center"
          disabled={isLoading}
          onClick={() => void loadMore()}
        >
          {t('loadMore')}
        </Button>
      )}
    </>
  )
}
