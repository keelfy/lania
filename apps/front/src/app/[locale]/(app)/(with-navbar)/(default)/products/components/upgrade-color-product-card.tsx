import { Product, UpgradeProductMetadata } from '@/models/product'
import { UserCheckIcon } from 'lucide-react'
import React from 'react'
import ProductCard from './product-card'
import { Currency } from '@/lib/currency'

type Props = React.ComponentProps<'div'> & {
  item: Product<UpgradeProductMetadata>
  currency: Currency
  previewOnly?: boolean
  featured?: boolean
}

const ICONS = {
  season_access: UserCheckIcon,
}

export default function UpgradeProductCard({ item, ...props }: Props) {
  const Icon = ICONS[item.metadata.action as keyof typeof ICONS]
  return (
    <ProductCard item={item} {...props}>
      <div className="flex items-center gap-2">
        {Icon && <Icon className="size-6" />}
        <h3 className="text-foreground text-2xl font-semibold">{item.name}</h3>
      </div>
    </ProductCard>
  )
}
