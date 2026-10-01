import { getProfileViolations } from '@/lib/api-endpoints'
import { serverApiFetcher } from '@/lib/server'
import { cn } from '@/lib/utils'
import { ProfileViolations, Violation, ViolationStatus } from '@/models/profile'
import { BanIcon, MicOffIcon } from 'lucide-react'
import { getTranslations } from 'next-intl/server'

// Nothing is known when the request fails, the same as for a season without a running server. loadedAt is the
// moment the term bars are drawn for.
export async function loadProfileViolations(
  profileId: string,
  seasonId?: string,
): Promise<ProfileViolations & { loadedAt: number }> {
  const res = await getProfileViolations(
    serverApiFetcher,
    profileId,
    seasonId,
  ).catch((error) => {
    console.error(error)
    return { available: false, violations: [] }
  })
  return { ...res, loadedAt: Date.now() }
}

const statusColors: Record<ViolationStatus, string> = {
  active: 'text-destructive',
  expired: 'text-muted-foreground',
  removed: 'text-green-600 dark:text-green-500',
}

// The term is cut into blocks, like the playtime bars of the profile.
const BLOCKS_MASK =
  'repeating-linear-gradient(to right, #000 0 6px, transparent 6px 8px)'

// How much of a temporary punishment in force has passed.
function TermBar({ ratio }: { ratio: number }) {
  return (
    <div
      aria-hidden
      className="bg-muted h-1.5 w-full max-w-60"
      style={{ maskImage: BLOCKS_MASK, WebkitMaskImage: BLOCKS_MASK }}
    >
      <div
        className="bg-destructive h-full"
        style={{ clipPath: `inset(0 ${(1 - ratio) * 100}% 0 0)` }}
      />
    </div>
  )
}

type Props = {
  violations: Violation[]
  // The moment the violations were loaded.
  now: number
  locale: string
  className?: string
}

// Bans and mutes of a profile, newest first.
export default async function ProfileViolationList({
  violations,
  now,
  locale,
  className,
}: Props) {
  const t = await getTranslations({ locale, namespace: 'violations' })
  const date = (millis: number) =>
    new Date(millis).toLocaleDateString(locale, {
      day: 'numeric',
      month: 'short',
      year: 'numeric',
      timeZone: 'UTC',
    })

  return (
    <ol className={cn('flex flex-col', className)}>
      {violations.map((violation) => {
        const KindIcon = violation.kind === 'ban' ? BanIcon : MicOffIcon
        const running =
          violation.status === 'active' && violation.expiresAt !== null
        return (
          <li
            key={`${violation.kind}-${violation.id}`}
            className="flex gap-3 py-3 not-last:border-b"
          >
            <KindIcon
              className={cn(
                'mt-0.5 size-4 shrink-0',
                statusColors[violation.status],
              )}
            />
            <div className="flex min-w-0 flex-1 flex-col gap-1">
              <div className="flex items-baseline justify-between gap-4">
                <p className="min-w-0 text-sm">
                  <span className="font-semibold">
                    {t(`kinds.${violation.kind}`)}
                  </span>{' '}
                  <span className="break-words">
                    {violation.reason || (
                      <span className="text-muted-foreground">
                        {t('noReason')}
                      </span>
                    )}
                  </span>
                </p>
                <span
                  className={cn(
                    'shrink-0 text-sm font-semibold',
                    statusColors[violation.status],
                  )}
                >
                  {t(`statuses.${violation.status}`)}
                </span>
              </div>
              <p className="text-muted-foreground flex flex-wrap gap-x-3 text-xs">
                <span>
                  {violation.issuedBy
                    ? t('issuedBy', { name: violation.issuedBy })
                    : t('issuedByServer')}
                </span>
                <time dateTime={new Date(violation.issuedAt).toISOString()}>
                  {date(violation.issuedAt)}
                </time>
                <span>
                  {violation.expiresAt === null
                    ? t('permanent')
                    : t('until', { date: date(violation.expiresAt) })}
                </span>
              </p>
              {running && (
                <TermBar
                  ratio={Math.min(
                    1,
                    Math.max(
                      0,
                      (now - violation.issuedAt) /
                        (violation.expiresAt! - violation.issuedAt),
                    ),
                  )}
                />
              )}
              {violation.status === 'removed' && (
                <p className="text-muted-foreground text-xs">
                  {violation.removedBy
                    ? t('removedBy', { name: violation.removedBy })
                    : t('removedByServer')}
                  {violation.removedAt !== null &&
                    ` ${date(violation.removedAt)}`}
                  {violation.removedReason && `: ${violation.removedReason}`}
                </p>
              )}
            </div>
          </li>
        )
      })}
    </ol>
  )
}
