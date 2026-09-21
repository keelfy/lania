import { getAccessMode, isFreeAccess } from '@/lib/access-mode'
import {
  getProfileCosmeticOptions,
  getSeasons,
  getUserProfiles,
} from '@/lib/api-endpoints'
import { getCurrentSession } from '@/lib/get-current-session'
import { serverApiFetcher } from '@/lib/server'
import { Profile, ProfileCosmeticOptions } from '@/models/profile'
import { cache } from 'react'

const DEFAULT_COSMETIC_OPTIONS: ProfileCosmeticOptions = {
  name: { colors: [], glythPrefixes: [], specialPrefixes: [] },
}

// The layout and the page both need these, and cache() makes them one request.
export const getProfiles = cache(async (): Promise<Profile[]> => {
  const session = await getCurrentSession()
  return getUserProfiles(serverApiFetcher, session?.identity?.id).catch(
    (error) => {
      console.error(error)
      return []
    },
  )
})

export const getAllSeasons = cache(() =>
  getSeasons(serverApiFetcher).catch(() => []),
)

export async function isFreeAccessSeason() {
  const seasons = await getAllSeasons()
  return isFreeAccess(getAccessMode(seasons.find((season) => season.isPrimary)))
}

// The profile a page of the profiles section shows: the requested one, or the first one.
export async function loadProfilePage(profileIdParam: string | undefined) {
  const [profiles, seasons] = await Promise.all([
    getProfiles(),
    getAllSeasons(),
  ])
  const profileId = profileIdParam ?? profiles[0]?.id
  const selectedProfile = profiles.find((profile) => profile.id === profileId)

  return { seasons, selectedProfile }
}

// The profile with what it wears in the season. The profiles of getProfiles() wear the primary season.
export async function loadProfileInSeason(
  profileId: string,
  seasonId: string,
): Promise<Profile | undefined> {
  const session = await getCurrentSession()
  const profiles = await getUserProfiles(
    serverApiFetcher,
    session?.identity?.id,
    seasonId,
  ).catch((error) => {
    console.error(error)
    return []
  })
  return profiles.find((profile) => profile.id === profileId)
}

export function loadCosmeticOptions(
  profileId: string,
  seasonId?: string,
): Promise<ProfileCosmeticOptions> {
  return getProfileCosmeticOptions(serverApiFetcher, profileId, seasonId).catch(
    (error) => {
      console.error(error)
      return DEFAULT_COSMETIC_OPTIONS
    },
  )
}
