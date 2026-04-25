import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { getOrders, getUserProfiles } from '@/lib/api-endpoints'
import {
  Currency,
  CURRENCY_COOKIE,
  CURRENCY_SYMBOLS,
  DEFAULT_CURRENCY,
} from '@/lib/currency'
import { getCurrentSession } from '@/lib/get-current-session'
import { serverApiFetcher } from '@/lib/server'
import { ArrowRightIcon, CheckIcon, ClockIcon, XIcon } from 'lucide-react'
import { getTranslations } from 'next-intl/server'
import { cookies } from 'next/headers'
import Link from 'next/link'
import { redirect } from 'next/navigation'
import OrderProductsSection from './order-products-section'

type Props = {
  params: Promise<{
    locale: string
  }>
}

const ORDER_STATUS_ICON = {
  completed: CheckIcon,
  failed: XIcon,
  processing: ClockIcon,
  created: ArrowRightIcon,
}

export default async function OrdersPage({ params }: Props) {
  const { locale } = await params
  const session = await getCurrentSession()
  const t = await getTranslations({ locale })
  const currency =
    ((await cookies()).get(CURRENCY_COOKIE)?.value as Currency) ??
    DEFAULT_CURRENCY

  if (session?.active !== true) {
    return redirect('/auth/login')
  }

  const orders = await getOrders(
    serverApiFetcher,
    session.identity?.id ?? '',
    locale,
  ).catch((err) => {
    console.error(err)
    return []
  })

  const profiles = await getUserProfiles(serverApiFetcher).catch((err) => {
    console.error(err)
    return []
  })

  return (
    <div className="flex flex-col gap-4">
      <h1 className="text-4xl font-extrabold tracking-tight">
        Мои заказы&nbsp;
        <span className="text-muted-foreground">({orders.length})</span>
      </h1>
      <div className="flex flex-col gap-4">
        {orders.length === 0 && (
          <div className="flex flex-col items-center justify-center gap-2 py-10 text-xl tracking-tight">
            <p>У вас пока нет заказов</p>
          </div>
        )}
        {orders.map((order) => {
          const amount = order.amounts.find(
            (amount) => amount.currency === currency,
          )?.amount
          const createdAt = new Date(order.createdAt)
          const badgeVariant =
            order.status === 'completed'
              ? 'default'
              : order.status === 'failed'
                ? 'destructive'
                : 'secondary'
          const StatusIcon = ORDER_STATUS_ICON[order.status]
          return (
            <div
              key={order.id}
              className="bg-card text-card-foreground flex flex-col gap-4 rounded-md border px-6 py-4"
            >
              <div
                key={order.id}
                className="flex flex-col items-start justify-between gap-6 lg:flex-row lg:items-center lg:gap-2"
              >
                <div className="flex flex-col gap-2">
                  <p className="text-xl font-semibold">
                    Заказ&nbsp;
                    <span className="text-muted-foreground hidden font-medium lg:inline">
                      [{order.id}]
                    </span>
                  </p>
                  <div className="flex items-center gap-2">
                    <ClockIcon className="size-4" />
                    <p className="text-md">
                      Время создания:&nbsp;
                      <span className="font-bold">
                        {createdAt.toLocaleString(locale)}
                      </span>
                    </p>
                  </div>
                  <OrderProductsSection
                    order={order}
                    profiles={profiles}
                    locale={locale}
                    currency={currency}
                  />
                </div>
                <div className="flex flex-col gap-3 lg:items-end">
                  <Badge variant={badgeVariant}>
                    <StatusIcon className="size-4" />
                    {t(`orders.status.${order.status}`)}
                  </Badge>
                  <p className="font-mono text-3xl font-bold">
                    {amount ?? 0}
                    {CURRENCY_SYMBOLS[currency]}
                  </p>
                </div>
              </div>
              {order.status === 'created' && (
                <Button size="default" asChild className="w-full md:w-fit">
                  <Link
                    href={`/${locale}/orders/${order.id}/donate`}
                    className="flex items-center gap-2"
                  >
                    <ArrowRightIcon className="inline-block size-4" />
                    Продолжить оплату
                  </Link>
                </Button>
              )}
            </div>
          )
        })}
      </div>
    </div>
  )
}
