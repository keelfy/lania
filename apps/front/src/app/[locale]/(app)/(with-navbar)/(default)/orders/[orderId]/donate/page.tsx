import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { getOrder } from '@/lib/api-endpoints'
import { Currency, CURRENCY_COOKIE, DEFAULT_CURRENCY } from '@/lib/currency'
import { serverApiFetcher } from '@/lib/server'
import { AlertTriangleIcon } from 'lucide-react'
import { cookies } from 'next/headers'
import Link from 'next/link'
import { redirect } from 'next/navigation'
import CopyButton from './copy-button'
import DonateAmountForm from './donate-amount-form'
import ProceedDonateForm from './proceed-donate-form'

type Props = {
  params: Promise<{
    locale: string
    orderId: string
  }>
}

export default async function PaymentPage({ params }: Props) {
  const { locale, orderId } = await params
  const defaultCurrency =
    ((await cookies()).get(CURRENCY_COOKIE)?.value as Currency) ??
    DEFAULT_CURRENCY

  const order = await getOrder(serverApiFetcher, orderId).catch((err) => {
    console.error(err)
    return undefined
  })

  if (!order) {
    return redirect('/basket')
  } else if (['completed', 'failed'].includes(order.status)) {
    return redirect(`/${locale}/orders/${orderId}/${order.status}`)
  }

  const availableCurrencyCodes = order.amounts
    .map((amount) => amount.currency)
    .sort()

  return (
    <div className="flex flex-col gap-0">
      <h1 className="text-5xl font-extrabold tracking-tight">Оплата донатом</h1>
      <div className="mt-4 flex flex-col gap-2">
        <p className="text-md">
          Чтобы система автоматически засчитала ваш донат, пожалуйста, введите
          следующие параметры:
        </p>
        <ol className="text-md ml-6 list-decimal">
          <DonateAmountForm
            defaultCurrency={defaultCurrency}
            availableCurrencyCodes={availableCurrencyCodes}
            amounts={order.amounts}
          />
          <li className="mt-4">
            <div className="flex items-center">
              <p className="text-lg font-semibold">Сообщение:</p>
              <p className="bg-muted ml-2 h-8 rounded-md rounded-r-none border px-2 py-1 font-mono font-semibold">
                {orderId}
              </p>
              <CopyButton
                copyText={orderId}
                className="h-8 rounded-l-none border-l-0 p-0"
                size="icon"
              />
            </div>
          </li>
        </ol>
      </div>
      <Alert className="mt-6" variant="destructive">
        <AlertTitle className="flex items-center gap-2">
          <AlertTriangleIcon className="size-4" />
          Внимание
        </AlertTitle>
        <AlertDescription>
          <p>
            Если ваш донат не будет засчитан по прошествии 12 часов, тогда,
            пожалуйста, обратитесь в&nbsp;
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
      <ProceedDonateForm orderId={orderId} />
    </div>
  )
}
