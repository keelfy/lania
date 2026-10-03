import { Currency } from '@/lib/currency'
import {
  NameColorProductMetadata,
  NamePrefixProductMetadata,
  Product,
  ProductCategory,
  PrivilegeProductMetadata,
  ProductMetadata,
  UpgradeProductMetadata,
} from '@/models/product'
import UsernameColorProductCard from './name-color-product-card'
import NamePrefixProductCard from './name-prefix-product-card'
import PrivilegeProductCard from './privilege-product-card'
import UpgradeProductCard from './upgrade-color-product-card'

export default function ProductCards({
  items,
  currency,
  previewOnly,
}: {
  items: Product<ProductMetadata>[]
  currency: Currency
  previewOnly: boolean
}) {
  return items.map((item) => {
    switch (item.category) {
      case ProductCategory.NameColor:
        return (
          <UsernameColorProductCard
            key={item.id}
            item={item as Product<NameColorProductMetadata>}
            currency={currency}
            previewOnly={previewOnly}
          />
        )
      case ProductCategory.Upgrade:
        return (
          <UpgradeProductCard
            key={item.id}
            item={item as Product<UpgradeProductMetadata>}
            currency={currency}
            previewOnly={previewOnly}
          />
        )
      case ProductCategory.NamePrefix:
        return (
          <NamePrefixProductCard
            key={item.id}
            item={item as Product<NamePrefixProductMetadata>}
            currency={currency}
            previewOnly={previewOnly}
          />
        )
      case ProductCategory.Privilege:
        return (
          <PrivilegeProductCard
            key={item.id}
            item={item as Product<PrivilegeProductMetadata>}
            currency={currency}
            previewOnly={previewOnly}
          />
        )
      default:
        return null
    }
  })
}
