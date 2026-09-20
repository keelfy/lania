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
  systemAddress?: string
  rconPort?: number
  rconPasswordSet: boolean
}

export type SaveSeason = {
  name: string
  previewImage?: string
  startDate: string
  endDate?: string
  publicAddress?: string
  systemAddress?: string
  rconPort?: number
  isActive: boolean
  isPrimary: boolean
  preregistration: boolean
  freeRegistration: boolean
  rconPassword?: string
}
