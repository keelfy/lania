import DeerIcon from '@/components/icons/DeerIcon'
import { cn } from '@/lib/utils'
import { Season } from '@/models/season'
import { ServerMapInfo } from '@/models/server'
import {
  CheckIcon,
  CopyIcon,
  HouseIcon,
  PickaxeIcon,
  UsersIcon,
} from 'lucide-react'
import pinger from 'minecraft-pinger'
import { getTranslations } from 'next-intl/server'
import Image from 'next/image'
import Link from 'next/link'
import React from 'react'
import type { RawMotdDescription } from '@/lib/motd'
import CopyableServerCard from './components/copyable-server-card'
import Motd from './components/motd'

type Props = {
  locale: string
  server: Season
  maps: ServerMapInfo[]
  status: pinger.Data | undefined
}

const ICON_MAP: Record<string, React.ElementType> = {
  house: HouseIcon,
  pickaxe: PickaxeIcon,
}

export default async function ServerStatusWithMaps({
  locale,
  server,
  maps,
  status,
}: Props) {
  const t = await getTranslations({ locale, namespace: 'worlds' })
  const address = server.publicAddress ?? ''

  return (
    <div className="flex h-min flex-col gap-2">
      <CopyableServerCard copyText={address} copyLabel={t('copyAddress')}>
        {(copied) => (
          <>
            <div className="flex items-center gap-4">
              <DeerIcon className="hidden size-14 rounded-sm bg-black/20 p-1 sm:inline-block" />
              <span className="text-sm sm:text-base">
                {status?.description ? (
                  <Motd
                    description={
                      status.description as unknown as RawMotdDescription
                    }
                  />
                ) : (
                  <span className="text-destructive">{t('status.error')}</span>
                )}
              </span>
            </div>
            <div className="flex flex-col items-end gap-1">
              <div className="flex items-center gap-2">
                <span className="text-md font-bold sm:text-lg">
                  {status?.players.online ?? <>&mdash;</>}&nbsp;/&nbsp;
                  {status?.players.max ?? <>&mdash;</>}
                </span>
                <UsersIcon className="text-muted-foreground size-5" />
              </div>
              <span className="text-muted-foreground flex items-center gap-1.5 font-mono text-sm">
                <span className="relative inline-block size-4">
                  <CheckIcon
                    className={cn(
                      'absolute inset-0 transition-all duration-200',
                      copied
                        ? 'scale-100 text-teal-500 opacity-100'
                        : 'scale-75 opacity-0',
                    )}
                  />
                  <CopyIcon
                    className={cn(
                      'absolute inset-0 size-3.5 transition-all duration-200',
                      copied
                        ? 'scale-75 opacity-0'
                        : 'scale-100 opacity-70 group-hover:opacity-100',
                    )}
                  />
                </span>
                {address}
              </span>
            </div>
          </>
        )}
      </CopyableServerCard>
      {maps.length > 0 && !!status?.description && (
        <>
          <h3 className="mt-1 text-2xl font-bold tracking-tight sm:mt-2">
            {t('mapsTitle')}
          </h3>
          <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
            {maps.map((map) => (
              <ServerMap key={map.id} locale={locale} map={map} />
            ))}
          </div>
        </>
      )}
    </div>
  )
}

type ServerMapProps = {
  locale: string
  map: ServerMapInfo
}

async function ServerMap({ locale, map }: ServerMapProps) {
  const t = await getTranslations({ locale, namespace: 'worlds' })

  return (
    <Link
      href={`/worlds/${map.id}`}
      className="bg-card group relative flex aspect-video w-full flex-col items-start justify-between gap-4 justify-self-center overflow-hidden rounded-md p-6 text-start shadow-md"
    >
      <Image
        src={map.image}
        alt={t('elements.' + map.id + '.title')}
        fill
        sizes="(min-width: 768px) 50vw, 100vw"
        quality={75}
        className="object-cover transition-transform duration-300 group-hover:scale-105"
      />
      <div className="pointer-events-none absolute inset-0 bg-gradient-to-t from-black/80 via-black/30 to-transparent" />
      <h3 className="z-10 flex items-center gap-2 text-2xl font-bold">
        {React.createElement(ICON_MAP[map.icon] || HouseIcon, {
          className: 'size-5',
        })}
        <span>{t('elements.' + map.id + '.title')}</span>
      </h3>
      <p className="z-10 text-sm">{t('elements.' + map.id + '.description')}</p>
      <div className="z-10 flex items-center gap-2">
        <p className="text-primary">{t('openMap')}</p>
        <p className="font-bold transition-transform duration-300 group-hover:translate-x-1">
          →
        </p>
      </div>
    </Link>
  )
}
