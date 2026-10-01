'use client'

import SeasonSelect from '@/components/ui/season-select'
import { Season } from '@/models/season'
import { useRouter } from 'next/navigation'

type Props = {
  seasons: Season[]
  selectedSeasonId: string | undefined
  profileId: string
}

// Picks the season the cosmetics on the cosmetics page are chosen for.
export default function CosmeticsSeasonSelect({
  seasons,
  selectedSeasonId,
  profileId,
}: Props) {
  const router = useRouter()
  return (
    <SeasonSelect
      seasons={seasons}
      selectedSeasonId={selectedSeasonId}
      onSelectSeasonId={(seasonId) => {
        router.push(`/profiles/cosmetics?id=${profileId}&s=${seasonId}`)
      }}
      className="w-full"
    />
  )
}
