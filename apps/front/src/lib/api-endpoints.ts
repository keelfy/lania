import {
  AdminCosmeticsCatalog,
  AdminProduct,
  AdminGrant,
  AdminProfile,
  AdminProfileDetails,
  AdminUser,
  AdminUserDetails,
  GrantCosmeticReq,
  GrantProductReq,
  GrantType,
  SaveNameColor,
  SaveNamePrefix,
  SaveProduct,
  ProfileMerge,
  ProfileMergeSummary,
  UploadedImage,
} from '@/models/admin'
import { BasketItem } from '@/models/basket'
import { NotificationList, NotificationQuery } from '@/models/notification'
import {
  CreateOrderReq,
  CreateOrderRes,
  Order,
  PurchasedProduct,
} from '@/models/order'
import { Product, ProductMetadata } from '@/models/product'
import {
  AdminSeason,
  SaveSeason,
  SaveSeasonScreenshot,
  Season,
  SeasonScreenshot,
} from '@/models/season'
import {
  Profile,
  PublicProfile,
  ProfileCosmeticOptions,
  ProfileDetails,
  ProfileResync,
  ProfileRole,
  ProfilesStats,
  ProfileStats,
  SelectCosmeticOptionReq,
  UsernameCheck,
} from '@/models/profile'
import { ApiFetcher } from './fetcher'
import { Paginated, TokenPaginated } from '@/models/types'

// Cosmetics belong to a season, so the endpoints that show or change them take the season.
// The API falls back to the primary season without a seasonId.
function seasonParams(seasonId?: string): URLSearchParams | undefined {
  return seasonId ? new URLSearchParams({ seasonId }) : undefined
}

export function getUserProfiles(
  fetcher: ApiFetcher,
  userId: string = 'undefined',
  seasonId?: string,
): Promise<Profile[]> {
  return fetcher<Profile[]>(
    `/v1/users/${userId}/profiles`,
    seasonParams(seasonId),
  )
}

export function getProfiles(
  fetcher: ApiFetcher,
  col: string = 'created_at',
  dir: string = 'asc',
  page: number = 0,
  search: string = '',
  size: number = 25,
  onlineOnly: boolean = false,
  staffOnly: boolean = false,
  // The season the cosmetics, the last seen date and the online status are of.
  seasonId?: string,
): Promise<Paginated<PublicProfile>> {
  const params = new URLSearchParams()
  params.set('column', col)
  params.set('direction', dir)
  params.set('page', page.toString())
  params.set('size', size.toString())
  if (search) params.set('search', search)
  if (onlineOnly) params.set('online', 'true')
  if (staffOnly) params.set('staff', 'true')
  if (seasonId) params.set('seasonId', seasonId)
  return fetcher<Paginated<PublicProfile>>('/v1/profiles', params)
}

// Playtime of the returned profiles is counted in the season.
export function getTopPlaytimeProfiles(
  fetcher: ApiFetcher,
  limit: number = 10,
  seasonId?: string,
): Promise<PublicProfile[]> {
  const params = new URLSearchParams()
  params.set('limit', limit.toString())
  if (seasonId) params.set('seasonId', seasonId)
  return fetcher<PublicProfile[]>('/v1/profiles/top-playtime', params)
}

// Online is the number of players on the server of the season, and is missing when that is unknown.
export function getProfilesStats(
  fetcher: ApiFetcher,
  seasonId?: string,
): Promise<ProfilesStats> {
  return fetcher<ProfilesStats>('/v1/profiles/stats', seasonParams(seasonId))
}

export function getProfileCosmeticOptions(
  fetcher: ApiFetcher,
  id: string,
  seasonId?: string,
): Promise<ProfileCosmeticOptions> {
  return fetcher<ProfileCosmeticOptions>(
    `/v1/profiles/${id}/cosmetics/options`,
    seasonParams(seasonId),
  )
}

