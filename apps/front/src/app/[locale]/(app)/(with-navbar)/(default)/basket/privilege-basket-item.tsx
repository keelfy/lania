'use client'

import { PrivilegeProductMetadata, Product } from '@/models/product'
import { KeyRoundIcon } from 'lucide-react'

type Props = {
  product: Product<PrivilegeProductMetadata>
}

export default function PrivilegeBasketItem({ product }: Props) {
  return (
    <div className="flex items-center gap-2">
      <KeyRoundIcon className="size-10" />
      {product.name}
    </div>
  )
}
