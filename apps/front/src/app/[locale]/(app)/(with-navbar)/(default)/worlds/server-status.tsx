import DeerIcon from '@/components/icons/DeerIcon'
import { cn } from '@/lib/utils'
import { Season } from '@/models/season'
import { ServerMapInfo } from '@/models/server'
import { HouseIcon, PickaxeIcon, UsersIcon } from 'lucide-react'
import pinger from 'minecraft-pinger'
import { getTranslations } from 'next-intl/server'
import Image from 'next/image'
import Link from 'next/link'
import React from 'react'
import type { RawMotdDescription } from '@/lib/motd'
import CopyableServerCard from './components/copyable-server-card'
import CopyStateIcon from './components/copy-state-icon'
import Motd from './components/motd'

type Props = {
  locale: string
  server: Season
  maps: ServerMapInfo[]
  status: pinger.Data | undefined
  variant?: 'primary' | 'secondary'
}

const ICON_MAP: Record<string, React.ElementType> = {
  house: HouseIcon,
  pickaxe: PickaxeIcon,
}

// A small dot standing in for online/offline, colored with the brand
// accent instead of a generic traffic-light green.
function StatusDot({ online }: { online: boolean }) {
  return (
    <span
      aria-hidden
      className={cn(
        'size-2 shrink-0 rounded-full',
        online ? 'bg-teal-400' : 'bg-muted-foreground/40',
      )}
    />
  )
}

export default async function ServerStatusWithMaps({
  locale,
  server,
  maps,
  status,
  variant = 'primary',
}: Props) {
  const t = await getTranslations({ locale, namespace: 'worlds' })
  const address = server.publicAddress ?? ''
  const online = !!status?.description

  if (variant === 'secondary') {
    return (
      <div className="flex flex-col gap-3">
        <CopyableServerCard
          copyText={address}
          copyLabel={t('copyAddress')}
          className="px-4 py-2.5"
        >
          <div className="flex items-center gap-3">
            <StatusDot online={online} />
            <span className="font-minecraft tracking-mc text-lg">
              {server.name}
            </span>
          </div>
          <div className="flex items-center gap-4">
            <span className="text-muted-foreground text-sm">
              {status?.players.online ?? <>&mdash;</>}&nbsp;/&nbsp;
              {status?.players.max ?? <>&mdash;</>}
            </span>
            <span className="text-muted-foreground flex items-center gap-1.5 font-mono text-xs">
              <CopyStateIcon />
              {address}
            </span>
          </div>
        </CopyableServerCard>
        {maps.length > 0 && online && (
          <div className="grid grid-cols-1 gap-3 md:grid-cols-2">
            {maps.map((map) => (
              <ServerMap key={map.id} locale={locale} map={map} compact />
            ))}
          </div>
        )}
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-4">
      <h2 className="font-minecraft tracking-mc text-3xl drop-shadow-[0_1.2px_1.2px_rgba(255,255,255,0.2)] sm:text-4xl">
        {server.name}
      </h2>
      <CopyableServerCard copyText={address} copyLabel={t('copyAddress')}>
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
            <StatusDot online={online} />
            <span className="text-md font-bold sm:text-lg">
              {status?.players.online ?? <>&mdash;</>}&nbsp;/&nbsp;
              {status?.players.max ?? <>&mdash;</>}
            </span>
            <UsersIcon className="text-muted-foreground size-5" />
          </div>
          <span className="text-muted-foreground flex items-center gap-1.5 font-mono text-sm">
            <CopyStateIcon />
            {address}
          </span>
        </div>
      </CopyableServerCard>
      {maps.length > 0 && online && (
        <>
          <h3 className="text-lg font-semibold">{t('mapsTitle')}</h3>
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
  compact?: boolean
}

async function ServerMap({ locale, map, compact = false }: ServerMapProps) {
  const t = await getTranslations({ locale, namespace: 'worlds' })

  return (
    <Link
      href={`/worlds/${map.id}`}
      className={cn(
        'bg-card group relative flex aspect-video w-full flex-col items-start justify-between gap-3 justify-self-center overflow-hidden rounded-md text-start shadow-md',
        compact ? 'p-4' : 'p-6',
      )}
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
      <h3
        className={cn(
          'z-10 flex items-center gap-2 font-bold',
          compact ? 'text-lg' : 'text-2xl',
        )}
      >
        {React.createElement(ICON_MAP[map.icon] || HouseIcon, {
          className: 'size-5',
        })}
        <span>{t('elements.' + map.id + '.title')}</span>
      </h3>
      {!compact && (
        <p className="z-10 text-sm">
          {t('elements.' + map.id + '.description')}
        </p>
      )}
      <div className="z-10 flex items-center gap-2">
        <p className="text-primary">{t('openMap')}</p>
        <p className="font-bold transition-transform duration-300 group-hover:translate-x-1">
          →
        </p>
      </div>
    </Link>
  )
}
