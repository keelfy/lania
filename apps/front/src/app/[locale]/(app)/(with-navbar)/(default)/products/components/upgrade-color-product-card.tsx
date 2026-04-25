import { Product, UpgradeProductMetadata } from '@/models/product'
import { UserCheckIcon } from 'lucide-react'
import React from 'react'
import ProductCard from './product-card'
import { Currency } from '@/lib/currency'

type Props = React.ComponentProps<'div'> & {
  item: Product<UpgradeProductMetadata>
  currency: Currency
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
        <p className="font-mono text-2xl font-bold">{item.name}</p>
      </div>
    </ProductCard>
  )
}
