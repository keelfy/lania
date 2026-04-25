'use client'

import { Button } from '@/components/ui/button'
import ProfileUsername from '@/components/ui/profile-username'
import { useBasket } from '@/context/basket'
import { deleteFromBasket } from '@/lib/api-endpoints'
import { clientApiFetcher } from '@/lib/client'
import { Currency, CURRENCY_SYMBOLS } from '@/lib/currency'
import { cn } from '@/lib/utils'
import { BasketItem } from '@/models/basket'
import {
  NameColorProductMetadata,
  NamePrefixProductMetadata,
  Product,
  ProductMetadata,
  UpgradeProductMetadata,
} from '@/models/product'
import { Profile } from '@/models/profile'
import { CheckCheckIcon, Loader2Icon, TrashIcon } from 'lucide-react'
import { useTranslations } from 'next-intl'
import { useRouter } from 'next/navigation'
import React from 'react'
import NameColorBasketItem from './name-color-basket-item'
import NamePrefixBasketItem from './name-prefix-basket-item'
import UpgradeBasketItem from './upgrade-basket-item'

type Props = {
  item: BasketItem
  product: Product<ProductMetadata>
  profile: Profile
  purchased: boolean
  index: number
  currency: Currency
}

export default function BasketItemElement({
  item,
  product,
  profile,
  purchased,
  index,
  currency,
}: Props) {
  const [isDeletingFromBasket, startDeletingFromBasket] = React.useTransition()

  const t = useTranslations('basket')
  const router = useRouter()
  const { removeItem } = useBasket()

  if (!product || !profile) {
    return null
  }

  let itemComponent = null
  switch (product?.category) {
    case 'name-color':
      itemComponent = (
        <NameColorBasketItem
          product={product as Product<NameColorProductMetadata>}
        />
      )
      break
    case 'upgrade':
      itemComponent = (
        <UpgradeBasketItem
          product={product as Product<UpgradeProductMetadata>}
        />
      )
      break
    case 'name-prefix':
      itemComponent = (
        <NamePrefixBasketItem
          product={product as Product<NamePrefixProductMetadata>}
        />
      )
      break
    default:
      itemComponent = null
  }

  const onDeleteFromBasket = (ids?: string[]) =>
    startDeletingFromBasket(async () => {
      try {
        await deleteFromBasket(clientApiFetcher, ids)
        removeItem(item.id)
      } catch (err) {
        console.error(err)
      }
    })

  return (
    <div
      key={item.productId + item.profileId}
      className="bg-card flex flex-col gap-2 rounded-md p-4 text-lg"
    >
      <div className="flex flex-col justify-between gap-4 lg:flex-row">
        <div className="space-y-2">
          <p className="text-muted-foreground mb-1">
            <span className="font-bold tracking-tight">{index}.</span>
            &nbsp;
            {product.description}
          </p>
          <button
            className="hover:bg-accent w-fit cursor-pointer rounded-md px-3 py-2 transition-colors"
            onClick={() => {
              router.push(`/products/${product.category}/${product.id}`)
            }}
          >
            {itemComponent}
          </button>
          {purchased && (
            <div className="hidden items-center gap-2 text-yellow-500 lg:flex">
              <CheckCheckIcon className="size-4" />
              <p className="text-sm">{t('items.purchased')}</p>
            </div>
          )}
        </div>
        <div className="flex flex-col gap-1 lg:hidden">
          <div className="flex flex-nowrap items-center gap-2 text-nowrap">
            <p className="text-muted-foreground text-md">
              {t('items.recipient')}:
            </p>
            <ProfileUsername profile={profile} className="inline-block" />
          </div>
          {purchased && (
            <div className="flex items-center gap-2 text-yellow-500">
              <CheckCheckIcon className="size-4" />
              <p className="text-sm">{t('items.purchased')}</p>
            </div>
          )}
        </div>
        <div className="mt-2 flex w-full flex-col items-end justify-between gap-2 lg:mt-0 lg:w-fit">
          <div className="flex items-center gap-4">
            <p className="text-2xl font-bold text-nowrap">
              <span
                className={cn(
                  purchased && 'text-muted-foreground line-through',
                )}
              >
                {product.price}
              </span>
              {purchased && <span>&nbsp;0</span>}
              &nbsp; {CURRENCY_SYMBOLS[currency]}&nbsp;
              <span className="text-muted-foreground">
                / {t('items.quantity')}
              </span>
            </p>
            <Button
              variant="destructive"
              size="icon"
              onClick={() => onDeleteFromBasket([item.id])}
              disabled={isDeletingFromBasket}
            >
              {isDeletingFromBasket ? (
                <Loader2Icon className="size-4 animate-spin" />
              ) : (
                <TrashIcon className="size-4" />
              )}
              <span className="sr-only">{t('items.delete')}</span>
            </Button>
          </div>
          <div className="mt-4 hidden flex-nowrap items-center gap-2 text-nowrap lg:flex">
            <p className="text-muted-foreground text-md">
              {t('items.recipient')}:
            </p>
            <ProfileUsername profile={profile} className="inline-block" />
          </div>
        </div>
      </div>
    </div>
  )
}
