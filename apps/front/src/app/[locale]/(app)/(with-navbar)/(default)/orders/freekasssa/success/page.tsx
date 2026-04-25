import { redirect } from 'next/navigation'

type Props = {
  searchParams: Promise<{
    MERCHANT_ORDER_ID: string
  }>
}

export default async function OrderSuccessPage({ searchParams }: Props) {
  const { MERCHANT_ORDER_ID: orderId } = await searchParams

  return redirect(`/orders/${orderId}/success`)
}
