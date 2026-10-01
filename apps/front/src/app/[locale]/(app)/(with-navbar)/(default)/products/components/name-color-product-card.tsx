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
      <div className="flex min-h-[2lh] items-center text-xl leading-tight">
        <h3
          title={item.name}
          className="text-foreground line-clamp-2 font-semibold text-balance"
        >
          {item.name}
        </h3>
      </div>
      <CatalogNamePreview colors={colors} />
    </ProductCard>
  )
}
