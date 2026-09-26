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
      {/* Two lines are always reserved so descriptions line up across a row. */}
      <div className="my-3 flex min-h-[2lh] items-center gap-3 text-xl leading-tight">
        <Image
          src={prefix}
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
    </ProductCard>
  )
}
