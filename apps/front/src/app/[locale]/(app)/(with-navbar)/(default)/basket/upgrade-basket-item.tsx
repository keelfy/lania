'use client'

import { Product, UpgradeProductMetadata } from '@/models/product'
import { UserCheckIcon } from 'lucide-react'

type Props = {
  product: Product<UpgradeProductMetadata>
}

const ICONS = {
  season_access: UserCheckIcon,
}

export default function UpgradeBasketItem({ product }: Props) {
  const Icon = ICONS[product.metadata.action as keyof typeof ICONS]
  return (
    <div className="flex items-center gap-2">
      {Icon && <Icon className="size-10" />}
      {product.name}
    </div>
  )
}
