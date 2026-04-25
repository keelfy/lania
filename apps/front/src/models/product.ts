export type Product<T extends ProductMetadata> = {
  id: string
  name: string
  description: string
  price: number
  category: ProductCategory
  soldCount: number
  metadata: T
  createdAt: Date
}

export enum ProductCategory {
  Upgrade = 'upgrade',
  NameColor = 'name-color',
  NamePrefix = 'name-prefix',
}

export type ProductMetadata = object

export type UpgradeProductMetadata = ProductMetadata & {
  action: string
}

export type NameColorProductMetadata = ProductMetadata & {
  colors: string[]
}

export type NamePrefixProductMetadata = ProductMetadata & {
  prefix: string
  namePrefixId: string
}
