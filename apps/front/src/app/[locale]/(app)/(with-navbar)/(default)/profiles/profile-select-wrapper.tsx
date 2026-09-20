'use client'

import ProfileSelect from '@/components/ui/profile-select'
import { Profile } from '@/models/profile'
import { useRouter, useSearchParams } from 'next/navigation'

type Props = {
  profiles: Profile[]
  selectedProfileId: string | undefined
}

export default function ProfileSelectWrapper({
  profiles,
  selectedProfileId,
}: Props) {
  const router = useRouter()
  // The chosen season stays when another profile is picked.
  const seasonId = useSearchParams().get('s')
  return (
    <ProfileSelect
      profiles={profiles}
      className="w-full lg:w-1/3"
      placeholder="Профиль"
      selectedProfileId={selectedProfileId}
      onSelectProfileId={(profileId) => {
        router.push(
          `/profiles?id=${profileId}${seasonId ? `&s=${seasonId}` : ''}`,
        )
      }}
    />
  )
}
