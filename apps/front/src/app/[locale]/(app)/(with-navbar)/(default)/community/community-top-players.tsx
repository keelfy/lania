import McUsername from '@/components/ui/mc-username'
import NamePrefixes from '@/components/ui/name-prefixes'
import PlayerFace from '@/components/ui/player-face'
import { getTopPlaytimeProfiles } from '@/lib/api-endpoints'
import { formatPlaytime } from '@/lib/playtime'
import { serverApiFetcher } from '@/lib/server'
import { cn } from '@/lib/utils'
import { getTranslations } from 'next-intl/server'
import Link from 'next/link'

type Props = {
  locale: string
}

// On phones the list is one column, so only the first places are shown to keep it short.
const PHONE_VISIBLE_PLACES = 5

const PODIUM_COLORS = [
  'text-yellow-500', // gold
  'text-zinc-400', // silver
  'text-amber-700', // bronze
]

export default async function CommunityTopPlayers({ locale }: Props) {
  const t = await getTranslations({ locale, namespace: 'community.top' })
  const tPlaytime = await getTranslations({
    locale,
    namespace: 'playerCard.playtime',
  })
  const profiles = await getTopPlaytimeProfiles(serverApiFetcher).catch(
    (err) => {
      console.error(err)
      return []
    },
  )
  if (profiles.length === 0) return null

  return (
    <section className="flex flex-col gap-3">
      <h2 className="text-2xl font-bold tracking-tight">{t('title')}</h2>
      <ol className="grid grid-cols-1 gap-2 sm:grid-cols-2 lg:grid-cols-5">
        {profiles.map((profile, index) => {
          const playtime = formatPlaytime(profile.playtime)
          return (
            <li
              key={profile.id}
              className={cn(index >= PHONE_VISIBLE_PLACES && 'hidden sm:block')}
            >
              <Link
                href={`/${locale}/community/${encodeURIComponent(profile.username)}`}
                className="hover:bg-accent flex items-center gap-3 rounded-md border p-3 transition-colors"
              >
                <span
                  className={cn(
                    'w-5 text-center text-lg font-bold',
                    PODIUM_COLORS[index] ?? 'text-muted-foreground',
                  )}
                >
                  {index + 1}
                </span>
                <PlayerFace player={profile} className="size-8" />
                <div className="flex min-w-0 flex-col">
                  <div className="flex items-center gap-1">
                    <NamePrefixes
                      cosmetics={profile.cosmetics.name}
                      size={16}
                    />
                    <McUsername
                      username={profile.username}
                      colors={profile.cosmetics.name.colors.colors}
                      className="truncate text-base"
                    />
                  </div>
                  <span className="text-muted-foreground text-sm">
                    {playtime.value} {tPlaytime(playtime.unit)}
                  </span>
                </div>
              </Link>
            </li>
          )
        })}
      </ol>
    </section>
  )
}
