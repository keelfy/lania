// How a player gets access to the active season. Both flags are inlined at
// build time. Pre-registration wins if both are set.
export type AccessMode = 'preregistration' | 'free' | 'paid'

export const accessMode: AccessMode =
  process.env.NEXT_PUBLIC_PREREGISTRATION === 'true'
    ? 'preregistration'
    : process.env.NEXT_PUBLIC_FREE_REGISTRATION === 'true'
      ? 'free'
      : 'paid'

// Nothing to buy: access is granted right away.
export const isFreeAccess = accessMode !== 'paid'
