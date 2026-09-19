import {
  Pagination,
  PaginationContent,
  PaginationEllipsis,
  PaginationItem,
  PaginationLink,
  PaginationNext,
  PaginationPrevious,
} from '@/components/ui/pagination'
import { getProfiles } from '@/lib/api-endpoints'
import { serverApiFetcher } from '@/lib/server'
import { getTranslations } from 'next-intl/server'
import { communityHref, DEFAULT_COMMUNITY_SORT } from './community-href'
import CommunityPlayerList from './community-player-list'
import CommunitySearch from './community-search'
import SelectCommunitySort from './select-profile-sort'

type Props = {
  params: Promise<{
    locale: string
  }>
  searchParams: Promise<{
    sort?: string
    q?: string
    page?: string
  }>
}

function GetPaginationItems({
  page,
  totalPages,
  locale,
  sort,
  search,
}: {
  page: number
  totalPages: number
  locale: string
  sort: string
  search: string
}) {
  const items = []
  const hrefFor = (target: number) =>
    communityHref({ locale, sort, search, page: target })

  if (page > 0) {
    items.push(
      <PaginationItem key="previous">
        <PaginationPrevious href={hrefFor(page - 1)} />
      </PaginationItem>,
    )

    if (page > 2) {
      items.push(
        <PaginationItem key="previous-ellipsis">
          <PaginationEllipsis />
        </PaginationItem>,
      )
    }

    items.push(
      <PaginationItem key="previous-page">
        <PaginationLink href={hrefFor(page - 1)}>{page}</PaginationLink>
      </PaginationItem>,
    )
  }

  items.push(
    <PaginationItem key="current-page">
      <PaginationLink href={hrefFor(page)} isActive>
        {page + 1}
      </PaginationLink>
    </PaginationItem>,
  )

  if (page < totalPages - 1) {
    items.push(
      <PaginationItem key="next-page">
        <PaginationLink href={hrefFor(page + 1)}>{page + 2}</PaginationLink>
      </PaginationItem>,
    )

    if (page < totalPages - 2) {
      items.push(
        <PaginationItem key="next-ellipsis">
          <PaginationEllipsis />
        </PaginationItem>,
      )
    }

    items.push(
      <PaginationItem key="next">
        <PaginationNext href={hrefFor(page + 1)} />
      </PaginationItem>,
    )
  }

  return <>{items}</>
}

export default async function CommunityPage({ params, searchParams }: Props) {
  const { locale } = await params
  const {
    sort: sortParam,
    q: searchParam,
    page: pageParam,
  } = await searchParams
  const t = await getTranslations({ locale, namespace: 'community' })
  const sort = sortParam ?? DEFAULT_COMMUNITY_SORT
  const search = searchParam?.trim() ?? ''
  const page = Math.max(0, parseInt(pageParam ?? '') || 0)
  const [col, dir] = sort.split('.')

  const paginatedProfiles = await getProfiles(
    serverApiFetcher,
    col,
    dir,
    page,
    search,
  ).catch((err) => {
    console.error(err)
    return { content: [], page: 0, size: 0, totalPages: 0, totalElements: 0 }
  })
  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <h1 className="text-4xl font-extrabold tracking-tight">{t('title')}</h1>
        <div className="flex w-full items-center gap-2 sm:w-auto">
          <CommunitySearch defaultValue={search} sort={sort} locale={locale} />
          <SelectCommunitySort
            defaultValue={sort}
            search={search}
            locale={locale}
          />
        </div>
      </div>
      {paginatedProfiles.content.length > 0 ? (
        <CommunityPlayerList profiles={paginatedProfiles.content} />
      ) : (
        <p className="text-muted-foreground py-10 text-center">
          {t('noResults')}
        </p>
      )}
      {paginatedProfiles.totalPages > 1 && (
        <Pagination>
          <PaginationContent>
            <GetPaginationItems
              page={page}
              totalPages={paginatedProfiles.totalPages}
              locale={locale}
              sort={sort}
              search={search}
            />
          </PaginationContent>
        </Pagination>
      )}
    </div>
  )
}
