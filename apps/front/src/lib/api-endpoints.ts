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
  SelectCosmeticOptionReq,
  UsernameCheck,
} from '@/models/profile'
import { ApiFetcher } from './fetcher'
import { Paginated } from '@/models/types'

export function getUserProfiles(
  fetcher: ApiFetcher,
  userId: string = 'undefined',
): Promise<Profile[]> {
  return fetcher<Profile[]>(`/v1/users/${userId}/profiles`)
}

export function getProfiles(
  fetcher: ApiFetcher,
  col: string = 'createdAt',
  dir: string = 'asc',
  page: number = 0,
  size: number = 40,
): Promise<Paginated<PublicProfile>> {
  const params = new URLSearchParams()
  params.set('column', col)
  params.set('direction', dir)
  params.set('page', page.toString())
  params.set('size', size.toString())
  return fetcher<Paginated<PublicProfile>>('/v1/profiles', params)
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
