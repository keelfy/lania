import { getAccessMode, isFreeAccess } from '@/lib/access-mode'
import {
  getProfileCosmeticOptions,
  getSeasons,
  getUserProfiles,
} from '@/lib/api-endpoints'
import { getCurrentSession } from '@/lib/get-current-session'
import { serverApiFetcher } from '@/lib/server'
import { Profile, ProfileCosmeticOptions } from '@/models/profile'

export const DEFAULT_COSMETIC_OPTIONS: ProfileCosmeticOptions = {
  name: { colors: [], glythPrefixes: [], specialPrefixes: [] },
}

const fetchProfiles = async (
  profileId: string | undefined,
  userId: string | undefined,
): Promise<[Profile[], ProfileCosmeticOptions]> => {
  if (profileId) {
    return await Promise.all([
      getUserProfiles(serverApiFetcher, userId).catch((error) => {
        console.error(error)
        return []
      }),
      getProfileCosmeticOptions(serverApiFetcher, profileId).catch((error) => {
        console.error(error)
        return DEFAULT_COSMETIC_OPTIONS
      }),
    ])
  }

  const profiles = await getUserProfiles(serverApiFetcher, userId).catch(
    (error) => {
      console.error(error)
      return []
    },
  )

  let cosmeticOptions: ProfileCosmeticOptions = {
    name: DEFAULT_COSMETIC_OPTIONS.name,
  }

  if (profiles.length > 0) {
    cosmeticOptions = await getProfileCosmeticOptions(
      serverApiFetcher,
      profiles[0].id,
    ).catch((error) => {
      console.error(error)
      return DEFAULT_COSMETIC_OPTIONS
    })
  }

  return [profiles, cosmeticOptions]
}

// What every page of the profiles section needs: the profiles of the user and the one that is selected.
export async function loadProfilePage(profileIdParam: string | undefined) {
  const [session, seasons] = await Promise.all([
    getCurrentSession(),
    getSeasons(serverApiFetcher).catch(() => []),
  ])
  const primarySeason = seasons.find((season) => season.isPrimary)
  const freeAccess = isFreeAccess(getAccessMode(primarySeason))

  const [profiles, cosmeticOptions] = await fetchProfiles(
    profileIdParam,
    session?.identity?.id,
  )

  const profileId =
    profileIdParam ?? (profiles.length > 0 ? profiles[0].id : undefined)
  const selectedProfile = profiles.find((profile) => profile.id === profileId)

  return {
    seasons,
    freeAccess,
    profiles,
    cosmeticOptions,
    profileId,
    selectedProfile,
  }
}

export type ProfilePageData = Awaited<ReturnType<typeof loadProfilePage>>
