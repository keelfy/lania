import TopPlayers from '@/components/top-players'
import { getProfilesFeed, getSeasons } from '@/lib/api-endpoints'
import { ProfileFeed } from '@/models/profile'
import { pickSeason } from '@/lib/seasons'
import { serverApiFetcher } from '@/lib/server'
import { getTranslations } from 'next-intl/server'
import {
  communityHref,
  communityProfileHref,
  DEFAULT_COMMUNITY_SORT,
} from './community-href'
import CommunityFilters from './community-filters'
import CommunityOnlineNow from './community-online-now'
import CommunityStats from './community-stats'
import CommunityPlayerList from './community-player-list'
import CommunitySearch from './community-search'
import SelectCommunitySeason from './select-community-season'
import SelectCommunitySort from './select-profile-sort'
import Link from 'next/link'
import { buttonVariants } from '@/components/ui/button'
import { SearchXIcon } from 'lucide-react'

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
  }>
}

export default async function CommunityPage({ params, searchParams }: Props) {
  const { locale } = await params
  const {
    sort: sortParam,
    q: searchParam,
    online: onlineParam,
    staff: staffParam,
    season: seasonParam,
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
  const [col, dir] = sort.split('.')
  // The online and staff filters only affect the player list below. Searching
  // still hides the summary blocks so they do not get in the way.
  const showSummary = !search

  const firstPage: ProfileFeed = await getProfilesFeed(serverApiFetcher, {
    col,
    dir,
    search,
    onlineOnly: online,
    staffOnly: staff,
    seasonId: contextSeason?.id,
  }).catch((err) => {
    console.error(err)
    return { content: [], totalElements: 0 }
  })
  return (
    <div className="flex flex-col gap-6">
      <header className="flex flex-wrap items-center justify-between gap-4">
        <h1 className="text-4xl font-extrabold tracking-tight">{t('title')}</h1>
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
      </header>
      <CommunityStats
        locale={locale}
        season={season}
        seasonName={contextSeason?.name}
        onlineAvailable={onlineAvailable}
        sort={sort}
        search={search}
        staff={staff}
        online={online}
      />
      {showSummary && onlineAvailable && (
        <CommunityOnlineNow locale={locale} season={season} />
      )}
      {showSummary && (
        <TopPlayers
          locale={locale}
          title={t('top.title')}
          season={season}
          href={(profile) =>
            communityProfileHref(locale, profile.username, season)
          }
        />
      )}
      <section
        id="players"
        className="flex scroll-mt-6 flex-col gap-4"
        aria-labelledby="players-heading"
      >
        <div className="flex flex-wrap items-baseline justify-between gap-2">
          <h2
            id="players-heading"
            className="text-2xl font-bold tracking-tight"
          >
            {t('players')}
          </h2>
          <p className="text-muted-foreground text-sm" role="status">
            {t('results', { count: firstPage.totalElements })}
          </p>
        </div>
        <div className="bg-card flex flex-wrap items-center gap-3 rounded-xl border p-3">
          <CommunitySearch
            defaultValue={search}
            sort={sort}
            locale={locale}
            online={online}
            staff={staff}
            season={season}
          />
          <CommunityFilters
            sort={sort}
            search={search}
            locale={locale}
            online={online}
            staff={staff}
            season={season}
            onlineAvailable={onlineAvailable}
          />
          <SelectCommunitySort
            key={sort}
            defaultValue={sort}
            search={search}
            locale={locale}
            online={online}
            staff={staff}
            season={season}
          />
          {(search || online || staff) && (
            <Link
              className={buttonVariants({ variant: 'ghost', size: 'sm' })}
              href={communityHref({ locale, sort, season })}
            >
              {t('resetFilters')}
            </Link>
          )}
        </div>
        {firstPage.content.length > 0 ? (
          <CommunityPlayerList
            // A new list starts from the first page again.
            key={[sort, search, online, staff, contextSeason?.id].join(':')}
            profiles={firstPage.content}
            nextCursor={firstPage.nextCursor}
            query={{
              sort,
              search,
              online,
              staff,
              seasonId: contextSeason?.id,
            }}
            locale={locale}
            season={season}
          />
        ) : (
          <div className="bg-card flex flex-col items-center gap-3 rounded-xl border border-dashed px-4 py-12 text-center">
            <SearchXIcon className="text-muted-foreground size-7" aria-hidden />
            <p className="font-medium">{t('noResults')}</p>
            <p className="text-muted-foreground max-w-sm text-sm">
              {t('noResultsHint')}
            </p>
            {(search || online || staff) && (
              <Link
                className={buttonVariants({ variant: 'outline', size: 'sm' })}
                href={communityHref({ locale, sort, season })}
              >
                {t('resetFilters')}
              </Link>
            )}
          </div>
        )}
      </section>
    </div>
  )
}
