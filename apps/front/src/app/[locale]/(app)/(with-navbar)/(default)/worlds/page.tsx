import { getMetadataLocale } from '@/i18n/metadata-locale'
import pinger from 'minecraft-pinger'
import { Metadata } from 'next'
import { getTranslations } from 'next-intl/server'
import { readFile } from 'node:fs/promises'
import path from 'node:path'
import { ServerInfo, ServerMapInfo } from '@/models/server'
import ServerStatusWithMaps from './server-status'
import { Noto_Sans } from 'next/font/google'
import { getSeasons } from '@/lib/api-endpoints'
import { serverApiFetcher } from '@/lib/server'
import { Season } from '@/models/season'

type Props = {
  params: Promise<{
    locale: string
  }>
}

const notoSans = Noto_Sans({
  subsets: ['latin'],
})

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const locale = await getMetadataLocale(params)
  const t = await getTranslations({ locale, namespace: 'worlds.metadata' })

  return {
    title: t('title'),
    description: t('description'),
    openGraph: {
      type: 'website',
      url: process.env.NEXT_PUBLIC_DOMAIN + '/worlds',
      title: t('title'),
      description: t('description'),
      siteName: 'Lania Network',
    },
  }
}

export default async function WorldPage({ params }: Props) {
  const { locale } = await params
  const t = await getTranslations({ locale, namespace: 'worlds' })

  const seasons = (
    await getSeasons(serverApiFetcher).catch(() => [] as Season[])
  ).filter((season): season is Season => season.isActive)

  const mapsPerServer = await readFile(
    path.join(process.cwd(), 'public/data/server-maps.json'),
    'utf-8',
  )
    .then((raw) => JSON.parse(raw))
    .catch((error) => {
      console.error('Failed to read server-maps.json', error)
      return {}
    })

  const statuses: { server: Season; status: pinger.Data | undefined }[] = []

  for (const server of seasons) {
    const { publicAddress } = server

    if (publicAddress) {
      let status: pinger.Data | undefined = undefined
      try {
        status = await Promise.race([
          pinger.pingPromise(publicAddress, 25565),
          new Promise<undefined>((_, reject) => setTimeout(reject, 1000)),
        ])
      } catch (error) {
        console.error(error)
      }
      statuses.push({ server, status })
    }
  }

  return (
    <div className="flex flex-col gap-10">
      {statuses.map(({ server, status }) => {
        const seasonMaps = mapsPerServer[server.id] as ServerMapInfo[];
        return (
        <div key={server.id} className="flex flex-col gap-4">
          <h1 className={`text-3xl font-extrabold ${notoSans.className} antialiased`}>
            {server.name}
          </h1>
          <ServerStatusWithMaps
            params={Promise.resolve({ locale, server, maps: seasonMaps, status })}
          />
        </div>
      )})}
    </div>
  )
}
