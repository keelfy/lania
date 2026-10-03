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

// One page of the shop. The counts cover every category and ignore the page, they label the category tabs.
export type ProductCatalog = {
  content: Product<ProductMetadata>[]
  nextCursor?: string
  counts: Record<string, number>
}

export enum ProductCategory {
  Upgrade = 'upgrade',
  NameColor = 'name-color',
  NamePrefix = 'name-prefix',
  Privilege = 'privilege',
}

export type ProductMetadata = object

export type UpgradeProductMetadata = ProductMetadata & {
  action: string
}

export type NameColorProductMetadata = ProductMetadata & {
  colors: string[]
  nameColorId: string
}

export type PrivilegeProductMetadata = ProductMetadata & {
  privilegeId: string
}

export type NamePrefixProductMetadata = ProductMetadata & {
  prefix: string
  noSpace: boolean
  namePrefixId: string
}
