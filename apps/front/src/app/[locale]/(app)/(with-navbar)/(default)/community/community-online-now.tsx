import McUsername from '@/components/ui/mc-username'
import PlayerFace from '@/components/ui/player-face'
import ScrollFade from '@/components/ui/scroll-fade'
import { getProfiles } from '@/lib/api-endpoints'
import { serverApiFetcher } from '@/lib/server'
import { getTranslations } from 'next-intl/server'
import Link from 'next/link'
import { communityHref } from './community-href'

type Props = {
  locale: string
}

const ONLINE_NOW_LIMIT = 30

export default async function CommunityOnlineNow({ locale }: Props) {
  const t = await getTranslations({ locale, namespace: 'community.onlineNow' })
  const online = await getProfiles(
    serverApiFetcher,
    undefined,
    undefined,
    0,
    '',
    ONLINE_NOW_LIMIT,
    true,
  ).catch((err) => {
    console.error(err)
    return null
  })
  if (!online || online.content.length === 0) return null

  return (
    <section className="flex flex-col gap-3">
      <div className="flex items-baseline justify-between gap-2">
        <h2 className="text-2xl font-bold tracking-tight">{t('title')}</h2>
        {online.totalElements > online.content.length && (
          <Link
            href={communityHref({ locale, online: true })}
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
                href={`/${locale}/community/${encodeURIComponent(profile.username)}`}
                className="hover:bg-accent flex w-24 flex-col items-center gap-1 rounded-md p-2 transition-colors"
              >
                <PlayerFace player={profile} className="size-10" />
                <McUsername
                  username={profile.username}
                  colors={profile.cosmetics.name.colors.colors}
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
