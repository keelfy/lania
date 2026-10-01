import McUsername from '@/components/ui/mc-username'
import { NameColorProductMetadata, Product } from '@/models/product'
import React from 'react'
import ProductCard from './product-card'
import { CatalogNamePreview } from './catalog-try-on'
import { Currency } from '@/lib/currency'

type Props = React.ComponentProps<'div'> & {
  item: Product<NameColorProductMetadata>
  currency: Currency
  previewOnly?: boolean
}

export default function UsernameColorProductCard({ item, ...props }: Props) {
  const { colors } = item.metadata as NameColorProductMetadata

  return (
    <ProductCard item={item} accent={colors[0]} {...props}>
      {/* Two lines are always reserved so descriptions line up across a row. */}
      <div className="my-1 flex min-h-[2lh] items-center text-xl leading-tight">
        <McUsername
          username={item.name}
          colors={colors}
          title={item.name}
          className="line-clamp-2 font-bold text-balance"
        />
      </div>
      <CatalogNamePreview colors={colors} />
    </ProductCard>
  )
}