export function getBasket(fetcher: ApiFetcher): Promise<BasketItem[]> {
  return fetcher<BasketItem[]>('/v1/basket')
}

export function getPurchases(
  fetcher: ApiFetcher,
  productIds: string[],
  seasonId?: string,
): Promise<PurchasedProduct[]> {
  const params = new URLSearchParams()
  if (seasonId) {
    params.set('seasonId', seasonId)
  }
  if (productIds) {
    params.set('productIds', productIds.join(','))
  }
  return fetcher<PurchasedProduct[]>(`/v1/purchases`, params)
}

export function getProductByIDs(
  fetcher: ApiFetcher,
  ids: string[],
  locale: string,
  currency?: string,
): Promise<Product<ProductMetadata>[]> {
  const params = new URLSearchParams()
  if (ids) {
    params.set('ids', ids.join(','))
  }
  if (locale) {
    params.set('locale', locale)
  }
  if (currency) {
    params.set('currency', currency)
  }
  return fetcher<Product<ProductMetadata>[]>(`/v1/products`, params)
}

export function getProducts(
  fetcher: ApiFetcher,
  category: string | undefined,
  locale: string,
  currency?: string,
): Promise<Product<ProductMetadata>[]> {
  const params = new URLSearchParams()
  if (category) {
    params.set('category', category)
  }
  if (locale) {
    params.set('locale', locale)
  }
  if (currency) {
    params.set('currency', currency)
  }
  return fetcher<Product<ProductMetadata>[]>(`/v1/products`, params)
}

export function getProduct<T extends ProductMetadata>(
  fetcher: ApiFetcher,
  id: string,
  locale: string,
  currency?: string,
): Promise<Product<T>> {
  const params = new URLSearchParams()
  if (locale) {
    params.set('locale', locale)
  }
  if (currency) {
    params.set('currency', currency)
  }
  return fetcher<Product<T>>(`/v1/products/${id}`, params)
}

export function getOrders(
  fetcher: ApiFetcher,
  userId: string,
  locale: string,
): Promise<Order[]> {
  return fetcher<Order[]>(
    `/v1/users/${userId}/orders`,
    new URLSearchParams({ locale }),
  )
}

export function checkUsername(
  fetcher: ApiFetcher,
  username: string,
  seasonId?: string,
): Promise<UsernameCheck[]> {
  const params = new URLSearchParams()
  if (seasonId) {
    params.set('seasonId', seasonId)
  }
  return fetcher<UsernameCheck[]>(
    `/v1/profiles/check-username/${username}`,
    params,
  )
}

export function requestFreeAccess(
  fetcher: ApiFetcher,
  seasonId: string,
  usernames: string[],
): Promise<void> {
  const params = new URLSearchParams()
  params.set('username', usernames.join(','))
  return fetcher<void>(`/v1/seasons/${seasonId}/access/pre-register`, params, {
    method: 'POST',
  })
}

export function registerForPrimarySeason(
  fetcher: ApiFetcher,
  seasonId: string,
  usernames: string[],
): Promise<void> {
  const params = new URLSearchParams()
  params.set('username', usernames.join(','))
  return fetcher<void>(`/v1/seasons/${seasonId}/access/register`, params, {
    method: 'POST',
  })
}

export function requestAccess(
  fetcher: ApiFetcher,
  seasonId: string,
  usernames: string[],
): Promise<void> {
  const params = new URLSearchParams()
  params.set('username', usernames.join(','))
  return fetcher<void>(`/v1/seasons/${seasonId}/get-access`, params, {
    method: 'POST',
  })
}

export function getProfileDetails(
  fetcher: ApiFetcher,
  id: string,
  seasonId?: string,
): Promise<ProfileDetails> {
  return fetcher<ProfileDetails>(`/v1/profiles/${id}`, seasonParams(seasonId))
}

