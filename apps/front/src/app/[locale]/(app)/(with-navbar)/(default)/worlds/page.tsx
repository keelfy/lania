import DeerIcon from '@/components/icons/DeerIcon'
import { HouseIcon, PickaxeIcon, UsersIcon } from 'lucide-react'
import pinger from 'minecraft-pinger'
import { Metadata } from 'next'
import { getTranslations } from 'next-intl/server'
import Image from 'next/image'
import MapButton from './map-button'

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
      url: 'https://lania.gg/worlds',
      title: t('title'),
      description: t('description'),
      siteName: 'Lania Network',
    },
  }
}

const maps = [
  {
    id: 'survival',
    icon: HouseIcon,
    image:
      'https://czx1jtlf2o.ufs.sh/f/0UHiIrRo6i8cJGZy0E5VXE04WIcQBZrqeYRp5GjxiFL6ty1T',
  },
  {
    id: 'farms',
    icon: PickaxeIcon,
    image:
      'https://czx1jtlf2o.ufs.sh/f/0UHiIrRo6i8cxmeBarJjowQqVUy7ZDtGeB9iHgs1vTAchKzW',
  },
]

export default async function WorldPage({ params }: Props) {
  const { locale } = await params
  let status = undefined
  const t = await getTranslations({ locale, namespace: 'worlds' })

  try {
    status = await Promise.race([
      pinger.pingPromise('play.lania.gg', 25565),
      new Promise<void>((_, reject) => setTimeout(reject, 1000)),
    ])
  } catch (error) {
    console.error(error)
  }

  return (
    <div className="flex flex-col gap-4">
      <h1 className="text-4xl font-extrabold tracking-tight">
        {t('status.title')}
      </h1>
      <div className="bg-card flex flex-col items-center justify-between gap-2 rounded-md px-4 py-3 shadow-md sm:flex-row">
        <div className="flex items-center gap-4">
          <DeerIcon className="hidden size-14 rounded-sm bg-black/20 p-1 sm:inline-block" />
          <label className="text-sm sm:text-base">
            {status ? (
              (
                status.description?.extra as unknown as {
                  text: string
                  color: string
                }[]
              )?.map((extra) => (
                <span key={extra.text}>
                  {extra.text.includes('\n') && <br />}
                  <span style={{ color: extra.color }}>{extra.text}</span>
                </span>
              ))
            ) : (
              <span className="text-destructive">{t('status.error')}</span>
            )}
          </label>
        </div>
        <div className="flex items-center gap-2">
          <label className="text-md font-bold sm:text-lg">
            {status?.players.online ?? <>&mdash;</>}&nbsp;/&nbsp;
            {status?.players.max ?? <>&mdash;</>}
          </label>
          <UsersIcon className="text-muted-foreground size-5" />
        </div>
      </div>
      <h1 className="mt-2 text-4xl font-extrabold tracking-tight sm:mt-4">
        {t('mapsTitle')}
      </h1>
      <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
        {maps.map((map) => (
          <MapButton
            key={map.id}
            mapId={map.id}
            className="bg-card group relative flex w-full flex-col items-start justify-between gap-4 justify-self-center overflow-hidden rounded-md p-6 text-start shadow-md"
          >
            <h2 className="z-10 flex w-fit items-center gap-2 rounded-xs bg-black/20 px-2 text-2xl font-bold">
              <map.icon className="size-5" />
              <span>{t('elements.' + map.id + '.title')}</span>
            </h2>
            <p className="z-10 w-fit rounded-xs bg-black/20 px-2 py-1 text-sm">
              {t('elements.' + map.id + '.description')}
            </p>
            <div className="z-10 flex items-center gap-2">
              <p className="text-primary z-10 w-fit rounded-xs bg-black/20 px-2">
                {t('openMap')}
              </p>
              <p className="translate-0 font-bold transition-transform duration-300 group-hover:translate-x-1">
                →
              </p>
            </div>
            <Image
              src={map.image}
              alt={t('elements.' + map.id + '.title')}
              fill
              className="absolute inset-0 transform rounded-md transition-transform duration-300 group-hover:scale-105"
              unoptimized
            />
            <div className="pointer-events-none absolute inset-0 bg-black opacity-30 transition-opacity duration-300 group-hover:opacity-0"></div>
          </MapButton>
        ))}
      </div>
    </div>
  )
}
