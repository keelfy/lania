'use client'

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
import { useMediaQuery } from '@/lib/use-media-query'
import { useNotifications } from '@/lib/use-notifications'
import { cn } from '@/lib/utils'
import { Notification } from '@/models/notification'
import { Link } from '@/i18n/navigation'
import { useLocale, useTranslations } from 'next-intl'
import React from 'react'
import { formatDateTime } from '../(with-navbar)/(default)/admin/format'

type Props = {
  // sessionActive is false for a visitor that is not signed in: they have no notifications.
  sessionActive: boolean
}

export default function NotificationsMenu({
  children,
  sessionActive,
}: React.PropsWithChildren<Props>) {
  const isDesktop = useMediaQuery('(min-width: 1024px)')
  const t = useTranslations('navbar.notifications')
  const { notifications, unreadCount, isLoading, refresh, markAllRead } =
    useNotifications(sessionActive)

  // Opening the menu shows the newest state and clears the counter.
  const handleOpenChange = (open: boolean) => {
    if (!open) return
    void refresh().then(() => {
      if (unreadCount > 0) void markAllRead()
    })
  }

  const list = (
    <NotificationList
      notifications={notifications}
      isLoading={isLoading}
      emptyLabel={t('noNotifications')}
      loadingLabel={t('loading')}
    />
  )

  if (isDesktop) {
    return (
      <Popover onOpenChange={handleOpenChange}>
        <div className="relative">
          <PopoverTrigger asChild>{children}</PopoverTrigger>
          <UnreadBadge count={unreadCount} />
        </div>
        <PopoverContent className="w-96 p-0">
          <div className="border-b px-4 py-3">
            <p className="text-sm font-medium">{t('title')}</p>
          </div>
          <ScrollArea className="max-h-96">{list}</ScrollArea>
        </PopoverContent>
      </Popover>
    )
  }

  return (
    <Drawer onOpenChange={handleOpenChange}>
      <div className="relative">
        <DrawerTrigger asChild>{children}</DrawerTrigger>
        <UnreadBadge count={unreadCount} />
      </div>
      <DrawerContent>
        <DrawerHeader>
          <DrawerTitle>{t('title')}</DrawerTitle>
        </DrawerHeader>
        <ScrollArea className="max-h-[60vh]">{list}</ScrollArea>
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
  emptyLabel: string
  loadingLabel: string
}

function NotificationList({
  notifications,
  isLoading,
  emptyLabel,
  loadingLabel,
}: ListProps) {
  if (isLoading && notifications.length === 0) {
    return (
      <p className="text-muted-foreground min-h-32 p-4 text-center text-sm">
        {loadingLabel}
      </p>
    )
  }

  if (notifications.length === 0) {
    return (
      <p className="text-muted-foreground min-h-32 p-4 text-center text-sm">
        {emptyLabel}
      </p>
    )
  }

  return (
    <ul className="divide-y">
      {notifications.map((notification) => (
        <NotificationItem key={notification.id} notification={notification} />
      ))}
    </ul>
  )
}

function NotificationItem({ notification }: { notification: Notification }) {
  const t = useTranslations('navbar.notifications.items')
  const locale = useLocale()
  const { payload } = notification
  const key =
    notification.type === 'cosmetic-granted'
      ? 'cosmeticGranted'
      : 'cosmeticRevoked'

  return (
    <li
      className={cn(
        'px-4 py-3 text-sm',
        !notification.readAt && 'bg-accent/40',
      )}
    >
      <p className="font-medium">{t(`${key}.title`)}</p>
      <p className="text-muted-foreground">
        {t(`${key}.body`, {
          item: payload.itemName,
          profile: payload.profileUsername,
        })}
      </p>
      <div className="text-muted-foreground mt-1 flex items-center justify-between gap-2 text-xs">
        <time dateTime={notification.createdAt}>
          {formatDateTime(Date.parse(notification.createdAt), locale)}
        </time>
        <Button variant="link" size="sm" className="h-auto p-0 text-xs" asChild>
          <Link href="/profiles">{t('openProfiles')}</Link>
        </Button>
      </div>
    </li>
  )
}