// Playtime of the profile in every season it played in.
export function getProfileStats(
  fetcher: ApiFetcher,
  id: string,
): Promise<ProfileStats> {
  return fetcher<ProfileStats>(`/v1/profiles/${id}/stats`)
}

export function getProfileDetailsByUsername(
  fetcher: ApiFetcher,
  username: string,
  seasonId?: string,
): Promise<ProfileDetails> {
  return fetcher<ProfileDetails>(
    `/v1/profiles/by-username/${encodeURIComponent(username)}`,
    seasonParams(seasonId),
  )
}

export function createOrder(
  fetcher: ApiFetcher,
  req: CreateOrderReq,
): Promise<CreateOrderRes> {
  return fetcher<CreateOrderRes>(`/v1/orders`, undefined, {
    method: 'POST',
    body: JSON.stringify(req),
  })
}

export function updateProfileNameColor(
  fetcher: ApiFetcher,
  id: string,
  option: SelectCosmeticOptionReq,
  seasonId?: string,
): Promise<void> {
  return fetcher<void>(
    `/v1/profiles/${id}/cosmetics/name-color`,
    seasonParams(seasonId),
    {
      method: 'POST',
      body: JSON.stringify(option),
    },
  )
}

export function updateProfileNamePrefix(
  fetcher: ApiFetcher,
  id: string,
  type: 'glyth' | 'special',
  option: SelectCosmeticOptionReq,
  seasonId?: string,
): Promise<void> {
  return fetcher<void>(
    `/v1/profiles/${id}/cosmetics/name-prefix/${type}`,
    seasonParams(seasonId),
    {
      method: 'POST',
      body: JSON.stringify(option),
    },
  )
}

export function addToBasket(
  fetcher: ApiFetcher,
  productId: string,
  profileId: string,
  seasonId?: string,
): Promise<void> {
  return fetcher<void>(`/v1/basket`, undefined, {
    method: 'POST',
    // The API falls back to the primary season without a seasonId.
    body: JSON.stringify({ productId, profileId, seasonId }),
  })
}

export function deleteFromBasket(
  fetcher: ApiFetcher,
  ids?: string[],
): Promise<void> {
  const params = new URLSearchParams()
  if (ids) {
    params.set('itemIds', ids.join(','))
  }
  return fetcher<void>(`/v1/basket`, params, {
    method: 'DELETE',
  })
}

export function getOrder(fetcher: ApiFetcher, id: string): Promise<Order> {
  return fetcher<Order>(`/v1/orders/${id}`)
}

export function getAdminUsers(
  fetcher: ApiFetcher,
  search: string = '',
  pageToken: string = '',
  size: number = 25,
): Promise<TokenPaginated<AdminUser>> {
  const params = new URLSearchParams()
  params.set('size', size.toString())
  if (search) params.set('search', search)
  if (pageToken) params.set('pageToken', pageToken)
  return fetcher<TokenPaginated<AdminUser>>('/v1/admin/users', params)
}

export function getAdminUser(
  fetcher: ApiFetcher,
  id: string,
): Promise<AdminUserDetails> {
  return fetcher<AdminUserDetails>(`/v1/admin/users/${id}`)
}

export function getAdminProfiles(
  fetcher: ApiFetcher,
  page: number = 0,
  search: string = '',
  size: number = 25,
): Promise<Paginated<AdminProfile>> {
  const params = new URLSearchParams()
  params.set('column', 'created_at')
  params.set('direction', 'desc')
  params.set('page', page.toString())
  params.set('size', size.toString())
  if (search) params.set('search', search)
  return fetcher<Paginated<AdminProfile>>('/v1/admin/profiles', params)
}

export function getAdminProfile(
  fetcher: ApiFetcher,
  id: string,
): Promise<AdminProfileDetails> {
  return fetcher<AdminProfileDetails>(`/v1/admin/profiles/${id}`)
}

