import { getMetadataLocale } from '@/i18n/metadata-locale'
import { Metadata } from 'next'
import { getTranslations } from 'next-intl/server'
import { pingServer } from './ping-server'
import ServerCard from './server-card'
import { ActiveSeasonWorlds, getActiveSeasonWorlds } from '@/lib/worlds'

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

export default async function WorldPage({ params }: Props) {
  const { locale } = await params
  const t = await getTranslations({ locale, namespace: 'worlds' })

  const seasons = await getActiveSeasonWorlds().catch(
    () => [] as ActiveSeasonWorlds[],
  )

  const statuses = await Promise.all(
    seasons
      .filter(({ season }) => !!season.publicAddress)
      .map(async ({ season, worlds }) => ({
        server: season,
        worlds,
        status: await pingServer(season.publicAddress!),
      })),
  )

  // The flagship season gets the hero treatment; everything else is
  // secondary and shown as compact status rows below it.
  const primary = statuses.find(({ server }) => server.isPrimary) ?? statuses[0]
  const rest = statuses.filter((entry) => entry !== primary)

  return (
    <div className="flex flex-col gap-4">
      <h1 className="text-4xl font-extrabold tracking-tight">{t('title')}</h1>
      {primary && (
        <ServerCard
          variant="primary"
          locale={locale}
          server={primary.server}
          worlds={primary.worlds}
          status={primary.status}
        />
      )}
      {rest.length > 0 && (
        <section className="flex flex-col gap-4">
          <h2 className="text-2xl font-bold tracking-tight">
            {t('otherServers')}
          </h2>
          <div className="flex flex-col gap-4">
            {rest.map(({ server, worlds, status }) => (
              <ServerCard
                key={server.id}
                variant="secondary"
                locale={locale}
                server={server}
                worlds={worlds}
                status={status}
              />
            ))}
          </div>
        </section>
      )}
    </div>
  )
}
