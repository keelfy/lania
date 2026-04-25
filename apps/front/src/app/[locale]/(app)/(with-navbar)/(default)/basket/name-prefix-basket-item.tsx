'use client'

import McUsername from '@/components/ui/mc-username'
import { NamePrefixProductMetadata, Product } from '@/models/product'
import Image from 'next/image'

type Props = {
  product: Product<NamePrefixProductMetadata>
}

export default function NamePrefixBasketItem({ product }: Props) {
  return (
    <div className="flex items-center gap-2 py-3">
      <Image
        src={product.metadata.prefix}
        alt={product.metadata.prefix}
        width={32}
        height={32}
        unoptimized
      />
      <McUsername
        username={product.name}
        className="text-center text-2xl font-bold"
      />
    </div>
  )
}
