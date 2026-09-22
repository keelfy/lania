import TopPlayers from '@/components/top-players'
import { AspectRatio } from '@/components/ui/aspect-ratio'
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
import { getSeasons, getSeasonScreenshots } from '@/lib/api-endpoints'
import { serverApiFetcher } from '@/lib/server'
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

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { id } = await params
  const seasons = await getSeasons(serverApiFetcher).catch(() => [])
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
    getSeasons(serverApiFetcher).catch(() => []),
    getSeasonScreenshots(serverApiFetcher, id).catch(() => []),
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

      <div className="flex flex-col gap-4">
        <AspectRatio
          ratio={16 / 9}
          className="relative overflow-hidden rounded-md"
        >
          <Image
            src={season.previewImage}
            alt={season.name}
            fill
            className="object-cover"
          />
          {season.isActive && (
            <Badge className="absolute top-2 right-2 border-transparent bg-teal-500 text-white">
              <span className="size-1.5 rounded-full bg-white" />
              {t('active')}
            </Badge>
          )}
        </AspectRatio>
        <h1 className="text-center text-4xl font-bold">{season.name}</h1>
      </div>

      <div className="flex flex-wrap items-center justify-center gap-4">
        <div className="flex items-center gap-1.5 text-sm">
          <CalendarIcon className="text-muted-foreground size-4" />
          {season.isActive
            ? `${t('since')} ${toLocalDate(season.startDate)}`
            : `${toLocalDate(season.startDate)} — ${toLocalDate(season.endDate)}`}
        </div>
        {durationLabel && (
          <div className="flex items-center gap-1.5 text-sm">
            <ClockIcon className="text-muted-foreground size-4" />
            {durationLabel}
          </div>
        )}
        {season.gameVersion && (
          <div className="text-sm">
            {t('gameVersion')}: {season.gameVersion}
          </div>
        )}
        {season.publicAddress && (
          <div className="text-sm">{season.publicAddress}</div>
        )}
        {season.worldUrl && (
          <Button variant="outline" size="sm" asChild>
            <a href={season.worldUrl} target="_blank" rel="noreferrer">
              <DownloadIcon className="size-4" />
              {t('worldDownload')}
            </a>
          </Button>
        )}
      </div>

      <p className="text-muted-foreground mx-auto max-w-2xl text-center">
        {t(`${season.id}.description`)}
      </p>

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
