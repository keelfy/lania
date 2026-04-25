import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { getOrder } from '@/lib/api-endpoints'
import { serverApiFetcher } from '@/lib/server'
import { ArrowLeftIcon, ClockIcon, ShoppingBasketIcon } from 'lucide-react'
import Link from 'next/link'
import { redirect } from 'next/navigation'

type Props = {
  params: Promise<{
    locale: string
    orderId: string
  }>
}

export default async function OrderFailurePage({ params }: Props) {
  const { locale, orderId } = await params
  const order = await getOrder(serverApiFetcher, orderId).catch((err) => {
    console.error(err)
    return undefined
  })

  if (order?.status === 'completed') {
    return redirect(`/${locale}/orders/${orderId}/${order.status}`)
  } else if (order?.status !== 'failed') {
    return redirect('/products')
  }

  return (
    <div className="flex flex-col items-center justify-center gap-10 py-20">
      <h1 className="text-destructive text-4xl font-bold">
        Упс... Что-то пошло не так!
      </h1>
      <div className="flex flex-col gap-4">
        <Alert className="max-w-xl">
          <AlertTitle className="flex items-center gap-2">
            <ClockIcon className="size-4" />
            Не удалось завершить оплату
          </AlertTitle>
          <AlertDescription>
            <p>
              Если оплату отклонили не вы сами, то попробуйте повторить оплату.
              Если проблема не решится, пожалуйста, обратитесь в&nbsp;
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
      <Button variant="secondary" asChild>
        <Link href="/basket">
          <ArrowLeftIcon className="size-4" />
          Вернуться в корзину товаров
          <ShoppingBasketIcon className="size-4" />
        </Link>
      </Button>
    </div>
  )
}
