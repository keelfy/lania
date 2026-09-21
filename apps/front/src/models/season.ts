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
  preregistration: boolean
  freeRegistration: boolean
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
}
