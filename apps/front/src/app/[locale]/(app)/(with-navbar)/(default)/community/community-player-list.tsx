'use client'

import { Button } from '@/components/ui/button'
import { useCursorFeed } from '@/lib/use-cursor-feed'
import { PublicProfile } from '@/models/profile'
import { LoaderCircleIcon } from 'lucide-react'
import { useTranslations } from 'next-intl'
import CommunityPlayerItem from './community-player-item'
import { CommunityFeedQuery, loadMoreProfiles } from './load-more-profiles'

type Props = {
  // The first page, the next ones are fetched by cursor when the end of the list scrolls into view.
  profiles: PublicProfile[]
  nextCursor?: string
  query: CommunityFeedQuery
  locale: string
  // The season the list is shown for, missing for the primary one.
  season?: string
}

export default function CommunityPlayerList({
  profiles,
  nextCursor,
  query,
  locale,
  season,
}: Props) {
  const t = useTranslations('community')
  const { pages, cursor, failed, pending, sentinel, loadMore } = useCursorFeed(
    nextCursor,
    async (requested) => {
      const page = await loadMoreProfiles(query, requested)
      return { items: page.profiles, nextCursor: page.nextCursor }
    },
  )

  // The list can shift between the requests, so a profile may come twice.
  const listed = [
    ...new Map(
      [profiles, ...pages.map((page) => page.items)]
        .flat()
        .map((profile) => [profile.id, profile]),
    ).values(),
  ]

  return (
    <>
      <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
        {listed.map((profile) => (
          <CommunityPlayerItem
            key={profile.id}
            profile={profile}
            locale={locale}
            season={season}
          />
        ))}
      </div>
      {cursor && (
        <div ref={sentinel} className="flex flex-col items-center gap-2">
          {failed && (
            <p role="alert" className="text-destructive text-sm">
              {t('loadMoreFailed')}
            </p>
          )}
          <Button variant="outline" onClick={loadMore} disabled={pending}>
            {pending && <LoaderCircleIcon className="animate-spin" />}
            {t('loadMore')}
          </Button>
        </div>
      )}
    </>
  )
}
