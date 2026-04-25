import AddToBasketButton from '@/app/[locale]/(app)/(with-navbar)/(default)/products/components/add-to-basket-btn'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Currency, CURRENCY_SYMBOLS } from '@/lib/currency'
import { cn } from '@/lib/utils'
import { Product, ProductMetadata } from '@/models/product'
import { ArrowUpRightIcon, SparklesIcon } from 'lucide-react'
import Link from 'next/link'
import React from 'react'

type Props = React.ComponentProps<'div'> & {
  item: Product<ProductMetadata>
  currency: Currency
}

export default function ProductCard({
  item,
  children,
  className,
  currency,
  ...props
}: React.PropsWithChildren<Props>) {
  const isNew =
    new Date(item.createdAt) > new Date(Date.now() - 1000 * 60 * 60 * 24 * 7)

  return (
    <div
      key={item.id}
      className={cn(
        'bg-card group relative flex h-full min-w-44 flex-col justify-between rounded-md border p-4',
        className,
      )}
      {...props}
    >
      <div className="space-y-3">
        {isNew && (
          <Badge
            className="absolute -top-2 -right-2 shadow-sm transition-transform group-hover:-translate-x-1 group-hover:scale-105"
            variant="secondary"
          >
            <SparklesIcon className="size-3 animate-pulse" />
            Новинка
          </Badge>
        )}
        {children}
        <p className="text-muted-foreground text-sm">{item.description}</p>
      </div>
      <div className="mt-3 space-y-3">
        <p className="text-xl font-bold">
          {item.price} {CURRENCY_SYMBOLS[currency]}
        </p>
        <div className="flex items-center justify-between gap-2">
          <Button variant="secondary" size="default" className="flex-1" asChild>
            <Link href={`/products/${item.category}/${item.id}`}>
              Подробнее
              <ArrowUpRightIcon className="size-4" />
            </Link>
          </Button>
          <AddToBasketButton
            productId={item.id}
            profileId={undefined}
            showText={false}
          />
        </div>
      </div>
    </div>
  )
}
