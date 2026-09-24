import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from '@/components/ui/breadcrumb'
import { getMetadataLocale } from '@/i18n/metadata-locale'
import { isAdminSession } from '@/lib/admin'
import { getCurrentSession } from '@/lib/get-current-session'
import { getMapDimensions, mapBaseUrl } from '@/lib/squaremap'
import { findActiveWorld } from '@/lib/worlds'
import { MapIcon } from 'lucide-react'
import { Metadata } from 'next'
import { getTranslations } from 'next-intl/server'
import { redirect } from 'next/navigation'
import { cache } from 'react'
import WorldMap from './world-map'

type Props = {
  params: Promise<{
    world: string
    locale: string
  }>
}

const findWorld = cache(findActiveWorld)

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const locale = await getMetadataLocale(params)
  const { world: slug } = await params
  const t = await getTranslations({ locale, namespace: 'worlds.maps' })
  const found = await findWorld(slug).catch(() => undefined)
  const world = found?.world.name ?? slug
  return {
    title: t('metadata.title', { world }),
    description: t('metadata.description', { world }),
    openGraph: {
      type: 'website',
      url: `${process.env.NEXT_PUBLIC_DOMAIN}/worlds/${slug}`,
      title: t('metadata.title', { world }),
      description: t('metadata.description', { world }),
      siteName: 'Lania Network',
    },
  }
}

export default async function WorldMapPage({ params }: Props) {
  const { world: slug, locale } = await params
  const t = await getTranslations({ locale, namespace: 'worlds.maps' })

  const found = await findWorld(slug).catch(() => undefined)
  const mapUrl = mapBaseUrl(found?.world.mapUrl)
  if (!found || !mapUrl) redirect('/worlds')
  const { season, world } = found

  const [dimensions, session] = await Promise.all([
    getMapDimensions(mapUrl).catch(() => []),
    getCurrentSession(),
  ])

  const header = (
    <Breadcrumb>
      <BreadcrumbList>
        <BreadcrumbItem>
          <BreadcrumbLink href="/worlds" className="flex items-center gap-2">
            <MapIcon className="size-3" />
            {t('breadcrumbs.title')}
          </BreadcrumbLink>
        </BreadcrumbItem>
        <BreadcrumbSeparator />
        <BreadcrumbItem>
          <BreadcrumbPage>
            {season.name} · {world.name}
          </BreadcrumbPage>
        </BreadcrumbItem>
      </BreadcrumbList>
    </Breadcrumb>
  )

  if (dimensions.length === 0) {
    return (
      <div className="flex h-full flex-col">
        <div className="flex min-h-11 items-center px-4 py-1.5">{header}</div>
        <p className="text-muted-foreground m-4 rounded-lg border p-6 text-center">
          {t('unavailable')}
        </p>
      </div>
    )
  }

  return (
    <WorldMap
      world={world}
      mapUrl={mapUrl}
      dimensions={dimensions}
      isAdmin={isAdminSession(session)}
      header={header}
    />
  )
}
