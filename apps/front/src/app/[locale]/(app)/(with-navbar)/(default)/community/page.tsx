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
import CommunityPlayerList from './community-player-list'
import SelectCommunitySort from './select-profile-sort'

type Props = {
  params: Promise<{
    locale: string
  }>
  searchParams: Promise<{
    sort?: string
    page?: string
  }>
}

function GetPaginationItems({
  page,
  totalPages,
  locale,
  sort,
}: {
  page: number
  totalPages: number
  locale: string
  sort: string
}) {
  const items = []

  if (page > 0) {
    items.push(
      <PaginationItem key="previous">
        <PaginationPrevious
          href={`/${locale}/community?sort=${sort}&page=${page - 1}`}
        />
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
        <PaginationLink
          href={`/${locale}/community?sort=${sort}&page=${page - 1}`}
        >
          {page}
        </PaginationLink>
      </PaginationItem>,
    )
  }

  items.push(
    <PaginationItem key="current-page">
      <PaginationLink
        href={`/${locale}/community?sort=${sort}&page=${page}`}
        isActive
      >
        {page + 1}
      </PaginationLink>
    </PaginationItem>,
  )

  if (page < totalPages - 1) {
    items.push(
      <PaginationItem key="next-page">
        <PaginationLink
          href={`/${locale}/community?sort=${sort}&page=${page + 1}`}
        >
          {page + 2}
        </PaginationLink>
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
        <PaginationNext
          href={`/${locale}/community?sort=${sort}&page=${page + 1}`}
        />
      </PaginationItem>,
    )
  }

  return <>{items}</>
}

export default async function CommunityPage({ params, searchParams }: Props) {
  const { locale } = await params
  const { sort: sortParam, page: pageParam } = await searchParams
  const t = await getTranslations({ locale, namespace: 'community' })
  const sort = sortParam ?? 'created_at.asc'
  const page = pageParam ? parseInt(pageParam) : 0
  const [col, dir] = sort.split('.')

  const paginatedProfiles = await getProfiles(
    serverApiFetcher,
    col,
    dir,
    page,
  ).catch((err) => {
    console.error(err)
    return { content: [], page: 0, size: 0, totalPages: 0, totalElements: 0 }
  })
  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center justify-between gap-2">
        <h1 className="text-4xl font-extrabold tracking-tight">{t('title')}</h1>
        <SelectCommunitySort defaultValue={sort} locale={locale} />
      </div>
      <CommunityPlayerList profiles={paginatedProfiles.content} />
      {paginatedProfiles.totalPages > 1 && (
        <Pagination>
          <PaginationContent>
            <GetPaginationItems
              page={page}
              totalPages={paginatedProfiles.totalPages}
              locale={locale}
              sort={sort}
            />
          </PaginationContent>
        </Pagination>
      )}
    </div>
  )
}
