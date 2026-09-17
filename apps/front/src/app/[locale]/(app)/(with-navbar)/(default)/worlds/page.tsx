import pinger from 'minecraft-pinger'
import { Metadata } from 'next'
import { getTranslations } from 'next-intl/server'
import { readFile } from 'node:fs/promises'
import path from 'node:path'
import { ServerInfo } from '@/models/server'
import ServerStatusWithMaps from './server-status'

type Props = {
  params: Promise<{
    locale: string
  }>
}

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { locale } = await params
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

  const servers: ServerInfo[] = await readFile(
    path.join(process.cwd(), 'public/data/servers.json'),
    'utf-8',
  )
    .then((raw) => JSON.parse(raw).servers)
    .catch((error) => {
      console.error('Failed to read servers.json', error)
      return []
    })

  const statuses: { server: ServerInfo; status: pinger.Data | undefined }[] = []

  for (const server of servers) {
    const { domain, port } = server

    if (domain && port) {
      let status: pinger.Data | undefined = undefined
      try {
        status = await Promise.race([
          pinger.pingPromise(domain, port),
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
      {statuses.map(({ server, status }) => (
        <div key={server.domain + ':' + server.port} className="flex flex-col gap-4">
          <h1 className="text-4xl font-extrabold">
            {server.name}
          </h1>
          <ServerStatusWithMaps
            key={server.domain + ':' + server.port}
            params={Promise.resolve({ locale, server, status })}
          />
        </div>
      ))}
    </div>
  )
}
