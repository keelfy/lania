export type Season = {
  id: string
  // The same in every language.
  name: string
  // Missing for a season that is not shown on the seasons page.
  previewImage?: string
  startDate: number
  endDate?: number
  publicAddress?: string
  isActive: boolean
  isPrimary: boolean
  // Whether the site knows who is online in the season: it is running and has a server.
  onlineAvailable: boolean
  preregistration: boolean
  freeRegistration: boolean
  // The client version needed to join, e.g. "1.21.1".
  gameVersion?: string
  // An absolute link to the world archive of a finished season.
  worldUrl?: string
}

export type AdminSeason = Season & {
  // host:port of the shell service that serves the season server.
  shellAddress?: string
  // Plan web interface of the season server.
  planUrl?: string
}

export type SaveSeason = {
  name: string
  previewImage?: string
  startDate: string
  endDate?: string
  publicAddress?: string
  shellAddress?: string
  planUrl?: string
  isActive: boolean
  isPrimary: boolean
  preregistration: boolean
  freeRegistration: boolean
  gameVersion?: string
  worldUrl?: string
}

// One server of the season network: survival, farms, creative. Each has its own map and chunk claims.
export type SeasonWorld = {
  id: string
  seasonId: string
  // The key of the world page, /worlds/<slug>; unique within the season.
  slug: string
  // The same in every language.
  name: string
  // s3://bucket/key, the same form as Season.previewImage.
  previewImage?: string
  // The squaremap of the world server. The world has no map page without it.
  mapUrl?: string
  // Chunks one profile may claim in the world, over its claim dimensions together.
  claimLimit: number
  // squaremap world names where chunks can be claimed; empty for a view-only map.
  claimDimensions: string[]
  position: number
}

export type SaveSeasonWorld = {
  slug: string
  name: string
  previewImage?: string
  mapUrl?: string
  claimLimit: number
  claimDimensions: string[]
  position: number
}

export type ScreenshotAuthor = {
  id: string
  username: string
}

export type SeasonScreenshot = {
  id: string
  // s3://bucket/key, the same form as Season.previewImage.
  image: string
  title?: string
  position: number
  authors: ScreenshotAuthor[]
}

export type SaveSeasonScreenshot = {
  image: string
  title?: string
  position: number
  authorProfileIds: string[]
}
