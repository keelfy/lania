import { Currency } from '@/lib/currency'
import { NamePrefixProductMetadata, Product } from '@/models/product'
import Image from 'next/image'
import React from 'react'
import ProductCard from './product-card'
import McUsername from '@/components/ui/mc-username'

type Props = React.ComponentProps<'div'> & {
  item: Product<NamePrefixProductMetadata>
  currency: Currency
}

export default function NamePrefixProductCard({ item, ...props }: Props) {
  const { prefix } = item.metadata as NamePrefixProductMetadata

  return (
    <ProductCard item={item} {...props}>
      <div className="flex items-center gap-2 py-3">
        <Image src={prefix} alt={prefix} width={32} height={32} unoptimized />
        <McUsername
          username={item.name}
          className="text-center text-xl font-bold"
        />
      </div>
    </ProductCard>
  )
}
