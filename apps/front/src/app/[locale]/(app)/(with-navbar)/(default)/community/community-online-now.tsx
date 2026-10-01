import McUsername from '@/components/ui/mc-username'
import PlayerFace from '@/components/ui/player-face'
import ScrollFade from '@/components/ui/scroll-fade'
import { getProfiles } from '@/lib/api-endpoints'
import { serverApiFetcher } from '@/lib/server'
import { getTranslations } from 'next-intl/server'
import Link from 'next/link'
import { communityHref, communityProfileHref } from './community-href'

type Props = {
  locale: string
  // The season the players are online in, missing for the primary one.
  season?: string
}

const ONLINE_NOW_LIMIT = 30

export default async function CommunityOnlineNow({ locale, season }: Props) {
  const t = await getTranslations({ locale, namespace: 'community.onlineNow' })
  const online = await getProfiles(
    serverApiFetcher,
    undefined,
    undefined,
    0,
    '',
    ONLINE_NOW_LIMIT,
    true,
    false,
    season,
  ).catch((err) => {
    console.error(err)
    return null
  })
  if (!online || online.content.length === 0) return null

  return (
    <section className="bg-card flex min-w-0 flex-col gap-3 rounded-xl border px-4 py-3">
      <div className="flex items-baseline justify-between gap-2">
        <h2 className="text-base font-semibold">{t('title')}</h2>
        {online.totalElements > 0 && (
          <Link
            href={`${communityHref({ locale, online: true, season })}#players`}
            className="text-muted-foreground hover:text-foreground text-sm underline-offset-4 hover:underline"
          >
            {t('showAll')}
          </Link>
        )}
      </div>
      <ScrollFade className="pb-2">
        <ul className="flex w-max gap-3">
          {online.content.map((profile) => (
            <li key={profile.id} className="shrink-0">
              <Link
                href={communityProfileHref(locale, profile.username, season)}
                aria-label={profile.username}
                title={profile.username}
                className="hover:bg-accent focus-visible:ring-ring flex w-24 flex-col items-center gap-2 rounded-lg p-2 transition-colors outline-none focus-visible:ring-2 motion-reduce:transition-none"
              >
                <PlayerFace player={profile} className="size-10 rounded-md" />
                <McUsername
                  username={profile.username}
                  colors={profile.cosmetics.name.colors?.colors}
                  className="w-full truncate text-center text-xs"
                />
              </Link>
            </li>
          ))}
        </ul>
      </ScrollFade>
    </section>
  )
}