export function transferProfileOwner(
  fetcher: ApiFetcher,
  id: string,
  email: string,
): Promise<AdminProfileDetails> {
  return fetcher<AdminProfileDetails>(
    `/v1/admin/profiles/${id}/owner`,
    undefined,
    {
      method: 'PUT',
      body: JSON.stringify({ email }),
    },
  )
}

export function releaseProfileOwner(
  fetcher: ApiFetcher,
  id: string,
): Promise<AdminProfileDetails> {
  return fetcher<AdminProfileDetails>(
    `/v1/admin/profiles/${id}/owner`,
    undefined,
    { method: 'DELETE' },
  )
}

export function setProfileRole(
  fetcher: ApiFetcher,
  id: string,
  role: ProfileRole,
): Promise<AdminProfileDetails> {
  return fetcher<AdminProfileDetails>(
    `/v1/admin/profiles/${id}/role`,
    undefined,
    {
      method: 'PUT',
      body: JSON.stringify({ role }),
    },
  )
}

// Writes the role, cosmetics and access of the profile to the game servers again.
// Only the owner of the profile can call it, and it has a cooldown.
// The report tells per season what was written, also when some servers failed.
export function resyncProfile(
  fetcher: ApiFetcher,
  id: string,
): Promise<ProfileResync> {
  return fetcher<ProfileResync>(`/v1/profiles/${id}/resync`, undefined, {
    method: 'POST',
  })
}

// The same for admins: any profile and no cooldown.
export function resyncProfileAsAdmin(
  fetcher: ApiFetcher,
  id: string,
): Promise<ProfileResync> {
  return fetcher<ProfileResync>(`/v1/admin/profiles/${id}/resync`, undefined, {
    method: 'POST',
  })
}

// Shows what merging sourceId into targetId would do, blockers included. Changes nothing.
export function previewMergeProfiles(
  fetcher: ApiFetcher,
  sourceId: string,
  targetId: string,
): Promise<ProfileMergeSummary> {
  const params = new URLSearchParams()
  params.set('targetProfileId', targetId)
  return fetcher<ProfileMergeSummary>(
    `/v1/admin/profiles/${sourceId}/merge`,
    params,
  )
}

// Moves every site record of sourceId into targetId, then deletes sourceId.
export function mergeProfiles(
  fetcher: ApiFetcher,
  sourceId: string,
  targetId: string,
): Promise<ProfileMergeSummary> {
  return fetcher<ProfileMergeSummary>(
    `/v1/admin/profiles/${sourceId}/merge`,
    undefined,
    { method: 'POST', body: JSON.stringify({ targetProfileId: targetId }) },
  )
}

// Every profile merged into id, newest first.
export function getProfileMerges(
  fetcher: ApiFetcher,
  id: string,
): Promise<ProfileMerge[]> {
  return fetcher<ProfileMerge[]>(`/v1/admin/profiles/${id}/merges`)
}

// Deletes the account of the signed in user and releases the game profiles.
// It fails with session_refresh_required when the user signed in too long ago.
export function deleteAccount(fetcher: ApiFetcher): Promise<void> {
  return fetcher<void>('/v1/account', undefined, { method: 'DELETE' })
}

export function getSeasons(fetcher: ApiFetcher): Promise<Season[]> {
  return fetcher<Season[]>('/v1/seasons')
}

export async function getPrimarySeason(fetcher: ApiFetcher): Promise<Season> {
  const seasons = await getSeasons(fetcher)
  const primary = seasons.find((season) => season.isPrimary)
  if (!primary) throw new Error('No primary season')
  return primary
}

export function getAdminSeasons(fetcher: ApiFetcher): Promise<AdminSeason[]> {
  return fetcher<AdminSeason[]>('/v1/admin/seasons')
}

export function createSeason(
  fetcher: ApiFetcher,
  season: SaveSeason,
): Promise<AdminSeason> {
  return fetcher<AdminSeason>('/v1/admin/seasons', undefined, {
    method: 'POST',
    body: JSON.stringify(season),
  })
}

