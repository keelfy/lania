import { getUserProfiles } from '@/lib/api-endpoints'
import { getCurrentSession } from '@/lib/get-current-session'
import { serverApiFetcher } from '@/lib/server'
import { Profile } from '@/models/profile'
import { cache } from 'react'

// The profiles of the signed-in user. The navbar and the profiles section both need them, and cache() makes them one request.
export const getCurrentUserProfiles = cache(async (): Promise<Profile[]> => {
  const session = await getCurrentSession()
  if (!session?.active) return []
  return getUserProfiles(serverApiFetcher, session.identity?.id).catch(
    (error) => {
      console.error(error)
      return []
    },
  )
})

// A licensed profile keyed to its Mojang UUID that is not verified yet: the owner can verify it now, and the site
// points them to it.
export function awaitsVerification(profile: Profile) {
  return (
    !profile.verified &&
    !!profile.mojangUuid &&
    profile.mojangUuid === profile.mcUuid
  )
}
