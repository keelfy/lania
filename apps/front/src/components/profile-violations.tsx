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

const statusStyles: Record<ViolationStatus, string> = {
  active: 'bg-destructive/15 text-destructive',
  expired: 'bg-muted text-muted-foreground',
  removed: 'bg-green-500/15 text-green-600 dark:text-green-500',
}

const iconStyles: Record<ViolationStatus, string> = {
  active: 'bg-destructive/15 text-destructive',
  expired: 'bg-muted text-muted-foreground',
  removed: 'bg-muted text-green-600 dark:text-green-500',
}

const MINUTE = 60_000
const HOUR = 60 * MINUTE
const DAY = 24 * HOUR

// The term in its largest whole unit, rounded to the nearest one.
function termUnit(millis: number): {
  unit: 'days' | 'hours' | 'minutes'
  count: number
} {
  if (millis >= DAY) return { unit: 'days', count: Math.round(millis / DAY) }
  if (millis >= HOUR) return { unit: 'hours', count: Math.round(millis / HOUR) }
  return { unit: 'minutes', count: Math.max(1, Math.round(millis / MINUTE)) }
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

// Bans and mutes of a profile, newest first. The reason leads each row; the kind and the term sit under it, the
// status and the date on the right.
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
        const expiresAt = violation.expiresAt
        let summary: string
        if (expiresAt === null) {
          summary = t('summary.permanent', { kind: violation.kind })
        } else {
          const term = termUnit(expiresAt - violation.issuedAt)
          summary = t('summary.temporary', {
            kind: violation.kind,
            duration: t(`durations.${term.unit}`, { count: term.count }),
          })
        }
        return (
          <li
            key={`${violation.kind}-${violation.id}`}
            className="flex gap-3 py-4 not-last:border-b"
          >
            <span
              className={cn(
                'flex size-8 shrink-0 items-center justify-center rounded-md',
                iconStyles[violation.status],
              )}
            >
              <KindIcon className="size-4" />
            </span>
            <div className="flex min-w-0 flex-1 flex-col gap-2">
              <div className="flex items-start justify-between gap-4">
                <div className="flex min-w-0 flex-col gap-0.5">
                  <p className="text-sm font-medium break-words">
                    {violation.reason || (
                      <span className="text-muted-foreground">
                        {t('noReason')}
                      </span>
                    )}
                  </p>
                  <p className="text-muted-foreground text-xs">
                    {summary},{' '}
                    {violation.issuedBy
                      ? t('issuedBy', { name: violation.issuedBy })
                      : t('issuedByServer')}
                  </p>
                </div>
                <div className="flex shrink-0 flex-col items-end gap-1">
                  <span
                    className={cn(
                      'rounded-sm px-1.5 py-0.5 text-xs font-medium',
                      statusStyles[violation.status],
                    )}
                  >
                    {t(`statuses.${violation.status}`)}
                  </span>
                  <time
                    dateTime={new Date(violation.issuedAt).toISOString()}
                    className="text-muted-foreground text-xs"
                  >
                    {date(violation.issuedAt)}
                  </time>
                </div>
              </div>
              {violation.status === 'active' && expiresAt !== null && (
                <div className="flex items-center gap-3">
                  <TermBar
                    ratio={Math.min(
                      1,
                      Math.max(
                        0,
                        (now - violation.issuedAt) /
                          (expiresAt - violation.issuedAt),
                      ),
                    )}
                  />
                  <span className="text-muted-foreground shrink-0 text-xs">
                    {t('until', { date: date(expiresAt) })}
                  </span>
                </div>
              )}
              {violation.status === 'removed' && (
                <p className="text-muted-foreground border-l-2 pl-2 text-xs">
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
