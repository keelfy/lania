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
import { Season, SeasonWorld } from '@/models/season'
import { FlagIcon, MapIcon } from 'lucide-react'
import pinger from 'minecraft-pinger'
import { getTranslations } from 'next-intl/server'
import Image from 'next/image'
import Link from 'next/link'
import type { RawMotdDescription } from '@/lib/motd'
import ClickToCopy from './components/click-to-copy'
import CopyStateIcon from './components/copy-state-icon'
import Motd from './components/motd'
import StopPropagation from './components/stop-propagation'

type Props = {
  locale: string
  server: Season
  worlds: SeasonWorld[]
  status: pinger.Data | undefined
  variant?: 'primary' | 'secondary'
}

export default async function ServerCard({
  locale,
  server,
  worlds,
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
      {worlds.length > 0 && (
        <CardContent className="flex flex-col gap-3">
          <h3 className="text-lg font-semibold">{t('mapsTitle')}</h3>
          <StopPropagation>
            <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
              {worlds.map((world) => (
                <WorldCard
                  key={world.id}
                  locale={locale}
                  world={world}
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

type WorldCardProps = {
  locale: string
  world: SeasonWorld
  compact?: boolean
}

async function WorldCard({ locale, world, compact = false }: WorldCardProps) {
  const t = await getTranslations({ locale, namespace: 'worlds' })
  const claims = !!world.mapUrl && world.claimDimensions.length > 0

  const content = (
    <>
      {world.previewImage && (
        <Image
          src={world.previewImage}
          alt={world.name}
          fill
          sizes="(min-width: 768px) 50vw, 100vw"
          quality={75}
          className="object-cover transition-transform duration-300 group-hover:scale-105"
        />
      )}
      <div className="pointer-events-none absolute inset-0 bg-gradient-to-t from-black/80 via-black/30 to-transparent" />
      <div className="z-10 flex w-full items-start justify-between gap-2">
        <h3
          className={cn(
            'flex items-center gap-2 font-bold',
            compact ? 'text-lg' : 'text-2xl',
          )}
        >
          <MapIcon className="size-5" />
          <span>{world.name}</span>
        </h3>
        {claims && (
          <Badge variant="secondary" className="gap-1">
            <FlagIcon className="size-3" />
            {t('claimsBadge')}
          </Badge>
        )}
      </div>
      <div className="z-10 flex items-center gap-2">
        {world.mapUrl ? (
          <>
            <p className="text-primary">{t('openMap')}</p>
            <p className="font-bold transition-transform duration-300 group-hover:translate-x-1">
              →
            </p>
          </>
        ) : (
          <p className="text-muted-foreground">{t('noMap')}</p>
        )}
      </div>
    </>
  )
  const className = cn(
    'bg-card group relative flex aspect-video w-full flex-col items-start justify-between gap-3 justify-self-center overflow-hidden rounded-md text-start shadow-md',
    compact ? 'p-4' : 'p-6',
  )

  return world.mapUrl ? (
    <Link href={`/worlds/${world.slug}`} className={className}>
      {content}
    </Link>
  ) : (
    <div className={className}>{content}</div>
  )
}
