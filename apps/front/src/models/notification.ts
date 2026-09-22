import { GrantType } from './admin'

export type NotificationType = 'cosmetic-granted' | 'cosmetic-revoked'

// The payload of a cosmetic-granted and a cosmetic-revoked notification.
// The item name and the profile name are copied in by the backend, so an old
// notification keeps the names it was made with.
export type CosmeticNotificationPayload = {
  profileId: string
  profileUsername: string
  grantType: Extract<GrantType, 'name-color' | 'name-prefix'>
  prefixType?: 'glyth' | 'special'
  itemId: string
  itemName: string
  // colors is set for a name color, prefixImage for a name prefix.
  // Both are missing from the notifications made before they were added.
  colors?: string[]
  prefixImage?: string
  seasonId?: string
  // seasonName is missing from the notifications made before it was added.
  seasonName?: string
}

export type Notification = {
  id: string
  type: NotificationType
  payload: CosmeticNotificationPayload
  readAt?: string
  createdAt: string
}

// unreadCount counts every unread notification, also the ones past the end of content.
export type NotificationList = {
  content: Notification[]
  unreadCount: number
  // hasMore tells that more notifications follow past the end of content.
  hasMore: boolean
}

export type NotificationQuery = {
  // unread leaves out the notifications that were read.
  unread?: boolean
  offset?: number
  limit?: number
}
