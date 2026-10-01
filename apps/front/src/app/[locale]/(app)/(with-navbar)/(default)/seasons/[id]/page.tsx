import TopPlayers from '@/components/top-players'
import { Badge } from '@/components/ui/badge'
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from '@/components/ui/breadcrumb'
import { Button } from '@/components/ui/button'
import { getCachedSeasons, getCachedSeasonScreenshots } from '@/lib/public-data'
import { formatSeasonDuration } from '@/lib/seasons'
import { CalendarIcon, ClockIcon, DownloadIcon } from 'lucide-react'
import { Metadata } from 'next'
import { getTranslations } from 'next-intl/server'
import Image from 'next/image'
import { notFound } from 'next/navigation'
import { communityProfileHref } from '../../community/community-href'
import SeasonScreenshots from './season-screenshots'

type Props = {
  params: Promise<{ locale: string; id: string }>
}

// The seasons that exist at build time are prerendered, a season added later is rendered on its first visit.
export async function generateStaticParams() {
  const seasons = await getCachedSeasons()
  return seasons
    .filter((season) => season.previewImage)
    .map((season) => ({ id: season.id }))
}

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { id } = await params
  const seasons = await getCachedSeasons().catch(() => [])
  const season = seasons.find((season) => season.id === id)
  if (!season) return {}

  return {
    title: season.name,
    openGraph: {
      type: 'website',
      url: `${process.env.NEXT_PUBLIC_DOMAIN}/seasons/${season.id}`,
      title: season.name,
      siteName: 'Lania Network',
    },
  }
}

export default async function SeasonPage({ params }: Props) {
  const { locale, id } = await params
  const t = await getTranslations({ locale, namespace: 'seasons' })

  const [seasons, screenshots] = await Promise.all([
    getCachedSeasons().catch(() => []),
    getCachedSeasonScreenshots(id, locale).catch(() => []),
  ])
  const season = seasons.find((season) => season.id === id)
  if (!season || !season.previewImage) notFound()

  const toLocalDate = (millis?: number) => {
    if (millis === undefined) return '?'
    return new Date(millis).toLocaleDateString(locale, {
      day: 'numeric',
      month: 'long',
      year: 'numeric',
      timeZone: 'UTC',
    })
  }

  const duration = formatSeasonDuration(season.startDate, season.endDate)
  const durationLabel = duration
    ? [
        duration.months > 0
          ? t('duration.months', { count: duration.months })
          : null,
        duration.days > 0 || duration.months === 0
          ? t('duration.days', { count: duration.days })
          : null,
      ]
        .filter(Boolean)
        .join(' ')
    : undefined

  return (
    <div className="flex flex-col gap-6">
      <Breadcrumb>
        <BreadcrumbList>
          <BreadcrumbItem>
            <BreadcrumbLink href={`/${locale}/seasons`}>
              {t('title')}
            </BreadcrumbLink>
          </BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>
            <BreadcrumbPage>{season.name}</BreadcrumbPage>
          </BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>

      <section className="flex flex-col gap-6">
        <div className="relative isolate flex min-h-60 items-end overflow-hidden rounded-lg bg-slate-950 sm:min-h-72 lg:min-h-80">
          <Image
            src={season.previewImage}
            alt={season.name}
            fill
            sizes="(max-width: 640px) calc(100vw - 32px), (max-width: 1024px) 100vw, 1024px"
            preload
            className="object-cover"
          />
          <div
            aria-hidden="true"
            className="absolute inset-0 bg-linear-to-t from-black/85 via-black/20 to-transparent"
          />
          {season.isActive && (
            <Badge className="absolute top-4 right-4 border-transparent bg-teal-500 text-white sm:top-6 sm:right-6">
              <span className="size-1.5 rounded-full bg-white" />
              {t('active')}
            </Badge>
          )}
          <h1 className="font-minecraft relative w-full px-6 pt-20 pb-5 text-4xl leading-tight font-normal text-balance wrap-anywhere text-white drop-shadow-md sm:px-8 sm:pb-7 sm:text-5xl lg:text-6xl">
            {season.name}
          </h1>
        </div>

        <div className="grid gap-6 sm:grid-cols-[1fr_auto] sm:gap-10">
          <p className="text-muted-foreground max-w-lg leading-relaxed">
            {t(`${season.id}.description`)}
          </p>
          <div className="flex min-w-0 flex-col items-start gap-3 text-sm">
            <div className="flex items-start gap-2">
              <CalendarIcon
                aria-hidden="true"
                className="text-muted-foreground mt-0.5 size-4 shrink-0"
              />
              <span>
                {season.isActive
                  ? `${t('since')} ${toLocalDate(season.startDate)}`
                  : `${toLocalDate(season.startDate)} — ${toLocalDate(season.endDate)}`}
              </span>
            </div>
            {durationLabel && (
              <div className="flex items-center gap-2">
                <ClockIcon
                  aria-hidden="true"
                  className="text-muted-foreground size-4 shrink-0"
                />
                <span>{durationLabel}</span>
              </div>
            )}
            {season.gameVersion && (
              <div className="text-muted-foreground">
                {t('gameVersion')}:{' '}
                <span className="text-foreground">{season.gameVersion}</span>
              </div>
            )}
            {season.publicAddress && (
              <div className="wrap-anywhere">{season.publicAddress}</div>
            )}
            {season.worldUrl && (
              <Button variant="outline" size="sm" className="mt-1" asChild>
                <a href={season.worldUrl} target="_blank" rel="noreferrer">
                  <DownloadIcon aria-hidden="true" className="size-4" />
                  {t('worldDownload')}
                </a>
              </Button>
            )}
          </div>
        </div>
      </section>

      <TopPlayers
        locale={locale}
        title={t('topPlayers')}
        season={season.id}
        href={(profile) =>
          communityProfileHref(locale, profile.username, season.id)
        }
      />

      <section className="flex flex-col gap-3">
        <h2 className="text-2xl font-bold tracking-tight">
          {t('screenshots.title')}
        </h2>
        {screenshots.length > 0 ? (
          <SeasonScreenshots
            locale={locale}
            seasonId={season.id}
            screenshots={screenshots}
          />
        ) : (
          <p className="text-muted-foreground text-sm">
            {t('screenshots.empty')}
          </p>
        )}
      </section>
    </div>
  )
}
