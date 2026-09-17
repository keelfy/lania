export type ServerInfo = {
  name: string
  domain: string
  port: number
  maps: ServerMapInfo[]
}

export type ServerMapInfo = {
  id: string
  icon: string
  image: string
}