'use client'

import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import { Label } from '@/components/ui/label'
import { getOrder } from '@/lib/api-endpoints'
import { clientApiFetcher } from '@/lib/client'
import { errorToast } from '@/lib/toasts'
import { cn } from '@/lib/utils'
import { Order } from '@/models/order'
import { ExternalLinkIcon, Loader2Icon } from 'lucide-react'
import Link from 'next/link'
import { useRouter } from 'next/navigation'
import React from 'react'

type Props = {
  orderId: string
}

const orderStatuses = {
  completed: 'completed',
  failed: 'failed',
  processing: 'processing',
  created: 'created',
}

const finalOrderStatuses = ['completed', 'failed']

export default function ProceedDonateForm({ orderId }: Props) {
  const [terms, setTerms] = React.useState(false)
  const [donateClicked, setDonateClicked] = React.useState(false)
  const [, startOrderFetching] = React.useTransition()
  const [order, setOrder] = React.useState<Order>()
  const router = useRouter()

  React.useEffect(() => {
    if (donateClicked) {
      // fetch order state every 3 seconds
      const interval = setInterval(() => {
        startOrderFetching(async () => {
          try {
            const order = await getOrder(clientApiFetcher, orderId)
            setOrder(order)
            if (finalOrderStatuses.includes(order.status)) {
              clearInterval(interval)
              router.push(`/orders/${orderId}/${order.status}`)
            }
          } catch (error) {
            console.error(error)
            clearInterval(interval)
            errorToast('Не удалось проверить статус заказа', error)
          }
        })
      }, 3000)
      return () => clearInterval(interval)
    }
  }, [donateClicked, orderId])

  return (
    <>
      <div className="mt-10 flex items-center gap-2">
        <Checkbox
          id="terms"
          checked={terms}
          onCheckedChange={(checked) =>
            setTerms(checked === 'indeterminate' ? false : checked)
          }
        />
        <Label htmlFor="terms">
          Я прочитал(-а) инструкцию по донату выше и введу все необходимые
          параметры
        </Label>
      </div>
      <Button className="mt-4" asChild size="lg" disabled={!terms}>
        <Link
          onClick={() => setDonateClicked(true)}
          href="https://www.donationalerts.com/r/laniabot"
          target="_blank"
          rel="noopener noreferrer"
          className={cn(
            terms ? 'opacity-100' : 'pointer-events-none opacity-70',
          )}
        >
          Перейти к донату <ExternalLinkIcon className="inline-block size-4" />
        </Link>
      </Button>
      {donateClicked &&
        (order?.status && finalOrderStatuses.includes(order?.status) ? (
          <div className="mt-8">
            <p className="text-3xl font-bold tracking-tight">
              Статус заказа: {orderStatuses[order?.status]}
            </p>
          </div>
        ) : (
          <div className="mt-8">
            <div className="flex items-center gap-2">
              <Loader2Icon className="text-muted-foreground size-8 animate-spin" />
              <p className="text-3xl font-bold tracking-tight">
                Проверка статуса заказа...
              </p>
            </div>
            <p className="text-muted-foreground mt-2">
              Мы ожидаем подтверждения доната. Это может занять некоторое время,
              пожалуйста, не закрывайте эту страницу
            </p>
          </div>
        ))}
    </>
  )
}
