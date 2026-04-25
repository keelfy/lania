import McUsername from '@/components/ui/mc-username'
import { NameColorProductMetadata, Product } from '@/models/product'
import React from 'react'
import ProductCard from './product-card'
import { Currency } from '@/lib/currency'

type Props = React.ComponentProps<'div'> & {
  item: Product<NameColorProductMetadata>
  currency: Currency
}

export default function UsernameColorProductCard({ item, ...props }: Props) {
  const { colors } = item.metadata as NameColorProductMetadata

  return (
    <ProductCard item={item} {...props}>
      <McUsername
        username={item.name}
        colors={colors}
        className="w-fit py-3 text-center text-2xl font-bold"
      />
    </ProductCard>
  )
}
