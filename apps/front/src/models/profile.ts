export type UsernameCheck = {
  status: 'taken' | 'owned_by_you' | 'available'
  hasAccess: boolean
}

export type ProfileRole = 'admin' | 'player' | 'mod' | 'owner'

export type SeasonAccess = {
  seasonId: string
  status: 'active' | 'inactive' | 'expired'
}

export type Profile = {
  id: string
  username: string
  role: ProfileRole
  cosmetics: ProfileCosmetics
  // The status for the primary season.
  accessStatus: 'active' | 'inactive' | 'expired'
  // The status for every running season and every ended season the profile has access to.
  accesses: SeasonAccess[]
  // The UUID the player has in game; for a licensed nickname it equals mojangUuid.
  mcUuid: string
  mojangUuid?: string
  // The owner proved in game that the licensed account is theirs.
  verified?: boolean
}

export type PublicProfile = Profile & {
  role: ProfileRole
  isOnline: boolean
  playtime: number
  lastSeenAt?: number
}

export type ProfilesStats = {
  total: number
  // Missing when the game server cannot be reached.
  online?: number
  newLastWeek: number
}

export type ProfileDetails = Profile & {
  isSlimModel: boolean
  status: 'online' | 'offline' | 'banned'
  lastSeenAt?: number
  firstSeenAt?: number
  role: ProfileRole
  playtime: number
  isOnline: boolean
}

// What one profile did in one season. Playtime is in milliseconds.
export type ProfileSeasonStats = {
  seasonId: string
  seasonName: string
  startDate: number
  // Missing while the season is running.
  endDate?: number
  isActive: boolean
  isPrimary: boolean
  playtime: number
}

export type ProfileStats = {
  // Summed over all seasons, in milliseconds.
  totalPlaytime: number
  // Only seasons the profile played in, the newest first.
  seasons: ProfileSeasonStats[]
}

export type NameColor = {
  id: string
  name: string
  colors: string[]
}

export type NamePrefix = {
  id: string
  name: string
  prefix: string
  image: string
}

export type NameCosmetics = {
  // Missing when the API could not read the cosmetics of the profile.
  colors?: NameColor
  glythPrefix?: NamePrefix
  specialPrefix?: NamePrefix
}

export type ProfileCosmetics = {
  name: NameCosmetics
}

export type ProfileNameColorOption = {
  id: string
  nameColorId: string
  name: string
  profileId: string
  colors: string[]
  forSeasonId?: string // undefined for all seasons
}

export type ProfileNamePrefixOption = {
  id: string
  namePrefixId: string
  name: string
  profileId: string
  prefix: string
  image: string
  forSeasonId?: string // undefined for all seasons
}

export type ProfileNameCosmeticOptions = {
  colors: ProfileNameColorOption[]
  glythPrefixes: ProfileNamePrefixOption[]
  specialPrefixes: ProfileNamePrefixOption[]
}

export type ProfileCosmeticOptions = {
  name: ProfileNameCosmeticOptions
}

export type SelectCosmeticOptionReq = {
  optionId: string | undefined
}

export type ProfileResyncPart = 'role' | 'cosmetics' | 'access'

export type PartResync = {
  part: ProfileResyncPart
  ok: boolean
  // Missing when the part was written.
  error?: string
}

export type SeasonResync = {
  seasonId: string
  seasonName: string
  ok: boolean
  parts: PartResync[]
}

// What a resync wrote to the server of every active season. The request succeeds even when some servers fail.
export type ProfileResync = {
  ok: boolean
  seasons: SeasonResync[]
}

// A verification request waits for the licensed player to join until expiresAt.
export type ProfileVerification = {
  expiresAt: number
}
