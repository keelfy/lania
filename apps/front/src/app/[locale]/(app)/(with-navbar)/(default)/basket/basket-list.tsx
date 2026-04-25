'use client'

import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import LoadingSpinner from '@/components/ui/loading-spinner'
import RichText from '@/components/ui/rich-text'
import { useBasket } from '@/context/basket'
import { createOrder, deleteFromBasket } from '@/lib/api-endpoints'
import { clientApiFetcher } from '@/lib/client'
import { CURRENCY_SYMBOLS, Currency } from '@/lib/currency'
import { errorToast } from '@/lib/toasts'
import { PaymentMethod, PurchasedProduct } from '@/models/order'
import { Product, ProductMetadata } from '@/models/product'
import { Profile } from '@/models/profile'
import {
  ArrowRightIcon,
  BanknoteIcon,
  BitcoinIcon,
  BrushCleaningIcon,
  ClockIcon,
  HandCoinsIcon,
  Loader2Icon,
} from 'lucide-react'
import { useTranslations } from 'next-intl'
import Link from 'next/link'
import React from 'react'
import BasketItemElement from './basket-item'

type Props = {
  profiles: Profile[]
  purchases: PurchasedProduct[]
  products: Product<ProductMetadata>[]
  currency: Currency
}

export default function BasketList({
  profiles,
  purchases,
  products,
  currency,
}: Props) {
  const t = useTranslations('basket')
  const { clearItems, items } = useBasket()

  const [isCreatingOrder, startCreatingOrder] = React.useTransition()
  const [isDeletingFromBasket, startDeletingFromBasket] = React.useTransition()

  const [paymentMethod, setPaymentMethod] = React.useState<PaymentMethod>(
    PaymentMethod.EasyDonate,
  )

  const ignoredItems = React.useMemo(() => {
    return items.filter((item) => {
      const product = products.find((product) => product.id === item.productId)
      const purchased = purchases.find(
        (purchase) =>
          purchase.productId === product?.id &&
          purchase.profileId === item.profileId,
      )
      return !product || purchased
    })
  }, [items, products, purchases])

  const sum = React.useMemo(() => {
    return items
      .filter((item) => !ignoredItems.some((i) => i.id == item.id))
      .reduce((acc, item) => {
        const product = products.find(
          (product) => product.id === item.productId,
        )
        return acc + (product?.price ?? 0)
      }, 0)
  }, [products, items, ignoredItems])

  const onCreateOrder = React.useCallback(() => {
    startCreatingOrder(async () => {
      await createOrder(clientApiFetcher, {
        paymentMethod,
        products: items
          .filter((item) => !ignoredItems.some((i) => i.id == item.id))
          .map((item) => ({
            id: item.productId,
            profileId: item.profileId,
          })),
      })
        .then((res) => {
          window.location.href = res.paymentUrl
        })
        .catch((err) => {
          errorToast(t('error'), err)
        })
    })
  }, [items, paymentMethod, ignoredItems, t])

  const onBasketClear = () => {
    startDeletingFromBasket(async () => {
      try {
        await deleteFromBasket(clientApiFetcher)
        clearItems()
      } catch (err) {
        console.error(err)
      }
    })
  }

  return (
    <div className="flex flex-col gap-10 lg:flex-row">
      <div className="flex flex-1 flex-col gap-4">
        {items.length === 0 && (
          <div className="bg-card rounded-md p-4 py-8 text-center text-lg">
            <RichText>
              {(tags) =>
                t.rich('empty', {
                  ...tags,
                  store: (chunks: React.ReactNode) => (
                    <Link
                      href="/products"
                      className="font-bold text-blue-500 underline-offset-4 hover:underline"
                    >
                      {chunks}
                    </Link>
                  ),
                })
              }
            </RichText>
          </div>
        )}
        {items.map((item, index) => {
          const product = products.find(
            (product) => product.id === item.productId,
          )
          const profile = profiles.find(
            (profile) => profile.id === item.profileId,
          )
          const purchased = purchases.find(
            (purchase) =>
              purchase.productId === item.productId &&
              purchase.profileId === item.profileId,
          )

          if (!product || !profile) return null

          return (
            <BasketItemElement
              key={item.id}
              item={item}
              product={product}
              profile={profile}
              purchased={purchased !== undefined}
              index={index + 1}
              currency={currency}
            />
          )
        })}
        <Button
          variant="destructive"
          disabled={items.length === 0 || isDeletingFromBasket}
          size="lg"
          onClick={onBasketClear}
          className="mt-4 w-fit self-end"
        >
          {isDeletingFromBasket ? (
            <LoadingSpinner />
          ) : (
            <BrushCleaningIcon className="size-4" />
          )}
          <span>{t('clearBasket')}</span>
        </Button>
      </div>
      <div className="flex w-full flex-col gap-6 lg:max-w-xs">
        <Card>
          <CardHeader>
            <CardTitle className="text-3xl tracking-tight">
              {t('total')}:{' '}
              <span className="font-bold">
                {sum} {CURRENCY_SYMBOLS[currency]}
              </span>
            </CardTitle>
          </CardHeader>
          <CardContent className="flex flex-col gap-2">
            <Button
              variant={
                paymentMethod === PaymentMethod.EasyDonate
                  ? 'outline'
                  : 'secondary'
              }
              className="h-auto w-full border-1 border-transparent py-2"
              onClick={() => setPaymentMethod(PaymentMethod.EasyDonate)}
            >
              <div className="flex flex-col items-center gap-1">
                <div className="flex items-center gap-2">
                  <BanknoteIcon className="size-4" />
                  {t('paymentMethod.easyDonate')}
                </div>
                <p className="text-muted-foreground text-xs">EasyDonate</p>
              </div>
            </Button>
            <div className="flex items-center gap-2">
              <Button
                variant={
                  paymentMethod === PaymentMethod.DonationAlerts
                    ? 'outline'
                    : 'secondary'
                }
                className="h-auto flex-1 border-1 border-transparent py-2"
                onClick={() => setPaymentMethod(PaymentMethod.DonationAlerts)}
              >
                <div className="flex flex-col items-center gap-1">
                  <div className="flex items-center gap-2">
                    <HandCoinsIcon className="size-4" />
                    {t('paymentMethod.donationAlerts')}
                  </div>
                  <p className="text-muted-foreground text-xs">
                    DonationAlerts
                  </p>
                </div>
              </Button>
              <Button
                variant={
                  paymentMethod === PaymentMethod.Freekassa
                    ? 'outline'
                    : 'secondary'
                }
                className="h-auto flex-1 border-1 border-transparent py-2"
                onClick={() => setPaymentMethod(PaymentMethod.Freekassa)}
              >
                <div className="flex flex-col items-center gap-1">
                  <div className="flex items-center gap-2">
                    <BitcoinIcon className="size-4" />
                    {t('paymentMethod.freekassa')}
                  </div>
                  <p className="text-muted-foreground text-xs">FreeKassa</p>
                </div>
              </Button>
            </div>
          </CardContent>
          <CardFooter className="flex flex-col gap-2">
            <Button
              disabled={
                isCreatingOrder ||
                items.length === 0 ||
                ignoredItems.length >= items.length
              }
              size="lg"
              className="flex w-full items-center gap-2"
              onClick={onCreateOrder}
            >
              {t('createOrder')}
              {isCreatingOrder ? (
                <Loader2Icon className="size-4 animate-spin" />
              ) : (
                <ArrowRightIcon className="size-4" />
              )}
            </Button>
            <p className="text-muted-foreground hidden text-xs">
              <RichText>
                {(tags) =>
                  t.rich('orderConditions', {
                    ...tags,
                    paymentConditions: (chunks: React.ReactNode) => (
                      <Link
                        href="/wiki/legal/payments"
                        className="hover:text-primary underline underline-offset-2"
                      >
                        {chunks}
                      </Link>
                    ),
                    refundConditions: (chunks: React.ReactNode) => (
                      <Link
                        href="/wiki/legal/refund"
                        className="hover:text-primary underline underline-offset-2"
                      >
                        {chunks}
                      </Link>
                    ),
                  })
                }
              </RichText>
            </p>
          </CardFooter>
        </Card>
        <Alert>
          <AlertTitle className="flex items-center gap-2">
            <ClockIcon className="size-3" />
            {t('attention')}
          </AlertTitle>
          <AlertDescription>
            <p>
              <RichText>
                {(tags) => t.rich('attentionDescription', { ...tags })}
              </RichText>
            </p>
          </AlertDescription>
        </Alert>
      </div>
    </div>
  )
}
