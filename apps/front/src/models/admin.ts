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

export type AdminNameColor = {
  id: string
  name: string
  // Gradient stops, empty for a plain color.
  colors: string[]
}

export type AdminNamePrefix = {
  id: string
  name: string
  image: string
}

// Every name color and name prefix that can be granted, for sale or not.
export type AdminCosmeticsCatalog = {
  nameColors: AdminNameColor[]
  namePrefixes: AdminNamePrefix[]
}

export type GrantCosmeticReq = {
  type: 'name-color' | 'name-prefix'
  itemId: string
  // Used for a name prefix only.
  prefixType?: 'glyth' | 'special'
  // Missing to grant the item for good.
  seasonId?: string
}
