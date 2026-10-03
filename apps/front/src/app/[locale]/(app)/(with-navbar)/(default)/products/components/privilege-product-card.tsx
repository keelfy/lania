import { PrivilegeProductMetadata, Product } from '@/models/product'
import { KeyRoundIcon } from 'lucide-react'
import React from 'react'
import ProductCard from './product-card'
import { Currency } from '@/lib/currency'

type Props = React.ComponentProps<'div'> & {
  item: Product<PrivilegeProductMetadata>
  currency: Currency
  previewOnly?: boolean
}

export default function PrivilegeProductCard({ item, ...props }: Props) {
  return (
    <ProductCard item={item} {...props}>
      <div className="flex items-center gap-2">
        <KeyRoundIcon className="size-6 shrink-0" />
        <h3 className="text-foreground text-2xl font-semibold">{item.name}</h3>
      </div>
    </ProductCard>
  )
}