export function updateSeason(
  fetcher: ApiFetcher,
  id: string,
  season: SaveSeason,
): Promise<AdminSeason> {
  return fetcher<AdminSeason>(`/v1/admin/seasons/${id}`, undefined, {
    method: 'PUT',
    body: JSON.stringify(season),
  })
}

export function deleteSeason(fetcher: ApiFetcher, id: string): Promise<void> {
  return fetcher<void>(`/v1/admin/seasons/${id}`, undefined, {
    method: 'DELETE',
  })
}

export function uploadSeasonPreview(
  fetcher: ApiFetcher,
  file: File,
): Promise<UploadedImage> {
  const form = new FormData()
  form.set('file', file)
  return fetcher<UploadedImage>('/v1/admin/uploads/season-preview', undefined, {
    method: 'POST',
    body: form,
  })
}

export function getSeasonScreenshots(
  fetcher: ApiFetcher,
  seasonId: string,
): Promise<SeasonScreenshot[]> {
  return fetcher<SeasonScreenshot[]>(`/v1/seasons/${seasonId}/screenshots`)
}

export function createSeasonScreenshot(
  fetcher: ApiFetcher,
  seasonId: string,
  screenshot: SaveSeasonScreenshot,
): Promise<SeasonScreenshot> {
  return fetcher<SeasonScreenshot>(
    `/v1/admin/seasons/${seasonId}/screenshots`,
    undefined,
    { method: 'POST', body: JSON.stringify(screenshot) },
  )
}

export function updateSeasonScreenshot(
  fetcher: ApiFetcher,
  seasonId: string,
  screenshotId: string,
  screenshot: SaveSeasonScreenshot,
): Promise<SeasonScreenshot> {
  return fetcher<SeasonScreenshot>(
    `/v1/admin/seasons/${seasonId}/screenshots/${screenshotId}`,
    undefined,
    { method: 'PUT', body: JSON.stringify(screenshot) },
  )
}

export function deleteSeasonScreenshot(
  fetcher: ApiFetcher,
  seasonId: string,
  screenshotId: string,
): Promise<void> {
  return fetcher<void>(
    `/v1/admin/seasons/${seasonId}/screenshots/${screenshotId}`,
    undefined,
    { method: 'DELETE' },
  )
}

export function uploadSeasonScreenshotImage(
  fetcher: ApiFetcher,
  seasonId: string,
  file: File,
): Promise<UploadedImage> {
  const form = new FormData()
  form.set('file', file)
  return fetcher<UploadedImage>(
    `/v1/admin/seasons/${seasonId}/screenshots/upload`,
    undefined,
    { method: 'POST', body: form },
  )
}

export function getAdminGrants(
  fetcher: ApiFetcher,
  profileId: string,
): Promise<AdminGrant[]> {
  return fetcher<AdminGrant[]>(`/v1/admin/profiles/${profileId}/grants`)
}

// Answers with the grants of the profile after the change.
export function grantProduct(
  fetcher: ApiFetcher,
  profileId: string,
  req: GrantProductReq,
): Promise<AdminGrant[]> {
  return fetcher<AdminGrant[]>(
    `/v1/admin/profiles/${profileId}/grants`,
    undefined,
    {
      method: 'POST',
      body: JSON.stringify(req),
    },
  )
}

export function getAdminCosmetics(
  fetcher: ApiFetcher,
): Promise<AdminCosmeticsCatalog> {
  return fetcher<AdminCosmeticsCatalog>('/v1/admin/cosmetics')
}

export function createNameColor(
  fetcher: ApiFetcher,
  item: SaveNameColor,
): Promise<AdminCosmeticsCatalog> {
  return fetcher('/v1/admin/cosmetics/name-colors', undefined, {
    method: 'POST',
    body: JSON.stringify(item),
  })
}

