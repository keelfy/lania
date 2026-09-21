import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { getProfileStats } from '@/lib/api-endpoints'
import { formatPlaytime } from '@/lib/playtime'
import { serverApiFetcher } from '@/lib/server'
import { cn } from '@/lib/utils'
import { getTranslations } from 'next-intl/server'
import Link from 'next/link'

type Props = React.ComponentProps<'section'> & {
  profileId: string
  // The name colors of the profile. The playtime bars are drawn in them.
  colors?: string[]
  locale: string
} & (
    | { variant?: 'full'; username?: string }
    // The summary only shows the total and links to the public profile, where the full list is.
    | { variant: 'summary'; username: string }
  )

// The bar is cut into blocks, like the experience bar in the game.
const BLOCKS_MASK =
  'repeating-linear-gradient(to right, #000 0 6px, transparent 6px 8px)'

// The gradient spans the whole track, so the same color always means the same place on the bar.
// A bar only shows the left part of it.
function PlaytimeBar({ ratio, colors }: { ratio: number; colors: string[] }) {
  const fill =
    colors.length > 1
      ? `linear-gradient(to right, ${colors.join(', ')})`
      : (colors[0] ?? 'var(--primary)')
  return (
    <div
      aria-hidden
      className="bg-muted h-3 w-full"
      style={{ maskImage: BLOCKS_MASK, WebkitMaskImage: BLOCKS_MASK }}
    >
      <div
        className="h-full"
        style={{
          background: fill,
          clipPath: `inset(0 ${(1 - ratio) * 100}% 0 0)`,
        }}
      />
    </div>
  )
}

// Playtime of a profile in every season it played in.
export default async function ProfileSeasonStats({
  profileId,
  colors = [],
  locale,
  variant = 'full',
  username,
  className,
  ...props
}: Props) {
  const t = await getTranslations({
    locale,
    namespace: 'community.profile.seasons',
  })
  const tUnits = await getTranslations({
    locale,
    namespace: 'playerCard.playtime',
  })
  const stats = await getProfileStats(serverApiFetcher, profileId).catch(
    (error) => {
      console.error(error)
      return undefined
    },
  )

  const month = (millis: number) =>
    new Date(millis).toLocaleDateString(locale, {
      month: 'short',
      year: 'numeric',
      timeZone: 'UTC',
    })
  const playtimeText = (millis: number) => {
    const { value, unit } = formatPlaytime(millis)
    return { value, unit: tUnits(unit) }
  }

  const longest = Math.max(...(stats?.seasons.map((s) => s.playtime) ?? [0]))
  const total = playtimeText(stats?.totalPlaytime ?? 0)

  if (variant === 'summary') {
    return (
      <section
        className={cn('flex items-center justify-between gap-4', className)}
        {...props}
      >
        <div className="flex flex-col gap-1">
          <h2 className="text-lg font-semibold tracking-tight">
            {t('summaryTitle')}
          </h2>
          <p className="text-muted-foreground text-sm">
            {!stats ? (
              t('loadError')
            ) : stats.seasons.length === 0 ? (
              t('empty')
            ) : (
              <>
                {t('total')}&nbsp;
                <span className="text-foreground font-semibold tabular-nums">
                  {total.value}
                </span>
                &nbsp;{total.unit}
              </>
            )}
          </p>
        </div>
        <Button asChild variant="outline" size="sm" className="shrink-0">
          <Link href={`/${locale}/community/${username}`}>
            {t('viewPublic')}
          </Link>
        </Button>
      </section>
    )
  }

  return (
    <section className={cn('flex flex-col gap-2', className)} {...props}>
      <div className="flex items-baseline justify-between gap-4">
        <h2 className="text-xl font-bold tracking-tight">{t('title')}</h2>
        {stats && stats.seasons.length > 0 && (
          <p className="text-muted-foreground text-sm">
            {t('total')}&nbsp;
            <span className="text-foreground font-semibold tabular-nums">
              {total.value}
            </span>
            &nbsp;{total.unit}
          </p>
        )}
      </div>

      {!stats ? (
        <p className="text-muted-foreground py-4 text-sm">{t('loadError')}</p>
      ) : stats.seasons.length === 0 ? (
        <p className="text-muted-foreground py-4 text-sm">{t('empty')}</p>
      ) : (
        <ol className="divide-y">
          {stats.seasons.map((season) => {
            const playtime = playtimeText(season.playtime)
            const running = season.isActive && season.endDate === undefined
            return (
              <li key={season.seasonId} className="flex flex-col gap-2 py-3">
                <div className="flex items-baseline justify-between gap-4">
                  <div className="flex min-w-0 flex-col">
                    <div className="flex items-center gap-2">
                      <span className="truncate font-semibold">
                        {season.seasonName}
                      </span>
                      {running && (
                        <Badge variant="outline" className="shrink-0">
                          {t('running')}
                        </Badge>
                      )}
                    </div>
                    <span className="text-muted-foreground text-xs">
                      {season.endDate === undefined
                        ? t('since', { date: month(season.startDate) })
                        : `${month(season.startDate)} — ${month(season.endDate)}`}
                    </span>
                  </div>
                  <span className="shrink-0 text-lg font-semibold tabular-nums">
                    {playtime.value}
                    <span className="text-muted-foreground ml-1 text-sm font-normal">
                      {playtime.unit}
                    </span>
                  </span>
                </div>
                <PlaytimeBar
                  ratio={longest > 0 ? season.playtime / longest : 0}
                  colors={colors}
                />
              </li>
            )
          })}
        </ol>
      )}
    </section>
  )
}
