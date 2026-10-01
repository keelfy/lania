import { Badge } from '@/components/ui/badge'
import { getProfileStats } from '@/lib/api-endpoints'
import { formatPlaytime } from '@/lib/playtime'
import { serverApiFetcher } from '@/lib/server'
import { cn } from '@/lib/utils'
import type { ProfileSeasonStats as SeasonStats } from '@/models/profile'
import { ChevronDownIcon } from 'lucide-react'
import { getTranslations } from 'next-intl/server'

type Props = React.ComponentProps<'section'> & {
  profileId: string
  locale: string
}

// Glyphs are drawn on a pixel grid like the game's own icons. '#' is a filled pixel.
const SKULL = [
  ' ###### ',
  '########',
  '#  ##  #',
  '#  ##  #',
  '########',
  '###  ###',
  ' ###### ',
  ' # ## # ',
]
const SWORD = [
  '     ###',
  '    ####',
  '   #### ',
  '#  ###  ',
  ' ####   ',
  '  ##    ',
  ' ## #   ',
  '##      ',
]

function PixelGlyph({ pixels }: { pixels: string[] }) {
  const path = pixels
    .flatMap((row, y) =>
      [...row].map((cell, x) => (cell === '#' ? `M${x} ${y}h1v1h-1z` : '')),
    )
    .join('')
  return (
    <svg
      aria-hidden
      viewBox="0 0 8 8"
      shapeRendering="crispEdges"
      className="size-3.5 shrink-0 fill-current"
    >
      <path d={path} />
    </svg>
  )
}

// Stats of a profile in every season it played in.
export default async function ProfileSeasonStats({
  profileId,
  locale,
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

  const count = (chunks: React.ReactNode) => (
    <span className="text-foreground font-semibold tabular-nums">{chunks}</span>
  )

  const total = playtimeText(stats?.totalPlaytime ?? 0)

  const renderSeason = (season: SeasonStats) => {
    const playtime = playtimeText(season.playtime)
    const running = season.isActive && season.endDate === undefined
    const hasCounters = season.deaths > 0 || season.mobKills > 0

    return (
      <li
        key={season.seasonId}
        className={cn(
          'flex flex-col gap-2 rounded-lg px-3 py-4',
          running && 'bg-muted/50',
        )}
      >
        <div className="flex items-start justify-between gap-4">
          <div className="flex min-w-0 flex-col gap-1">
            <div className="flex flex-wrap items-center gap-2">
              <span className="font-semibold break-words">
                {season.seasonName}
              </span>
              {running && (
                <Badge variant="outline" className="shrink-0">
                  {t('running')}
                </Badge>
              )}
            </div>
            <span className="text-muted-foreground text-sm">
              {season.endDate === undefined
                ? t('since', { date: month(season.startDate) })
                : `${month(season.startDate)} — ${month(season.endDate)}`}
            </span>
          </div>
          <span className="shrink-0 text-lg font-semibold tabular-nums">
            {playtime.value.toLocaleString(locale)}
            <span className="text-muted-foreground ml-1 text-sm font-normal">
              {playtime.unit}
            </span>
          </span>
        </div>
        {hasCounters && (
          <p className="text-muted-foreground flex flex-wrap items-center gap-x-3 gap-y-1 text-sm">
            <span className="flex items-center gap-1.5">
              <PixelGlyph pixels={SKULL} />
              <span>
                {t.rich('deaths', { count: season.deaths, n: count })}
              </span>
            </span>
            <span aria-hidden>·</span>
            <span className="flex items-center gap-1.5">
              <PixelGlyph pixels={SWORD} />
              <span>
                {t.rich('mobKills', { count: season.mobKills, n: count })}
              </span>
            </span>
          </p>
        )}
      </li>
    )
  }

  return (
    <section className={cn('flex flex-col gap-4', className)} {...props}>
      <div className="flex flex-col gap-1">
        <h2 className="text-xl font-bold tracking-tight">{t('title')}</h2>
        {stats && stats.seasons.length > 0 && (
          <p className="text-muted-foreground text-sm">
            {t.rich('summary', {
              value: total.value,
              unit: total.unit,
              count: stats.seasons.length,
              n: count,
            })}
          </p>
        )}
      </div>

      {!stats ? (
        <p className="text-muted-foreground py-4 text-sm">{t('loadError')}</p>
      ) : stats.seasons.length === 0 ? (
        <p className="text-muted-foreground py-4 text-sm">{t('empty')}</p>
      ) : (
        <div>
          <ol className="divide-y">
            {stats.seasons.slice(0, 3).map(renderSeason)}
          </ol>
          {stats.seasons.length > 3 && (
            <details className="group border-t">
              <summary className="text-muted-foreground hover:text-foreground focus-visible:ring-ring flex cursor-pointer list-none items-center justify-between gap-3 rounded-md px-3 py-3 text-sm font-medium transition-colors focus-visible:ring-2 focus-visible:outline-none [&::-webkit-details-marker]:hidden">
                <span className="group-open:hidden">
                  {t('showAll', { count: stats.seasons.length })}
                </span>
                <span className="hidden group-open:inline">
                  {t('showLess')}
                </span>
                <ChevronDownIcon className="size-4 shrink-0 transition-transform group-open:rotate-180 motion-reduce:transition-none" />
              </summary>
              <ol start={4} className="divide-y">
                {stats.seasons.slice(3).map(renderSeason)}
              </ol>
            </details>
          )}
        </div>
      )}
    </section>
  )
}