export function updateNameColor(
  fetcher: ApiFetcher,
  id: string,
  item: SaveNameColor,
): Promise<AdminCosmeticsCatalog> {
  return fetcher(`/v1/admin/cosmetics/name-colors/${id}`, undefined, {
    method: 'PUT',
    body: JSON.stringify(item),
  })
}

export function createNamePrefix(
  fetcher: ApiFetcher,
  item: SaveNamePrefix,
): Promise<AdminCosmeticsCatalog> {
  return fetcher('/v1/admin/cosmetics/name-prefixes', undefined, {
    method: 'POST',
    body: JSON.stringify(item),
  })
}

export function updateNamePrefix(
  fetcher: ApiFetcher,
  id: string,
  item: SaveNamePrefix,
): Promise<AdminCosmeticsCatalog> {
  return fetcher(`/v1/admin/cosmetics/name-prefixes/${id}`, undefined, {
    method: 'PUT',
    body: JSON.stringify(item),
  })
}

// token is the in-game glyth token (":glyth_popcat:"), used to name the object key.
export function uploadGlythPreview(
  fetcher: ApiFetcher,
  file: File,
  token: string,
  name: string,
): Promise<UploadedImage> {
  const form = new FormData()
  form.set('file', file)
  form.set('token', token)
  form.set('name', name)
  return fetcher<UploadedImage>('/v1/admin/uploads/glyth-preview', undefined, {
    method: 'POST',
    body: form,
  })
}

export function getAdminProducts(fetcher: ApiFetcher): Promise<AdminProduct[]> {
  return fetcher('/v1/admin/products')
}

export function createAdminProduct(
  fetcher: ApiFetcher,
  item: SaveProduct,
): Promise<AdminProduct> {
  return fetcher('/v1/admin/products', undefined, {
    method: 'POST',
    body: JSON.stringify(item),
  })
}

export function updateAdminProduct(
  fetcher: ApiFetcher,
  id: string,
  item: SaveProduct,
): Promise<AdminProduct> {
  return fetcher(`/v1/admin/products/${id}`, undefined, {
    method: 'PUT',
    body: JSON.stringify(item),
  })
}

// Answers with the grants of the profile after the change.
export function grantCosmetic(
  fetcher: ApiFetcher,
  profileId: string,
  req: GrantCosmeticReq,
): Promise<AdminGrant[]> {
  return fetcher<AdminGrant[]>(
    `/v1/admin/profiles/${profileId}/grants/cosmetic`,
    undefined,
    {
      method: 'POST',
      body: JSON.stringify(req),
    },
  )
}

// Answers with the grants of the profile after the change. Repeating it for a revoked grant
// only repeats the update of the game server.
export function revokeGrant(
  fetcher: ApiFetcher,
  profileId: string,
  type: GrantType,
  grantId: string,
): Promise<AdminGrant[]> {
  return fetcher<AdminGrant[]>(
    `/v1/admin/profiles/${profileId}/grants/${type}/${grantId}`,
    undefined,
    { method: 'DELETE' },
  )
}

function notificationQueryParams(query: NotificationQuery) {
  const params = new URLSearchParams()
  params.set('limit', (query.limit ?? 20).toString())
  if (query.offset) params.set('offset', query.offset.toString())
  if (query.unread) params.set('unread', 'true')
  return params
}

export function getNotifications(
  fetcher: ApiFetcher,
  query: NotificationQuery = {},
): Promise<NotificationList> {
  return fetcher<NotificationList>(
    '/v1/notifications',
    notificationQueryParams(query),
  )
}

// An empty ids list marks every unread notification as read.
// The answer is the list the query asks for, after the change.
export function markNotificationsRead(
  fetcher: ApiFetcher,
  ids: string[] = [],
  query: NotificationQuery = {},
): Promise<NotificationList> {
  return fetcher<NotificationList>(
    '/v1/notifications/read',
    notificationQueryParams(query),
    {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ ids }),
    },
  )
}
