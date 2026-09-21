'use client'

import SeasonSelect from '@/components/ui/season-select'
import { Season } from '@/models/season'
import { useRouter } from 'next/navigation'
import React from 'react'
import { CommunityFilters, communityHref } from './community-href'

type Props = CommunityFilters & {
  seasons: Season[]
  // The season the page is shown for.
  selectedSeasonId: string | undefined
  sort: string
  search: string
  locale: string
}

// Picks the season that decides what the page shows: cosmetics, last seen dates and who is online.
export default function SelectCommunitySeason({
  seasons,
  selectedSeasonId,
  sort,
  search,
  locale,
  online,
  staff,
}: Props) {
  const router = useRouter()
  const [isPending, startTransition] = React.useTransition()

  const onSelectSeasonId = (seasonId: string) => {
    const season = seasons.find((season) => season.id === seasonId)
    startTransition(() => {
      router.push(
        communityHref({
          locale,
          sort,
          search,
          // Nobody is online in a season without a running server, so the filter would leave an empty list.
          online: online && season?.onlineAvailable,
          staff,
          season: season?.isPrimary ? undefined : seasonId,
        }),
      )
    })
  }

  return (
    <SeasonSelect
      seasons={seasons}
      selectedSeasonId={selectedSeasonId}
      onSelectSeasonId={onSelectSeasonId}
      disabled={isPending}
      className="w-auto min-w-40"
    />
  )
}
