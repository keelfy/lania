export type UsernameCheck = {
  status: 'taken' | 'owned_by_you' | 'available'
  hasAccess: boolean
}

export type ProfileRole = 'admin' | 'player' | 'mod' | 'owner'

export type Profile = {
  id: string
  uuid: string
  username: string
  role: ProfileRole
  cosmetics: ProfileCosmetics
  accessStatus: 'active' | 'inactive' | 'expired'
  mojangUuid?: string
}

export type PublicProfile = Profile & {
  role: ProfileRole
  isOnline: boolean
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
  colors: NameColor
  glythPrefix: NamePrefix
  specialPrefix: NamePrefix
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
