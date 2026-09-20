import { ProfileRole } from './profile'

export type AdminUser = {
  id: string
  email: string
  username?: string
  avatarUrl?: string
  role: ProfileRole
  createdAt?: number
}

export type AdminProfile = {
  id: string
  mcUuid: string
  username: string
  ownerUserId?: string
  role: ProfileRole
  createdAt: number
}

export type AdminUserDetails = AdminUser & {
  profiles: AdminProfile[]
}

export type AdminProfileDetails = AdminProfile & {
  // Missing while nobody owns the profile.
  owner?: AdminUser
}

export type AdminSeason = {
  id: string
  seasonNumber: number
  startDate: number
  endDate?: number
  // The season that runs on the game server.
  isActive: boolean
}

export type GrantType = 'access' | 'name-color' | 'name-prefix'

export type AdminGrant = {
  id: string
  type: GrantType
  seasonId?: string
  // Name of the name color or name prefix, missing for access.
  name?: string
  prefixType?: 'glyth' | 'special'
  // How access was obtained, missing for a name color and a name prefix.
  source?: string
  orderItemId?: string
  grantedBy?: string
  createdAt: number
  revokedAt?: number
  revokedBy?: string
}

export type GrantProductReq = {
  productId: string
  seasonId: string
}
