import McUsername from '@/components/ui/mc-username'
import NamePrefixes from '@/components/ui/name-prefixes'
import PlayerFace from '@/components/ui/player-face'
import PlayerProfileLink from '@/components/ui/player-profile-link'
import { getTopPlaytimeProfiles } from '@/lib/api-endpoints'
import { formatPlaytime } from '@/lib/playtime'
import { serverApiFetcher } from '@/lib/server'
import { cn } from '@/lib/utils'
import { PublicProfile } from '@/models/profile'
import { getTranslations } from 'next-intl/server'
import type { CSSProperties } from 'react'
import styles from './top-players.module.css'

type Props = {
  locale: string
  title: string
  // The season the playtime is counted in, missing for the primary one.
  season?: string
  href: (profile: PublicProfile) => string
}

const PODIUM_COLORS = [
  'text-yellow-500', // gold
  'text-zinc-400', // silver
  'text-amber-700', // bronze
]

const PODIUM_ACCENTS = ['#eab308', '#a1a1aa', '#d97706']

export default async function TopPlayers({
  locale,
  title,
  season,
  href,
}: Props) {
  const tPlaytime = await getTranslations({
    locale,
    namespace: 'playerCard.playtime',
  })
  const profiles = await getTopPlaytimeProfiles(
    serverApiFetcher,
    undefined,
    season,
  ).catch((err) => {
    console.error(err)
    return []
  })
  if (profiles.length === 0) return null

  const longestPlaytime = Math.max(
    ...profiles.map((profile) => profile.playtime),
  )
  return (
    <section className="flex min-w-0 flex-col gap-3">
      <h2 className="text-xl font-bold tracking-tight">{title}</h2>
      <ol className="flex snap-x gap-3 overflow-x-auto pb-2 sm:grid sm:grid-cols-3 sm:overflow-visible sm:pb-0 lg:grid-cols-5">
        {profiles.map((profile, index) => {
          const playtime = formatPlaytime(profile.playtime)
          const colors = profile.cosmetics.name.colors?.colors
          const accent = colors?.[0] ?? 'var(--primary)'
          const podiumAccent = PODIUM_ACCENTS[index]
          const ratio =
            longestPlaytime > 0 ? profile.playtime / longestPlaytime : 0
          return (
            <li key={profile.id} className="w-44 shrink-0 snap-start sm:w-auto">
              <PlayerProfileLink
                href={href(profile)}
                accent={podiumAccent ?? accent}
                aria-label={profile.username}
                className={cn(
                  'flex h-full flex-col gap-2 rounded-xl border p-3',
                  podiumAccent && styles.podium,
                )}
                style={
                  podiumAccent
                    ? ({ '--podium-color': podiumAccent } as CSSProperties)
                    : undefined
                }
              >
                {podiumAccent && (
                  <span aria-hidden className={styles.particles}>
                    <span />
                    <span />
                    <span />
                  </span>
                )}
                <div
                  className="flex min-w-0 items-center gap-2"
                  title={profile.username}
                >
                  <span
                    className={cn(
                      'w-4 shrink-0 text-sm font-semibold tabular-nums',
                      PODIUM_COLORS[index] ?? 'text-muted-foreground/60',
                    )}
                  >
                    {index + 1}
                  </span>
                  <div className="flex min-w-0 items-center gap-1">
                    <NamePrefixes
                      cosmetics={profile.cosmetics.name}
                      size={14}
                      gap="0.25rem"
                    />
                    <McUsername
                      username={profile.username}
                      colors={colors}
                      className="min-w-0 truncate text-sm"
                    />
                  </div>
                </div>
                <div className="mt-auto flex flex-col gap-2">
                  <div className="flex items-center gap-2">
                    <PlayerFace
                      player={profile}
                      className={cn(
                        'size-6 rounded-sm',
                        podiumAccent && styles.face,
                      )}
                    />
                    <span className="text-muted-foreground text-sm tabular-nums">
                      {playtime.value.toLocaleString(locale)}{' '}
                      {tPlaytime(playtime.unit)}
                    </span>
                  </div>
                  {longestPlaytime > 0 && (
                    <div
                      aria-hidden
                      className="bg-muted h-1 overflow-hidden rounded-full"
                    >
                      <div
                        className="h-full rounded-full opacity-75"
                        style={{
                          width: `${ratio * 100}%`,
                          background:
                            colors && colors.length > 1
                              ? `linear-gradient(to right, ${colors.join(', ')})`
                              : accent,
                        }}
                      />
                    </div>
                  )}
                </div>
              </PlayerProfileLink>
            </li>
          )
        })}
      </ol>
    </section>
  )
}
