import DeerIcon from '@/components/icons/DeerIcon'
import { Badge } from '@/components/ui/badge'
import {
  Card,
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
import WorldInteraction from './components/world-interaction'

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
    <Card className="overflow-hidden">
      <CardHeader className="flex flex-col items-start gap-4 sm:flex-row sm:flex-wrap">
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
          <CardDescription className="text-foreground/80 min-h-10 text-sm leading-5 [overflow-wrap:anywhere] whitespace-pre-wrap">
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
        <div className="flex items-center gap-3 self-start">
          {server.gameVersion && (
            <span className="text-muted-foreground text-xs">
              Minecraft {server.gameVersion}
            </span>
          )}
          <Badge
            variant="outline"
            className={cn(
              'px-2.5 py-1 text-sm font-semibold tabular-nums',
              online && 'border-teal-500/30 bg-teal-500/10 text-teal-300',
            )}
          >
            {status?.players.online ?? '—'}&nbsp;/&nbsp;
            {status?.players.max ?? '—'}
          </Badge>
        </div>
      </CardHeader>
      <CardContent>
        <ClickToCopy
          copyText={address}
          copyLabel={t('copyAction')}
          copiedLabel={t('addressCopied')}
          errorLabel={t('copyError')}
        >
          <CopyStateIcon />
        </ClickToCopy>
      </CardContent>
      {worlds.length > 0 && (
        <CardContent className="flex flex-col gap-3">
          <h3 className="text-lg font-semibold">{t('mapsTitle')}</h3>
          <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
            {worlds.map((world, index) => (
              <WorldCard
                key={world.id}
                locale={locale}
                world={world}
                compact={!isPrimary}
                featured={isPrimary && index === 0 && worlds.length % 2 === 1}
              />
            ))}
          </div>
        </CardContent>
      )}
    </Card>
  )
}

type WorldCardProps = {
  locale: string
  world: SeasonWorld
  compact?: boolean
  featured?: boolean
}

async function WorldCard({
  locale,
  world,
  compact = false,
  featured = false,
}: WorldCardProps) {
  const t = await getTranslations({ locale, namespace: 'worlds' })
  const claims = !!world.mapUrl && world.claimDimensions.length > 0
  const descriptionKey = `descriptions.${world.slug}`
  const description = t.has(descriptionKey) ? t(descriptionKey) : undefined

  const content = (
    <>
      {world.previewImage && (
        <Image
          src={world.previewImage}
          alt={world.name}
          fill
          sizes={
            featured
              ? '(min-width: 1024px) 976px, 100vw'
              : '(min-width: 1024px) 480px, (min-width: 768px) 50vw, 100vw'
          }
          quality={75}
          className="object-cover"
        />
      )}
      <div className="pointer-events-none absolute inset-0 bg-gradient-to-t from-black/85 via-black/10 to-black/20" />
      <div className="z-10 flex w-full items-start justify-between gap-2">
        <div className="flex min-w-0 flex-col gap-1">
          <h3
            className={cn(
              'flex items-center gap-2 font-bold',
              compact ? 'text-lg' : 'text-2xl',
            )}
          >
            <MapIcon className="size-5" />
            <span>{world.name}</span>
          </h3>
          {description && (
            <p className="text-foreground/80 text-sm leading-5">
              {description}
            </p>
          )}
        </div>
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
            <p className="font-bold transition-transform duration-300 motion-safe:group-hover:translate-x-1 motion-reduce:transition-none">
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
    'relative flex h-full w-full flex-col items-start justify-between gap-3 text-start focus-visible:outline-2 focus-visible:outline-offset-[-4px] focus-visible:outline-teal-300',
    compact ? 'p-4' : 'p-6',
  )

  return (
    <WorldInteraction
      className={cn(
        'aspect-video',
        featured && 'md:col-span-2 md:aspect-[2.5/1]',
      )}
    >
      {world.mapUrl ? (
        <Link href={`/worlds/${world.slug}`} className={className}>
          {content}
        </Link>
      ) : (
        <div className={className}>{content}</div>
      )}
    </WorldInteraction>
  )
}
