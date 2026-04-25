import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { getOrder } from '@/lib/api-endpoints'
import { serverApiFetcher } from '@/lib/server'
import {
  AlertCircleIcon,
  ArrowLeftIcon,
  ArrowRightIcon,
  ClockIcon,
  ShoppingBagIcon,
  UserIcon,
} from 'lucide-react'
import Link from 'next/link'
import { redirect } from 'next/navigation'

type Props = {
  params: Promise<{
    locale: string
    orderId: string
  }>
}

export default async function OrderSuccessPage({ params }: Props) {
  const { locale, orderId } = await params
  const order = await getOrder(serverApiFetcher, orderId).catch((err) => {
    console.error(err)
    return undefined
  })

  if (order?.status === 'failed') {
    return redirect(`/${locale}/orders/${orderId}/${order.status}`)
  } else if (order?.status !== 'completed') {
    return redirect('/products')
  }

  return (
    <div className="flex flex-col items-center justify-center gap-10 py-20">
      <h1 className="text-4xl font-bold">Спасибо за покупку!</h1>
      <div className="flex flex-col gap-4">
        <Alert>
          <AlertTitle className="flex items-center gap-2">
            <AlertCircleIcon className="size-4" />
            <p>
              Если вы купили украшение, то для его активации перейдите в
              настройки профиля.
            </p>
          </AlertTitle>
        </Alert>
        <Alert>
          <AlertTitle className="flex items-center gap-2">
            <ClockIcon className="size-4" />
            Купленный товар будет доступен в вашем профиле максимум через час
          </AlertTitle>
          <AlertDescription>
            <p>
              Если спустя час товар не появился в вашем профиле, пожалуйста,
              обратитесь в&nbsp;
              <Link
                href={process.env.NEXT_PUBLIC_SUPPORT_URL ?? '#'}
                className="text-blue-600 underline"
              >
                поддержку
              </Link>
              .
            </p>
          </AlertDescription>
        </Alert>
      </div>
      <div className="flex items-center gap-6">
        <Button variant="secondary" asChild>
          <Link href="/products">
            <ArrowLeftIcon className="size-4" />
            Вернуться в магазин
            <ShoppingBagIcon className="size-4" />
          </Link>
        </Button>
        <Button asChild>
          <Link href="/profiles">
            <UserIcon className="size-4" />
            Перейти в настройки профиля
            <ArrowRightIcon className="size-4" />
          </Link>
        </Button>
      </div>
    </div>
  )
}
