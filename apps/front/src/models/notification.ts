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
  seasonId?: string
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
}
