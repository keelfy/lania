import { getMetadataLocale } from '@/i18n/metadata-locale'
import { isAdminSession } from '@/lib/admin'
import { getSeasons } from '@/lib/api-endpoints'
import { getCurrentSession } from '@/lib/get-current-session'
import { serverApiFetcher } from '@/lib/server'
import { MapWorld } from '@/models/claim'
import { Season } from '@/models/season'
import { FlagIcon } from 'lucide-react'
import { Metadata } from 'next'
import { getTranslations } from 'next-intl/server'
import { Suspense } from 'react'
import ClaimsMap from './claims-map'

type Props = {
  params: Promise<{ locale: string }>
}

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const locale = await getMetadataLocale(params)
  const t = await getTranslations({ locale, namespace: 'claims' })
  return { title: t('title'), description: t('description') }
}

type SquaremapSettings = {
  worlds: { name: string; type: MapWorld['type'] }[]
}

type SquaremapWorldSettings = {
  spawn: { x: number; z: number }
  zoom: { max: number; extra: number }
}

const worldOrder: MapWorld['type'][] = ['normal', 'nether', 'the_end']

async function fetchJson<T>(url: string): Promise<T> {
  const response = await fetch(url, { next: { revalidate: 300 } })
  if (!response.ok) throw new Error(`${url}: ${response.status}`)
  return response.json() as Promise<T>
}

// squaremap sends no CORS headers, so its settings are read here and not in the browser.
async function getMapWorlds(mapUrl: string): Promise<MapWorld[]> {
  const settings = await fetchJson<SquaremapSettings>(
    `${mapUrl}/tiles/settings.json`,
  )
  const worlds = await Promise.all(
    settings.worlds.map(async ({ name, type }) => {
      const world = await fetchJson<SquaremapWorldSettings>(
        `${mapUrl}/tiles/${name}/settings.json`,
      )
      return {
        name,
        type,
        maxZoom: world.zoom.max,
        extraZoom: world.zoom.extra,
        spawn: world.spawn,
      }
    }),
  )
  return worlds.sort(
    (a, b) => worldOrder.indexOf(a.type) - worldOrder.indexOf(b.type),
  )
}

// The primary season when it has a map, otherwise any running season that has one.
function pickClaimsSeason(seasons: Season[]) {
  const claimable = seasons.filter((season) => season.isActive && season.mapUrl)
  return claimable.find((season) => season.isPrimary) ?? claimable[0]
}

export default async function ClaimsPage({ params }: Props) {
  const { locale } = await params
  const t = await getTranslations({ locale, namespace: 'claims' })

  const [seasons, session] = await Promise.all([
    getSeasons(serverApiFetcher).catch(() => []),
    getCurrentSession(),
  ])
  const season = pickClaimsSeason(seasons)
  const mapUrl = season?.mapUrl?.replace(/\/+$/, '')
  const worlds = mapUrl ? await getMapWorlds(mapUrl).catch(() => []) : []

  return (
    <div className="flex flex-col gap-4 py-6">
      <div className="flex flex-col gap-1">
        <h1 className="flex items-center gap-2 text-3xl font-extrabold tracking-tight">
          <FlagIcon className="size-7" />
          {t('title')}
        </h1>
        <p className="text-muted-foreground">{t('description')}</p>
      </div>
      {season && mapUrl && worlds.length > 0 ? (
        <Suspense>
          <ClaimsMap
            seasonId={season.id}
            mapUrl={mapUrl}
            claimLimit={season.claimLimit}
            worlds={worlds}
            isAdmin={isAdminSession(session)}
          />
        </Suspense>
      ) : (
        <p className="text-muted-foreground rounded-lg border p-6 text-center">
          {season && mapUrl ? t('mapUnavailable') : t('noSeason')}
        </p>
      )}
    </div>
  )
}
