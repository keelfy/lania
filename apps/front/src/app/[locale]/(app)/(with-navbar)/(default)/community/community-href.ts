export const DEFAULT_COMMUNITY_SORT = 'created_at.desc'

export type CommunityFilters = {
  online?: boolean
  staff?: boolean
  // The season the page is shown for: its cosmetics, last seen dates and online status.
  // Missing means the primary season, and keeps the URL short.
  season?: string
}

type CommunityHrefParams = CommunityFilters & {
  locale: string
  sort?: string
  search?: string
  // Zero-based, the URL shows it starting from 1.
  page?: number
}

export function communityHref({
  locale,
  sort = DEFAULT_COMMUNITY_SORT,
  search,
  online,
  staff,
  season,
  page = 0,
}: CommunityHrefParams): string {
  const params = new URLSearchParams()
  params.set('sort', sort)
  if (season) params.set('season', season)
  if (search) params.set('q', search)
  if (online) params.set('online', 'true')
  if (staff) params.set('staff', 'true')
  params.set('page', (page + 1).toString())
  return `/${locale}/community?${params.toString()}`
}

// The profile page of a player, in the same season as the list it was opened from.
export function communityProfileHref(
  locale: string,
  username: string,
  season?: string,
): string {
  const href = `/${locale}/community/${encodeURIComponent(username)}`
  return season ? `${href}?season=${season}` : href
}
