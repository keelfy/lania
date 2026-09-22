import DeerIcon from '@/components/icons/DeerIcon'
import { Badge } from '@/components/ui/badge'
import {
  CardAction,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { cn } from '@/lib/utils'
import { Season } from '@/models/season'
import { ServerMapInfo } from '@/models/server'
import { HouseIcon, PickaxeIcon } from 'lucide-react'
import pinger from 'minecraft-pinger'
import { getTranslations } from 'next-intl/server'
import Image from 'next/image'
import Link from 'next/link'
import React from 'react'
import type { RawMotdDescription } from '@/lib/motd'
import ClickToCopy from './components/click-to-copy'
import CopyStateIcon from './components/copy-state-icon'
import Motd from './components/motd'
import StopPropagation from './components/stop-propagation'

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

export default async function ServerCard({
  locale,
  server,
  maps,
  status,
  variant = 'primary',
}: Props) {
  const t = await getTranslations({ locale, namespace: 'worlds' })
  const address = server.publicAddress ?? ''
  const online = !!status?.description
  const isPrimary = variant === 'primary'

  return (
    <ClickToCopy copyText={address} copyLabel={t('copyAddress')}>
      <CardHeader className="flex flex-row items-start gap-4">
        {isPrimary && (
          <DeerIcon className="hidden size-14 shrink-0 rounded-sm bg-black/20 p-1.5 sm:inline-block" />
        )}
        <div className="flex min-w-0 flex-1 flex-col gap-1.5">
          <CardTitle
            className={cn(
              'font-minecraft tracking-mc translate-y-0.5',
              isPrimary ? 'text-2xl' : 'text-lg',
            )}
          >
            {server.name}
          </CardTitle>
          <CardDescription className="text-foreground/80 text-sm">
            {status?.description ? (
              <Motd
                description={
                  status.description as unknown as RawMotdDescription
                }
              />
            ) : (
              <span className="text-destructive">{t('status.error')}</span>
            )}
          </CardDescription>
        </div>
        <CardAction className="flex flex-col items-end gap-1.5">
          <Badge
            variant="outline"
            className={cn(
              'px-2.5 py-1 text-sm font-semibold',
              online && 'border-teal-500/30 bg-teal-500/10 text-teal-300',
            )}
          >
            {status?.players.online ?? '—'}&nbsp;/&nbsp;
            {status?.players.max ?? '—'}
          </Badge>
          <span className="text-muted-foreground flex items-center gap-1.5 font-mono text-xs">
            <CopyStateIcon />
            {address}
          </span>
        </CardAction>
      </CardHeader>
      {maps.length > 0 && online && (
        <CardContent className="flex flex-col gap-3">
          <h3 className="text-lg font-semibold">{t('mapsTitle')}</h3>
          <StopPropagation>
            <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
              {maps.map((map) => (
                <ServerMap
                  key={map.id}
                  locale={locale}
                  map={map}
                  compact={!isPrimary}
                />
              ))}
            </div>
          </StopPropagation>
        </CardContent>
      )}
    </ClickToCopy>
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
