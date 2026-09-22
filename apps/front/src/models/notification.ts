import { GrantType } from './admin'

export type NotificationType =
  | 'cosmetic-granted'
  | 'cosmetic-revoked'
  | 'profile-merged'

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

// The payload of a profile-merged notification: an admin carried sourceUsername's data into this profile.
// sourceUsername no longer exists as a profile once the merge ran.
export type ProfileMergeNotificationPayload = {
  profileId: string
  profileUsername: string
  sourceUsername: string
}

type NotificationBase = {
  id: string
  readAt?: string
  createdAt: string
}

export type CosmeticNotification = NotificationBase & {
  type: 'cosmetic-granted' | 'cosmetic-revoked'
  payload: CosmeticNotificationPayload
}

export type ProfileMergeNotification = NotificationBase & {
  type: 'profile-merged'
  payload: ProfileMergeNotificationPayload
}

// A discriminated union on type, so narrowing on notification.type also narrows notification.payload.
export type Notification = CosmeticNotification | ProfileMergeNotification

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
