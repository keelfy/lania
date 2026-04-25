'use client'

import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import { ScrollArea, ScrollBar } from '@/components/ui/scroll-area'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { getProductByIDs } from '@/lib/api-endpoints'
import { clientApiFetcher } from '@/lib/client'
import { Currency, CURRENCY_SYMBOLS } from '@/lib/currency'
import { cn } from '@/lib/utils'
import { Order } from '@/models/order'
import { Product, ProductMetadata } from '@/models/product'
import { Profile } from '@/models/profile'
import {
  ChevronDownIcon,
  ExternalLinkIcon,
  HandCoinsIcon,
  Loader2Icon,
} from 'lucide-react'
import { useTranslations } from 'next-intl'
import Link from 'next/link'
import React from 'react'

type Props = {
  order: Order
  profiles: Profile[]
  locale: string
  currency: Currency
}

export default function OrderProductsSection({
  order,
  profiles,
  locale,
  currency,
}: Props) {
  const [isOpen, setIsOpen] = React.useState(false)
  const [products, setProducts] = React.useState<Product<ProductMetadata>[]>([])
  const [areProductsLoading, startProductsLoading] = React.useTransition()
  const t = useTranslations()

  React.useEffect(() => {
    if (!isOpen) return
    startProductsLoading(async () => {
      const uniqueProductIds = [
        ...new Set(order.items.map((item) => item.productId)),
      ]
      await getProductByIDs(
        clientApiFetcher,
        uniqueProductIds,
        locale,
        currency,
      ).then(setProducts)
    })
  }, [isOpen, locale, currency, order.items])

  return (
    <div className="flex flex-col gap-2">
      <div className="flex flex-wrap items-center gap-2">
        <HandCoinsIcon className="size-4" />
        <Label className="text-md" htmlFor="items-count">
          Кол-во товаров:
          <span className="font-bold">{order.items.length}</span>шт.
        </Label>
        <Button
          variant="ghost"
          size="sm"
          className="text-muted-foreground h-6"
          onClick={() => setIsOpen(!isOpen)}
        >
          [Показать список товаров]
          <ChevronDownIcon
            className={cn(
              'size-4 rotate-90 transition-transform duration-200',
              isOpen && 'rotate-0',
            )}
          />
        </Button>
        {areProductsLoading && (
          <Loader2Icon className="text-muted-foreground size-4 animate-spin" />
        )}
      </div>
      {isOpen && products.length > 0 && (
        <ScrollArea className="w-[19rem] rounded-md border p-4 whitespace-nowrap md:w-full">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Название</TableHead>
                <TableHead>Категория</TableHead>
                <TableHead>Получатель</TableHead>
                <TableHead>Цена</TableHead>
                <TableHead>Количество</TableHead>
                <TableHead>Ссылка</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {order.items.map((orderItem) => {
                const product = products.find(
                  (item) => item.id === orderItem.productId,
                )
                const amount = orderItem.amounts?.find(
                  (amount) => amount.currency === currency,
                )?.amount
                const profile = profiles.find(
                  (profile) => profile.id === orderItem.profileId,
                )
                return (
                  <TableRow key={orderItem.id}>
                    <TableCell>{product?.name}</TableCell>
                    <TableCell>
                      {t(`products.categories.${product?.category}`)}
                    </TableCell>
                    <TableCell>{profile?.username}</TableCell>
                    <TableCell>
                      {amount} {CURRENCY_SYMBOLS[currency]}
                    </TableCell>
                    <TableCell>{orderItem.quantity}</TableCell>
                    <TableCell>
                      <Link
                        href={`/${locale}/products/${product?.category}/${product?.id}`}
                        target="_blank"
                        className="text-muted-foreground hover:text-primary flex flex-nowrap items-center gap-1 transition-colors"
                      >
                        <ExternalLinkIcon className="size-4" />
                        {product?.name}
                      </Link>
                    </TableCell>
                  </TableRow>
                )
              })}
            </TableBody>
          </Table>
          <ScrollBar orientation="horizontal" />
        </ScrollArea>
      )}
    </div>
  )
}
