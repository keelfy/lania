export type Account = {
  id: string
  state: string
  created_at: string
  traits: { email: string }
}
export type Profile = {
  id: string
  minecraftUUID: string
  username: string
  ownerId: string | null
}
export type Product = {
  id: string
  name: string
  category: 'upgrade' | 'name-color' | 'name-prefix'
}
export type Grant = {
  id: string
  kind: 'access' | 'color' | 'prefix'
  name: string
  season: string | null
  protected: boolean
}
export type AdminData = {
  actor: string
  search: string
  owner: string
  previous: string
  next: string
  activeSeason: string
  attachProfile: Profile | null
  candidate: Account | null
  accounts: Account[] | null
  account: Account | null
  profiles: Profile[] | null
  profile: Profile | null
  products: Product[] | null
  seasons: { id: string; number: number }[] | null
  grants: Grant[] | null
}
export const categoryLabels = {
  upgrade: 'Доступ к сезону',
  'name-color': 'Цвет имени',
  'name-prefix': 'Префикс',
  access: 'Доступ к сезону',
  color: 'Цвет имени',
  prefix: 'Префикс',
} as const
