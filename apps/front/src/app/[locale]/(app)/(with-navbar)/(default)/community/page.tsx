import {
  Pagination,
  PaginationContent,
  PaginationEllipsis,
  PaginationItem,
  PaginationLink,
  PaginationNext,
  PaginationPrevious,
} from '@/components/ui/pagination'
import { getProfiles, getSeasons } from '@/lib/api-endpoints'
import { pickSeason } from '@/lib/seasons'
import { serverApiFetcher } from '@/lib/server'
import { getTranslations } from 'next-intl/server'
import { communityHref, DEFAULT_COMMUNITY_SORT } from './community-href'
import CommunityFilters from './community-filters'
import CommunityOnlineNow from './community-online-now'
import CommunityStats from './community-stats'
import CommunityTopPlayers from './community-top-players'
import CommunityPlayerList from './community-player-list'
import CommunitySearch from './community-search'
import SelectCommunitySeason from './select-community-season'
import SelectCommunitySort from './select-profile-sort'

type Props = {
  params: Promise<{
    locale: string
  }>
  searchParams: Promise<{
    sort?: string
    q?: string
    online?: string
    staff?: string
    // The season the page is shown for, the primary one when missing.
    season?: string
    page?: string
  }>
}

function GetPaginationItems({
  page,
  totalPages,
  locale,
  sort,
  search,
  online,
  staff,
  season,
}: {
  page: number
  totalPages: number
  locale: string
  sort: string
  search: string
  online: boolean
  staff: boolean
  season?: string
}) {
  const items = []
  const hrefFor = (target: number) =>
    communityHref({
      locale,
      sort,
      search,
      online,
      staff,
      season,
      page: target,
    })

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
    online: onlineParam,
    staff: staffParam,
    season: seasonParam,
    page: pageParam,
  } = await searchParams
  const t = await getTranslations({ locale, namespace: 'community' })
  const sort = sortParam ?? DEFAULT_COMMUNITY_SORT
  const search = searchParam?.trim() ?? ''
  const staff = staffParam === 'true'
  const seasons = await getSeasons(serverApiFetcher).catch((err) => {
    console.error(err)
    return []
  })
  // Everything on the page is of one season: the requested one, or the primary one.
  const contextSeason = pickSeason(seasons, seasonParam)
  const season = contextSeason?.isPrimary ? undefined : contextSeason?.id
  // Nobody is known to be online in a season without a running server.
  const onlineAvailable = contextSeason?.onlineAvailable ?? true
  const online = onlineParam === 'true' && onlineAvailable
  const page = Math.max(0, (parseInt(pageParam ?? '') || 1) - 1)
  const [col, dir] = sort.split('.')
  // The online ribbon and the season top are only shown on the untouched list, so they do not get in the way of searching.
  const isDefaultView = page === 0 && !search && !online && !staff

  const paginatedProfiles = await getProfiles(
    serverApiFetcher,
    col,
    dir,
    page,
    search,
    undefined,
    online,
    staff,
    contextSeason?.id,
  ).catch((err) => {
    console.error(err)
    return { content: [], page: 0, size: 0, totalPages: 0, totalElements: 0 }
  })
  return (
    <div className="flex flex-col gap-4">
      <h1 className="text-4xl font-extrabold tracking-tight">{t('title')}</h1>
      <CommunityStats locale={locale} season={season} />
      {isDefaultView && onlineAvailable && (
        <CommunityOnlineNow locale={locale} season={season} />
      )}
      {isDefaultView && <CommunityTopPlayers locale={locale} season={season} />}
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div className="flex flex-wrap items-center gap-2">
          {seasons.length > 1 && (
            <SelectCommunitySeason
              seasons={[...seasons].sort(
                (a, b) => Number(b.isPrimary) - Number(a.isPrimary),
              )}
              selectedSeasonId={contextSeason?.id}
              sort={sort}
              search={search}
              locale={locale}
              online={online}
              staff={staff}
            />
          )}
          <CommunityFilters
            sort={sort}
            search={search}
            locale={locale}
            online={online}
            staff={staff}
            season={season}
            onlineAvailable={onlineAvailable}
          />
        </div>
        <div className="flex w-full items-center gap-2 sm:w-auto">
          <CommunitySearch
            defaultValue={search}
            sort={sort}
            locale={locale}
            online={online}
            staff={staff}
            season={season}
          />
          <SelectCommunitySort
            defaultValue={sort}
            search={search}
            locale={locale}
            online={online}
            staff={staff}
            season={season}
          />
        </div>
      </div>
      {paginatedProfiles.content.length > 0 ? (
        <CommunityPlayerList
          profiles={paginatedProfiles.content}
          locale={locale}
          season={season}
        />
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
              online={online}
              staff={staff}
              season={season}
            />
          </PaginationContent>
        </Pagination>
      )}
    </div>
  )
}
