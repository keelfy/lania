'use client'

import { getNotifications, markNotificationsRead } from '@/lib/api-endpoints'
import { clientApiFetcher } from '@/lib/client'
import { Notification, NotificationQuery } from '@/models/notification'
import React from 'react'

// POLL_INTERVAL_MS is how often the bell menu asks the API for new notifications.
const POLL_INTERVAL_MS = 60_000

// The bell menu shows the unread notifications only, the rest live on the notifications page.
const UNREAD_QUERY: NotificationQuery = { unread: true, limit: 20 }

// NOTIFICATIONS_READ_EVENT tells the other notification lists on the page that something was read,
// so the bell counter and the notifications page do not wait for the next poll to agree.
const NOTIFICATIONS_READ_EVENT = 'lania:notifications-read'

// source names the list that read something, so it can skip its own announcement.
export function announceNotificationsRead(source: symbol) {
  window.dispatchEvent(
    new CustomEvent(NOTIFICATIONS_READ_EVENT, { detail: source }),
  )
}

// onNotificationsRead calls the listener when another list than source read something.
// It returns the unsubscribe function.
export function onNotificationsRead(source: symbol, listener: () => void) {
  const handle = (event: Event) => {
    if ((event as CustomEvent<symbol>).detail !== source) listener()
  }
  window.addEventListener(NOTIFICATIONS_READ_EVENT, handle)
  return () => window.removeEventListener(NOTIFICATIONS_READ_EVENT, handle)
}

// useNotifications keeps the unread notifications of the signed in user up to date.
// It polls while the tab is visible only, and refreshes as soon as the tab comes back.
export function useNotifications(enabled: boolean) {
  const [notifications, setNotifications] = React.useState<Notification[]>([])
  const [unreadCount, setUnreadCount] = React.useState(0)
  const [isLoading, setIsLoading] = React.useState(enabled)
  const [source] = React.useState(() => Symbol('bell'))

  const refresh = React.useCallback(async () => {
    if (!enabled) return
    try {
      const list = await getNotifications(clientApiFetcher, UNREAD_QUERY)
      setNotifications(list.content)
      setUnreadCount(list.unreadCount)
    } catch (error) {
      console.error(error)
    } finally {
      setIsLoading(false)
    }
  }, [enabled])

  // markRead drops the notifications from the menu right away and then asks the API.
  // Empty ids marks every unread notification.
  const markRead = React.useCallback(
    async (ids: string[] = []) => {
      if (!enabled) return
      setNotifications((current) =>
        ids.length === 0 ? [] : current.filter((n) => !ids.includes(n.id)),
      )
      setUnreadCount((count) =>
        ids.length === 0 ? 0 : Math.max(0, count - ids.length),
      )
      try {
        const list = await markNotificationsRead(
          clientApiFetcher,
          ids,
          UNREAD_QUERY,
        )
        setNotifications(list.content)
        setUnreadCount(list.unreadCount)
        announceNotificationsRead(source)
      } catch (error) {
        console.error(error)
        void refresh()
      }
    },
    [enabled, refresh, source],
  )

  React.useEffect(() => {
    // A visitor that is not signed in has nothing to poll for.
    if (!enabled) return

    const poll = () => {
      if (document.visibilityState === 'visible') void refresh()
    }

    // The first load waits for the render to finish, the rest is the interval and the tab coming back.
    const firstLoad = setTimeout(poll, 0)
    const interval = setInterval(poll, POLL_INTERVAL_MS)
    document.addEventListener('visibilitychange', poll)
    const unsubscribe = onNotificationsRead(source, poll)

    return () => {
      clearTimeout(firstLoad)
      clearInterval(interval)
      document.removeEventListener('visibilitychange', poll)
      unsubscribe()
    }
  }, [enabled, refresh, source])

  return { notifications, unreadCount, isLoading, refresh, markRead }
}
