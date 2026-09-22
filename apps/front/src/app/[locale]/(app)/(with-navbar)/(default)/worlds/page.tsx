import { getMetadataLocale } from '@/i18n/metadata-locale'
import pinger from 'minecraft-pinger'
import { Metadata } from 'next'
import { unstable_cache } from 'next/cache'
import { getTranslations } from 'next-intl/server'
import { readFile } from 'node:fs/promises'
import path from 'node:path'
import { ServerMapInfo } from '@/models/server'
import ServerStatusWithMaps from './server-status'
import { getSeasons } from '@/lib/api-endpoints'
import { serverApiFetcher } from '@/lib/server'
import { Season } from '@/models/season'

type Props = {
  params: Promise<{
    locale: string
  }>
}

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

const pingServer = unstable_cache(
  async (publicAddress: string): Promise<pinger.Data | undefined> => {
    let timeoutId: ReturnType<typeof setTimeout> | undefined
    try {
      return await Promise.race([
        pinger.pingPromise(publicAddress, 25565),
        new Promise<undefined>((_, reject) => {
          timeoutId = setTimeout(reject, 1000)
        }),
      ])
    } catch (error) {
      console.error(error)
      return undefined
    } finally {
      clearTimeout(timeoutId)
    }
  },
  ['worlds-server-ping'],
  { revalidate: 15 },
)

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

  const statuses = await Promise.all(
    seasons
      .filter((server) => !!server.publicAddress)
      .map(async (server) => ({
        server,
        status: await pingServer(server.publicAddress!),
      })),
  )

  const mapsFor = (server: Season) =>
    (mapsPerServer[server.id] as ServerMapInfo[] | undefined) ?? []

  // The flagship season gets the hero treatment; everything else is
  // secondary and shown as compact status rows below it.
  const primary = statuses.find(({ server }) => server.isPrimary) ?? statuses[0]
  const rest = statuses.filter((entry) => entry !== primary)

  return (
    <div className="flex flex-col gap-10">
      <h1 className="text-2xl font-semibold">{t('title')}</h1>
      {primary && (
        <section className="border-border/60 flex flex-col gap-4 rounded-lg border p-4 sm:p-6">
          <ServerStatusWithMaps
            variant="primary"
            locale={locale}
            server={primary.server}
            maps={mapsFor(primary.server)}
            status={primary.status}
          />
        </section>
      )}
      {rest.length > 0 && (
        <section className="flex flex-col gap-3">
          <h2 className="text-lg font-semibold">{t('otherServers')}</h2>
          <div className="flex flex-col gap-5">
            {rest.map(({ server, status }) => (
              <ServerStatusWithMaps
                key={server.id}
                variant="secondary"
                locale={locale}
                server={server}
                maps={mapsFor(server)}
                status={status}
              />
            ))}
          </div>
        </section>
      )}
    </div>
  )
}
