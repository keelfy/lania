'use client'

import SeasonSelect from '@/components/ui/season-select'
import { Season } from '@/models/season'
import { useRouter } from 'next/navigation'

type Props = {
  seasons: Season[]
  selectedSeasonId: string | undefined
  profileId: string
}

export default function AccessSeasonSelect({
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
        router.push(`/profiles?id=${profileId}&s=${seasonId}`)
      }}
      className="w-full"
    />
  )
}
