'use client'

import McUsername from '@/components/ui/mc-username'
import { Link } from '@/i18n/navigation'
import { formatTimeAgo } from '@/lib/time-ago'
import { cn } from '@/lib/utils'
import { Notification } from '@/models/notification'
import { useLocale, useTranslations } from 'next-intl'
import Image from 'next/image'

type Props = {
  notification: Notification
  // markUnread sets an unread notification apart from the read ones around it.
  // The bell menu lists unread notifications only, so it leaves this off.
  markUnread?: boolean
  // onOpen runs when the notification is clicked, before the navigation.
  onOpen: (notification: Notification) => void
}

// NotificationItem is one notification of the bell menu and of the notifications page.
// A click opens the profile settings the notification is about and reads the notification.
export default function NotificationItem({
  notification,
  markUnread = false,
  onOpen,
}: Props) {
  const t = useTranslations('navbar.notifications.items')
  const { payload } = notification
  const granted = notification.type === 'cosmetic-granted'
  const unread = markUnread && !notification.readAt
  const kind =
    payload.grantType === 'name-color'
      ? 'nameColor'
      : payload.prefixType === 'special'
        ? 'special'
        : 'glyth'

  return (
    <li>
      <Link
        href={{
          pathname: '/profiles/settings',
          query: {
            id: payload.profileId,
            ...(payload.seasonId ? { s: payload.seasonId } : {}),
          },
        }}
        onClick={() => onOpen(notification)}
        className={cn(
          'group hover:bg-accent/60 focus-visible:ring-ring relative block px-4 py-3 text-sm outline-none focus-visible:ring-2 focus-visible:ring-inset',
          unread &&
            'bg-accent/40 before:bg-primary before:absolute before:inset-y-0 before:left-0 before:w-0.5',
        )}
      >
        <div className="flex items-start justify-between gap-3">
          {/* The profile name in its new look answers both "what" and "for whom" at a glance. */}
          <div
            className={cn(
              'flex min-w-0 items-center gap-1.5',
              !granted && 'opacity-50 grayscale',
            )}
          >
            {payload.prefixImage && (
              <Image
                src={payload.prefixImage}
                alt=""
                width={18}
                height={18}
                unoptimized
                className="shrink-0"
              />
            )}
            <McUsername
              username={payload.profileUsername}
              colors={payload.colors}
              className="truncate text-base"
            />
          </div>
          <RelativeTime date={notification.createdAt} />
        </div>
        <div className="mt-1 flex items-end justify-between gap-3">
          <div className="min-w-0">
            <p
              className={cn(
                markUnread && !unread
                  ? 'text-muted-foreground'
                  : 'text-foreground',
              )}
            >
              {t(`${granted ? 'granted' : 'revoked'}.${kind}`, {
                item: payload.itemName,
              })}
            </p>
            {payload.seasonName && (
              <p className="text-muted-foreground text-xs">
                {t('season', { season: payload.seasonName })}
              </p>
            )}
          </div>
          {granted && (
            <span className="text-primary shrink-0 text-xs font-medium group-hover:underline">
              {t('install')}
            </span>
          )}
        </div>
      </Link>
    </li>
  )
}

// RelativeTime shows how long ago something happened, the exact moment sits in the tooltip.
function RelativeTime({ date }: { date: string }) {
  const t = useTranslations('playerCard.timeAgo')
  const locale = useLocale()
  const ago = formatTimeAgo(date)

  return (
    <time
      dateTime={date}
      title={new Date(date).toLocaleString(locale, {
        dateStyle: 'medium',
        timeStyle: 'short',
      })}
      className="text-muted-foreground shrink-0 pt-0.5 text-xs"
      suppressHydrationWarning
    >
      {ago.unit === 'now'
        ? t('now')
        : t('ago', { value: ago.value, unit: t(ago.unit) })}
    </time>
  )
}
