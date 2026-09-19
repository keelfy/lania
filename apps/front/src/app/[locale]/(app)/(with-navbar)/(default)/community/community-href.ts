export const DEFAULT_COMMUNITY_SORT = 'created_at.asc'

type CommunityHrefParams = {
  locale: string
  sort?: string
  search?: string
  page?: number
}

export function communityHref({
  locale,
  sort = DEFAULT_COMMUNITY_SORT,
  search,
  page = 0,
}: CommunityHrefParams): string {
  const params = new URLSearchParams()
  params.set('sort', sort)
  if (search) params.set('q', search)
  params.set('page', page.toString())
  return `/${locale}/community?${params.toString()}`
}
