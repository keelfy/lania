import { Season } from '@/models/season'

// How a player gets access to a season. Pre-registration wins if both flags are set.
export type AccessMode = 'preregistration' | 'free' | 'paid'

export function getAccessMode(
  season?: Pick<Season, 'preregistration' | 'freeRegistration'>,
): AccessMode {
  return season?.preregistration
    ? 'preregistration'
    : season?.freeRegistration
      ? 'free'
      : 'paid'
}

// Nothing to buy: access is granted right away.
export function isFreeAccess(mode: AccessMode) {
  return mode !== 'paid'
}
