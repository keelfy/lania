export type Season = {
  id: string
  // The same in every language.
  name: string
  // Missing for a season that is not shown on the seasons page.
  previewImage?: string
  startDate: number
  endDate?: number
  // The season that runs on the game server.
  isActive: boolean
}

export type AdminSeason = Season & {
  seasonNumber: number
  serverIp?: string
  serverPort?: number
  rconPasswordSet: boolean
}

export type SaveSeason = {
  seasonNumber: number
  name: string
  previewImage?: string
  startDate: string
  endDate?: string
  serverIp?: string
  serverPort?: number
  isActive: boolean
  rconPassword?: string
}
