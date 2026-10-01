import { getProfilesStats } from '@/lib/api-endpoints'
import { serverApiFetcher } from '@/lib/server'
import { cn } from '@/lib/utils'
import { RadioIcon, UserPlusIcon, UsersIcon } from 'lucide-react'
import { getTranslations } from 'next-intl/server'
import Link from 'next/link'
import { communityHref, type CommunityFilters } from './community-href'

type Props = CommunityFilters & {
  locale: string
  seasonName?: string
  onlineAvailable: boolean
  sort: string
  search: string
}

export default async function CommunityStats({
  locale,
  season,
  seasonName,
  onlineAvailable,
  sort,
  search,
  staff,
}: Props) {
  const t = await getTranslations({ locale, namespace: 'community.stats' })
  const stats = await getProfilesStats(serverApiFetcher, season).catch(
    (err) => {
      console.error(err)
      return null
    },
  )
  if (!stats) return null

  const items = [
    {
      key: 'total',
      value: stats.total,
      icon: UsersIcon,
      context: t('allSeasons'),
    },
    ...(onlineAvailable
      ? [
          {
            key: 'online',
            value: stats.online,
            icon: RadioIcon,
            context: seasonName
              ? t('inSeason', { season: seasonName })
              : undefined,
          },
        ]
      : []),
    {
      key: 'newLastWeek',
      value: stats.newLastWeek,
      icon: UserPlusIcon,
      context: t('allSeasons'),
    },
  ] as const

  return (
    <dl
      className={cn(
        'bg-card grid divide-x rounded-xl border',
        onlineAvailable ? 'grid-cols-3' : 'grid-cols-2',
      )}
    >
      {items.map(({ key, value, icon: Icon, context }) => (
        <div
          key={key}
          className="flex min-w-0 flex-col gap-1 px-3 py-3 sm:px-5 sm:py-4"
        >
          <dt className="text-muted-foreground text-xs sm:text-sm">
            {t(key)}
            {context && (
              <span className="sr-only sm:not-sr-only sm:mt-1 sm:block sm:text-xs">
                {context}
              </span>
            )}
          </dt>
          <dd className="order-first">
            {key === 'online' && value !== undefined ? (
              <Link
                href={`${communityHref({ locale, sort, search, staff, season, online: true })}#players`}
                aria-label={`${t('online')}: ${value}`}
                className="focus-visible:ring-ring flex w-fit items-center gap-2 rounded-sm outline-none focus-visible:ring-2 sm:gap-3"
              >
                <Icon
                  aria-hidden
                  className={cn(
                    'size-4',
                    value > 0 ? 'text-emerald-500' : 'text-muted-foreground',
                  )}
                />
                <span className="text-2xl font-semibold tabular-nums">
                  {value.toLocaleString(locale)}
                </span>
              </Link>
            ) : (
              <span className="flex items-center gap-2 sm:gap-3">
                <Icon
                  aria-hidden
                  className="text-muted-foreground size-4 shrink-0"
                />
                <span
                  className={cn(
                    'font-semibold tabular-nums',
                    value === undefined
                      ? 'text-muted-foreground text-sm sm:text-base'
                      : 'text-2xl',
                  )}
                >
                  {value === undefined
                    ? t('unavailable')
                    : value.toLocaleString(locale)}
                </span>
              </span>
            )}
          </dd>
        </div>
      ))}
    </dl>
  )
}
