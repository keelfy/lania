import { ProfileResync, ProfileRole } from './profile'

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

export type ProfileMergeBlockerKind =
  | 'same-profile'
  | 'different-owners'
  | 'live-season-playtime'

export type ProfileMergeBlocker = {
  kind: ProfileMergeBlockerKind
  // Set for live-season-playtime only.
  seasonNames?: string[]
}

export type ProfileMergeCounts = {
  playtimeMoved: number
  playtimeSummed: number
  accessesMoved: number
  accessesDropped: number
  violationsMoved: number
  nameColorOptionsMoved: number
  nameColorOptionsDropped: number
  namePrefixOptionsMoved: number
  namePrefixOptionsDropped: number
  seasonCosmeticsMoved: number
  seasonCosmeticsDropped: number
  prefixesMoved: number
  prefixesDropped: number
  orderItemsMoved: number
  basketItemsMoved: number
  basketItemsDropped: number
  notificationsRepointed: number
  screenshotAuthorsMoved: number
  screenshotAuthorsDropped: number
}

// What merging the source profile into the target would do (a preview) or did (the real merge).
export type ProfileMergeSummary = {
  sourceProfileId: string
  sourceUsername: string
  targetProfileId: string
  targetUsername: string
  roleBefore: ProfileRole
  roleAfter: ProfileRole
  ownerUserId?: string
  counts: ProfileMergeCounts
  // Empty when nothing stops the merge.
  blockers?: ProfileMergeBlocker[]
  canMerge: boolean
  // The report of updating the season servers. Missing for a preview.
  resync?: ProfileResync
}

// One profile merged into another one and then deleted, kept for support history.
export type ProfileMerge = {
  id: string
  sourceProfileId: string
  sourceMcUuid: string
  sourceUsername: string
  mergedBy: string
  createdAt: number
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
  prefix: string
  noSpace: boolean
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

export type SaveNameColor = Pick<AdminNameColor, 'name' | 'colors'>
export type SaveNamePrefix = Pick<
  AdminNamePrefix,
  'name' | 'prefix' | 'image' | 'noSpace'
>

export type AdminProductLocalization = {
  locale: 'ru' | 'en'
  name: string
  description: string
}

export type AdminProductPrice = {
  currency: string
  amount: number
}

export type AdminProduct = {
  id: string
  category: 'upgrade' | 'name-color' | 'name-prefix'
  priceName: 'season_access' | 'name_color' | 'name_prefix'
  metadata: { action?: string; nameColorId?: string; namePrefixId?: string }
  isActive: boolean
  easyDonateProductId?: number
  soldCount: number
  localizations: AdminProductLocalization[]
  prices: AdminProductPrice[]
}

// The result of an admin uploading an image, ready to store as an image field.
export type UploadedImage = {
  location: string
  width: number
  height: number
}

export type SaveProduct = {
  category: AdminProduct['category']
  cosmeticId?: string
  priceName: AdminProduct['priceName']
  isActive: boolean
  easyDonateProductId?: number
  localizations: AdminProductLocalization[]
}

export type EasyDonateProduct = {
  easyDonateProductId: number
}

export type CreateEasyDonateProduct = {
  userAuth: string
  name: string
  description: string
  priceName: AdminProduct['priceName']
  image?: File
}
