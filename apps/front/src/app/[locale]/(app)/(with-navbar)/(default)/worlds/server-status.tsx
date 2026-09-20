import DeerIcon from '@/components/icons/DeerIcon';
import { ServerInfo, ServerMapInfo } from '@/models/server';
import { HouseIcon, PickaxeIcon, Server, UsersIcon } from 'lucide-react';
import pinger from 'minecraft-pinger';
import { getTranslations } from 'next-intl/server';
import MapButton from './map-button';
import React from 'react';
import Image from 'next/image';
import SmallCopyIpButton from './components/small-copy-ip-button';
import { Season } from '@/models/season';

type Props = {
  params: Promise<{
    locale: string
    server: Season
    maps: ServerMapInfo[]
    status: pinger.Data | undefined
  }>
}

const ICON_MAP: Record<string, React.ElementType> = {
  house: HouseIcon,
  pickaxe: PickaxeIcon,
}

export default async function ServerStatusWithMaps({ params, }: Props) {
  const { locale, server, maps, status } = await params
  const t = await getTranslations({ locale, namespace: 'worlds' })
  
  return (
    <div className="flex flex-col gap-2 h-min">
      <div className="bg-card flex flex-col items-center justify-between gap-2 rounded-md px-4 py-3 shadow-md sm:flex-row">
        <div className="flex items-center gap-4">
          <DeerIcon className="hidden size-14 rounded-sm bg-black/20 p-1 sm:inline-block" />
          <label className="text-sm sm:text-base">
            {status?.description ? (
              <>
                <span key={status.description.text}>
                    {status.description.text.includes('\n') && <br />}
                    <span style={{ color: status.description as unknown as { color: string } }.color }>
                      {status.description.text}
                    </span>
                  </span>
                {(
                  status.description?.extra as unknown as {
                    text: string
                    color: string
                  }[]
                )?.map((extra) => (
                  <span key={extra.text}>
                    {extra.text.includes('\n') && <br />}
                    <span style={{ color: extra.color }}>{extra.text}</span>
                  </span>
                ))}
              </>
            ) : (
              <span className="text-destructive">{t('status.error')}</span>
            )}
          </label>
        </div>
        <div className="flex flex-col items-end gap-1">
          <div className="flex items-center gap-2">
            <label className="text-md font-bold sm:text-lg">
              {status?.players.online ?? <>&mdash;</>}&nbsp;/&nbsp;
              {status?.players.max ?? <>&mdash;</>}
            </label>
            <UsersIcon className="text-muted-foreground size-5" />
          </div>
          <SmallCopyIpButton copyText={server.publicAddress ?? "127.0.0.1"} />
        </div>
      </div>
      {maps.length > 0 && !!status?.description && (
        <>
          <h2 className="mt-1 text-2xl font-bold tracking-tight sm:mt-2">
            {t('mapsTitle')}
          </h2>
          <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
            {maps.map((map) => (
              <ServerMap
                key={map.id}
                params={Promise.resolve({ locale, server, map, status })}
              />
            ))}
          </div>
        </>
      )}
    </div>
  )
}

type ServerMapProps = {
  params: Promise<{
    locale: string
    server: Season
    map: ServerMapInfo
    status: pinger.Data | undefined
  }>
}

async function ServerMap({ params }: ServerMapProps) {
  const { locale, server, map, status } = await params
  const t = await getTranslations({ locale, namespace: 'worlds' })

  return (
    <MapButton 
      mapId={map.id}
      className="bg-card group relative flex w-full flex-col items-start justify-between gap-4 justify-self-center overflow-hidden rounded-md p-6 text-start shadow-md"
    >
      <h2 className="z-10 flex w-fit items-center gap-2 rounded-xs bg-black/20 px-2 text-2xl font-bold">
        {React.createElement(ICON_MAP[map.icon] || HouseIcon, {
          className: 'size-5',
        })}
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
        width={400}
        height={400}
        quality={75}
        className="absolute w-full inset-0 rounded-md top-1/2 -translate-y-1/2 transform transition-transform duration-300 group-hover:scale-105"
      />
      <div className="pointer-events-none absolute inset-0 bg-black opacity-30 transition-opacity duration-300 group-hover:opacity-0"></div>
    </MapButton>
  )
}
