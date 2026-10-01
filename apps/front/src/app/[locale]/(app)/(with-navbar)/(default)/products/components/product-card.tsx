import AddToBasketButton from '@/app/[locale]/(app)/(with-navbar)/(default)/products/components/add-to-basket-btn'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Currency, CURRENCY_SYMBOLS } from '@/lib/currency'
import { cn } from '@/lib/utils'
import { Product, ProductMetadata } from '@/models/product'
import { ArrowUpRightIcon, SparklesIcon } from 'lucide-react'
import { Link } from '@/i18n/navigation'
import { useTranslations } from 'next-intl'
import React from 'react'
import ProductInteraction from './product-interaction'

type Props = React.ComponentProps<'div'> & {
  item: Product<ProductMetadata>
  currency: Currency
  previewOnly?: boolean
  accent?: string
  accentImage?: { src: string; unoptimized?: boolean }
  featured?: boolean
}

export default function ProductCard({
  item,
  children,
  className,
  currency,
  previewOnly = false,
  accent,
  accentImage,
  featured = false,
  ...props
}: React.PropsWithChildren<Props>) {
  const t = useTranslations('products.card')
  // Server Component: Date.now() reflects the current request time, not a
  // memoized render — the purity rule targets client render idempotency.
  const isNew =
    // eslint-disable-next-line react-hooks/purity
    new Date(item.createdAt) > new Date(Date.now() - 1000 * 60 * 60 * 24 * 7)

  return (
    <ProductInteraction
      accent={accent}
      accentImage={accentImage}
      key={item.id}
      className={cn(
        'bg-card group relative flex h-full min-w-44 flex-col justify-between rounded-md border p-4',
        featured &&
          'border-primary/25 bg-primary/5 md:flex-row md:items-center md:gap-6 md:p-6',
        className,
      )}
      {...props}
    >
      <div className="min-w-0 flex-1 space-y-3">
        {isNew && (
          <Badge
            className="absolute -top-2 right-2 z-10 shadow-sm"
            variant="secondary"
          >
            <SparklesIcon className="size-3" />
            {t('new')}
          </Badge>
        )}
        {children}
        <p className="text-muted-foreground text-sm">{item.description}</p>
      </div>
      <div
        className={cn(
          'mt-3 space-y-3',
          featured && 'md:mt-0 md:w-52 md:shrink-0',
        )}
      >
        <p className="text-xl font-bold">
          {item.price} {CURRENCY_SYMBOLS[currency]}
        </p>
        <div className="flex items-center justify-between gap-2">
          <Button
            variant="secondary"
            size="icon"
            title={t('open')}
            aria-label={`${t('open')}: ${item.name}`}
            disabled={previewOnly}
            asChild={!previewOnly}
          >
            {previewOnly ? (
              <>
                <span className="sr-only">{t('open')}</span>
                <ArrowUpRightIcon className="size-4" />
              </>
            ) : (
              <Link href={`/products/${item.category}/${item.id}`}>
                <span className="sr-only">{t('open')}</span>
                <ArrowUpRightIcon className="size-4" />
              </Link>
            )}
          </Button>
          <AddToBasketButton
            productId={item.id}
            profileId={undefined}
            showText
            className="order-first min-w-0 flex-1"
            disabled={previewOnly}
          />
        </div>
      </div>
    </ProductInteraction>
  )
}
