'use client'

import McUsername from '@/components/ui/mc-username'
import { NameColorProductMetadata, Product } from '@/models/product'

type Props = {
  product: Product<NameColorProductMetadata>
}

export default function NameColorBasketItem({ product }: Props) {
  return (
    <McUsername
      username={product.name}
      colors={product.metadata.colors}
      className="text-4xl font-bold"
    />
  )
}
