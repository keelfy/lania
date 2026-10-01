import { Profile } from '@/models/profile'

const CRAFATAR_URL = 'https://crafatar-pub.neodium.fr'

export const STEVE_SKIN_URL = '/images/steve_skin.png'
export const STEVE_FACE_URL = '/images/steve_face.jpg'

export type SkinTexture = {
  url: string
  slim: boolean
}

type SkinOwner = Pick<Profile, 'mojangUuid' | 'skin'> & {
  // The model of the licensed skin, known on profile details only.
  isSlimModel?: boolean
}

// The account whose skin crafatar draws: the one the skin was copied from in game, or the licensed one.
function crafatarAccount(profile: SkinOwner): string | undefined {
  return profile.skin ? profile.skin.mojangUuid : profile.mojangUuid
}

// The skin the player wears in game: the one chosen in game, then the licensed one, then null for Steve.
export function skinTexture(
  profile: SkinOwner | undefined,
): SkinTexture | null {
  if (!profile) return null
  if (profile.skin?.textureUrl) {
    return { url: profile.skin.textureUrl, slim: profile.skin.slim }
  }
  const account = crafatarAccount(profile)
  if (!account) return null
  return {
    url: `${CRAFATAR_URL}/skins/${account}`,
    slim: profile.skin ? profile.skin.slim : (profile.isSlimModel ?? false),
  }
}

// A skin chosen in game hides the cape of the licensed account.
export function capeUrl(profile: SkinOwner | undefined): string {
  return profile?.mojangUuid && !profile.skin
    ? `${CRAFATAR_URL}/capes/${profile.mojangUuid}`
    : ''
}

// The face crafatar renders, or null when the face has to be cut from the texture or the player is Steve.
export function crafatarFaceUrl(profile: SkinOwner | undefined): string | null {
  if (!profile || profile.skin?.textureUrl) return null
  const account = crafatarAccount(profile)
  return account ? `${CRAFATAR_URL}/avatars/${account}?size=64` : null
}

// Whether the player has a face of its own, not Steve.
export function hasOwnFace<T extends SkinOwner>(
  profile: T | undefined,
): profile is T {
  return !!profile?.skin?.textureUrl || !!crafatarFaceUrl(profile)
}

// Sent on window after the skin of a profile changed on the site, with the profile id as the detail, so what shows
// the skin loads it again.
export const SKIN_CHANGED_EVENT = 'lania:skin-changed'
