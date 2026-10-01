import { Currency } from '@/lib/currency'
import { NamePrefixProductMetadata, Product } from '@/models/product'
import React from 'react'
import ProductCard from './product-card'
import { CatalogNamePreview } from './catalog-try-on'

type Props = React.ComponentProps<'div'> & {
  item: Product<NamePrefixProductMetadata>
  currency: Currency
  previewOnly?: boolean
}

export default function NamePrefixProductCard({
  item,
  previewOnly,
  ...props
}: Props) {
  const { prefix } = item.metadata as NamePrefixProductMetadata

  return (
    <ProductCard item={item} previewOnly={previewOnly} {...props}>
      {/* Two lines are always reserved so descriptions line up across a row. */}
      <div className="flex min-h-[2lh] items-center text-xl leading-tight">
        <h3
          title={item.name}
          className="text-foreground line-clamp-2 font-semibold text-balance"
        >
          {item.name}
        </h3>
      </div>
      <CatalogNamePreview prefix={prefix} unoptimized={previewOnly} />
    </ProductCard>
  )
}
