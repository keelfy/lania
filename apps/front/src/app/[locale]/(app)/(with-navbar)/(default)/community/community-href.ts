export const DEFAULT_COMMUNITY_SORT = 'created_at.asc'

export type CommunityFilters = {
  online?: boolean
  staff?: boolean
}

type CommunityHrefParams = CommunityFilters & {
  locale: string
  sort?: string
  search?: string
  page?: number
}

export function communityHref({
  locale,
  sort = DEFAULT_COMMUNITY_SORT,
  search,
  online,
  staff,
  page = 0,
}: CommunityHrefParams): string {
  const params = new URLSearchParams()
  params.set('sort', sort)
  if (search) params.set('q', search)
  if (online) params.set('online', 'true')
  if (staff) params.set('staff', 'true')
  params.set('page', page.toString())
  return `/${locale}/community?${params.toString()}`
}
