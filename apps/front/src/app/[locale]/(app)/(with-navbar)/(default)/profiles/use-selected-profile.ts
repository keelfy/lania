'use client'

import { Profile } from '@/models/profile'
import { useSearchParams } from 'next/navigation'

// The layout does not get the search params, so the parts of it that need the selected profile read them here.
// Like the pages, they fall back to the first profile.
export function useSelectedProfile(profiles: Profile[]) {
  const id = useSearchParams().get('id') ?? profiles[0]?.id
  return profiles.find((profile) => profile.id === id)
}
