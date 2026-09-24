import { getMetadataLocale } from '@/i18n/metadata-locale'
import { isAdminSession } from '@/lib/admin'
import { getSeasons } from '@/lib/api-endpoints'
import { getCurrentSession } from '@/lib/get-current-session'
import { serverApiFetcher } from '@/lib/server'
import { getMapWorlds, mapBaseUrl, pickClaimsSeason } from '@/lib/squaremap'
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

export default async function ClaimsPage({ params }: Props) {
  const { locale } = await params
  const t = await getTranslations({ locale, namespace: 'claims' })

  const [seasons, session] = await Promise.all([
    getSeasons(serverApiFetcher).catch(() => []),
    getCurrentSession(),
  ])
  const season = pickClaimsSeason(seasons)
  const mapUrl = mapBaseUrl(season)
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
