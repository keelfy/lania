'use client'

import NotificationItem from '@/components/notification-item'
import { Button } from '@/components/ui/button'
import {
  Drawer,
  DrawerContent,
  DrawerHeader,
  DrawerTitle,
  DrawerTrigger,
} from '@/components/ui/drawer'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Link } from '@/i18n/navigation'
import { useMediaQuery } from '@/lib/use-media-query'
import { useNotifications } from '@/lib/use-notifications'
import { Notification } from '@/models/notification'
import { useTranslations } from 'next-intl'
import React from 'react'

type Props = {
  // sessionActive is false for a visitor that is not signed in: they have no notifications.
  sessionActive: boolean
}

// NotificationsMenu lists the unread notifications. A notification is read when it is clicked
// or with "Mark all as read", the read ones are on the notifications page.
export default function NotificationsMenu({
  children,
  sessionActive,
}: React.PropsWithChildren<Props>) {
  const isDesktop = useMediaQuery('(min-width: 1024px)')
  const t = useTranslations('navbar.notifications')
  const [open, setOpen] = React.useState(false)
  const { notifications, unreadCount, isLoading, refresh, markRead } =
    useNotifications(sessionActive)

  const handleOpenChange = (next: boolean) => {
    setOpen(next)
    // Opening the menu shows the newest state.
    if (next) void refresh()
  }

  const handleOpen = (notification: Notification) => {
    setOpen(false)
    void markRead([notification.id])
  }

  const markAllRead = unreadCount > 0 && (
    <Button
      variant="ghost"
      size="sm"
      className="text-muted-foreground h-7 px-2 text-xs"
      onClick={() => void markRead()}
    >
      {t('markAllRead')}
    </Button>
  )

  const body = (
    <NotificationList
      notifications={notifications}
      isLoading={isLoading}
      onOpen={handleOpen}
    />
  )

  const footer = (
    <Link
      href="/notifications"
      onClick={() => setOpen(false)}
      className="text-muted-foreground hover:text-foreground hover:bg-accent/60 focus-visible:ring-ring block border-t px-4 py-2.5 text-center text-sm outline-none focus-visible:ring-2 focus-visible:ring-inset"
    >
      {unreadCount > notifications.length
        ? t('allWithMore', { count: unreadCount - notifications.length })
        : t('all')}
    </Link>
  )

  if (isDesktop) {
    return (
      <Popover open={open} onOpenChange={handleOpenChange}>
        <div className="relative">
          <PopoverTrigger asChild>{children}</PopoverTrigger>
          <UnreadBadge count={unreadCount} />
        </div>
        <PopoverContent align="end" className="w-96 p-0">
          <div className="flex min-h-12 items-center justify-between gap-2 border-b py-2 pr-2 pl-4">
            <p className="text-sm font-medium">{t('title')}</p>
            {markAllRead}
          </div>
          <ScrollArea className="max-h-[28rem]">{body}</ScrollArea>
          {footer}
        </PopoverContent>
      </Popover>
    )
  }

  return (
    <Drawer open={open} onOpenChange={handleOpenChange}>
      <div className="relative">
        <DrawerTrigger asChild>{children}</DrawerTrigger>
        <UnreadBadge count={unreadCount} />
      </div>
      <DrawerContent>
        <DrawerHeader className="flex-row items-center justify-between">
          <DrawerTitle>{t('title')}</DrawerTitle>
          {markAllRead}
        </DrawerHeader>
        <ScrollArea className="max-h-[60vh]">{body}</ScrollArea>
        {footer}
      </DrawerContent>
    </Drawer>
  )
}

function UnreadBadge({ count }: { count: number }) {
  if (count <= 0) return null

  return (
    <span className="bg-primary text-primary-foreground pointer-events-none absolute -top-1 -right-1 flex h-4 min-w-4 items-center justify-center rounded-full px-1 text-[10px] leading-none font-medium">
      {count > 99 ? '99+' : count}
    </span>
  )
}

type ListProps = {
  notifications: Notification[]
  isLoading: boolean
  onOpen: (notification: Notification) => void
}

function NotificationList({ notifications, isLoading, onOpen }: ListProps) {
  const t = useTranslations('navbar.notifications')

  if (notifications.length === 0) {
    return (
      <p className="text-muted-foreground p-6 text-center text-sm">
        {isLoading ? t('loading') : t('noUnread')}
      </p>
    )
  }

  return (
    <ul className="divide-y">
      {notifications.map((notification) => (
        <NotificationItem
          key={notification.id}
          notification={notification}
          onOpen={onOpen}
        />
      ))}
    </ul>
  )
}
