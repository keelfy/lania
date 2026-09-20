import {
  AdminCosmeticsCatalog,
  AdminGrant,
  AdminProfile,
  AdminProfileDetails,
  AdminSeason,
  AdminUser,
  AdminUserDetails,
  GrantCosmeticReq,
  GrantProductReq,
  GrantType,
} from '@/models/admin'
import { BasketItem } from '@/models/basket'
import {
  CreateOrderReq,
  CreateOrderRes,
  Order,
  PurchasedProduct,
} from '@/models/order'
import { Product, ProductMetadata } from '@/models/product'
import {
  Profile,
  PublicProfile,
  ProfileCosmeticOptions,
  ProfileDetails,
  ProfilesStats,
  SelectCosmeticOptionReq,
  UsernameCheck,
} from '@/models/profile'
import { ApiFetcher } from './fetcher'
import { Paginated, TokenPaginated } from '@/models/types'

export function getUserProfiles(
  fetcher: ApiFetcher,
  userId: string = 'undefined',
): Promise<Profile[]> {
  return fetcher<Profile[]>(`/v1/users/${userId}/profiles`)
}

export function getProfiles(
  fetcher: ApiFetcher,
  col: string = 'created_at',
  dir: string = 'asc',
  page: number = 0,
  search: string = '',
  size: number = 40,
  onlineOnly: boolean = false,
  staffOnly: boolean = false,
): Promise<Paginated<PublicProfile>> {
  const params = new URLSearchParams()
  params.set('column', col)
  params.set('direction', dir)
  params.set('page', page.toString())
  params.set('size', size.toString())
  if (search) params.set('search', search)
  if (onlineOnly) params.set('online', 'true')
  if (staffOnly) params.set('staff', 'true')
  return fetcher<Paginated<PublicProfile>>('/v1/profiles', params)
}

// Playtime of the returned profiles is counted in the active season only.
export function getTopPlaytimeProfiles(
  fetcher: ApiFetcher,
  limit: number = 10,
): Promise<PublicProfile[]> {
  const params = new URLSearchParams()
  params.set('limit', limit.toString())
  return fetcher<PublicProfile[]>('/v1/profiles/top-playtime', params)
}

export function getProfilesStats(fetcher: ApiFetcher): Promise<ProfilesStats> {
  return fetcher<ProfilesStats>('/v1/profiles/stats')
}

export function getProfileCosmeticOptions(
  fetcher: ApiFetcher,
  id: string,
): Promise<ProfileCosmeticOptions> {
  return fetcher<ProfileCosmeticOptions>(`/v1/profiles/${id}/cosmetics/options`)
}

export function getBasket(fetcher: ApiFetcher): Promise<BasketItem[]> {
  return fetcher<BasketItem[]>('/v1/basket')
}

export function getPurchases(
  fetcher: ApiFetcher,
  productIds: string[],
): Promise<PurchasedProduct[]> {
  const params = new URLSearchParams()
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
): Promise<UsernameCheck[]> {
  return fetcher<UsernameCheck[]>(`/v1/profiles/check-username/${username}`)
}

export function requestFreeAccess(
  fetcher: ApiFetcher,
  usernames: string[],
): Promise<void> {
  const activeSeason = process.env.NEXT_PUBLIC_ACTIVE_SEASON_ID
  if (!activeSeason || activeSeason.length === 0) {
    throw new Error('No active season')
  }

  const params = new URLSearchParams()
  params.set('username', usernames.join(','))
  return fetcher<void>(
    `/v1/seasons/${activeSeason}/access/pre-register`,
    params,
    {
      method: 'POST',
    },
  )
}

export function registerForActiveSeason(
  fetcher: ApiFetcher,
  usernames: string[],
): Promise<void> {
  const activeSeason = process.env.NEXT_PUBLIC_ACTIVE_SEASON_ID
  if (!activeSeason || activeSeason.length === 0) {
    throw new Error('No active season')
  }

  const params = new URLSearchParams()
  params.set('username', usernames.join(','))
  return fetcher<void>(`/v1/seasons/${activeSeason}/access/register`, params, {
    method: 'POST',
  })
}

export function requestAccess(
  fetcher: ApiFetcher,
  usernames: string[],
): Promise<void> {
  const activeSeason = process.env.NEXT_PUBLIC_ACTIVE_SEASON_ID
  if (!activeSeason || activeSeason.length === 0) {
    throw new Error('No active season')
  }

  const params = new URLSearchParams()
  params.set('username', usernames.join(','))
  return fetcher<void>(`/v1/seasons/${activeSeason}/get-access`, params, {
    method: 'POST',
  })
}

export function getProfileDetails(
  fetcher: ApiFetcher,
  id: string,
): Promise<ProfileDetails> {
  return fetcher<ProfileDetails>(`/v1/profiles/${id}`)
}

export function getProfileDetailsByUsername(
  fetcher: ApiFetcher,
  username: string,
): Promise<ProfileDetails> {
  return fetcher<ProfileDetails>(
    `/v1/profiles/by-username/${encodeURIComponent(username)}`,
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
): Promise<void> {
  return fetcher<void>(`/v1/profiles/${id}/cosmetics/name-color`, undefined, {
    method: 'POST',
    body: JSON.stringify(option),
  })
}

export function updateProfileNamePrefix(
  fetcher: ApiFetcher,
  id: string,
  type: 'glyth' | 'special',
  option: SelectCosmeticOptionReq,
): Promise<void> {
  return fetcher<void>(
    `/v1/profiles/${id}/cosmetics/name-prefix/${type}`,
    undefined,
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
): Promise<void> {
  return fetcher<void>(`/v1/basket`, undefined, {
    method: 'POST',
    body: JSON.stringify({ productId, profileId }),
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

export function getAdminSeasons(fetcher: ApiFetcher): Promise<AdminSeason[]> {
  return fetcher<AdminSeason[]>('/v1/admin/seasons')
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
