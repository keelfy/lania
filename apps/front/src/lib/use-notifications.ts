'use client'

import { getNotifications, markNotificationsRead } from '@/lib/api-endpoints'
import { clientApiFetcher } from '@/lib/client'
import { Notification } from '@/models/notification'
import React from 'react'

// POLL_INTERVAL_MS is how often the bell menu asks the API for new notifications.
const POLL_INTERVAL_MS = 60_000

// useNotifications keeps the notifications of the signed in user up to date.
// It polls while the tab is visible only, and refreshes as soon as the tab comes back.
export function useNotifications(enabled: boolean) {
  const [notifications, setNotifications] = React.useState<Notification[]>([])
  const [unreadCount, setUnreadCount] = React.useState(0)
  const [isLoading, setIsLoading] = React.useState(enabled)

  const refresh = React.useCallback(async () => {
    if (!enabled) return
    try {
      const list = await getNotifications(clientApiFetcher)
      setNotifications(list.content)
      setUnreadCount(list.unreadCount)
    } catch (error) {
      console.error(error)
    } finally {
      setIsLoading(false)
    }
  }, [enabled])

  const markAllRead = React.useCallback(async () => {
    if (!enabled) return
    try {
      const list = await markNotificationsRead(clientApiFetcher)
      setNotifications(list.content)
      setUnreadCount(list.unreadCount)
    } catch (error) {
      console.error(error)
    }
  }, [enabled])

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

    return () => {
      clearTimeout(firstLoad)
      clearInterval(interval)
      document.removeEventListener('visibilitychange', poll)
    }
  }, [enabled, refresh])

  return { notifications, unreadCount, isLoading, refresh, markAllRead }
}
