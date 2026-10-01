import { Currency } from '@/lib/currency'
import { NamePrefixProductMetadata, Product } from '@/models/product'
import Image from 'next/image'
import React from 'react'
import ProductCard from './product-card'
import { CatalogNamePreview } from './catalog-try-on'
import McUsername from '@/components/ui/mc-username'

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
      <div className="my-1 flex min-h-[2lh] items-center gap-3 text-xl leading-tight">
        <Image
          src={prefix}
          unoptimized={previewOnly}
          alt=""
          width={32}
          height={32}
          className="size-8 shrink-0"
        />
        <McUsername
          username={item.name}
          title={item.name}
          className="line-clamp-2 font-bold text-balance"
        />
      </div>
      <CatalogNamePreview prefix={prefix} unoptimized={previewOnly} />
    </ProductCard>
  )
}
