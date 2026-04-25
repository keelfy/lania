import {
  getBasket,
  getProductByIDs,
  getUserProfiles,
  getPurchases,
} from '@/lib/api-endpoints'
import { Currency, CURRENCY_COOKIE, DEFAULT_CURRENCY } from '@/lib/currency'
import { cookies } from 'next/headers'
import BasketList from './basket-list'
import BasketTitle from './basket-title'
import { serverApiFetcher } from '@/lib/server'

type Props = {
  params: Promise<{
    locale: string
  }>
}

export default async function BasketPage({ params }: Props) {
  const { locale } = await params
  const currency =
    (await cookies()).get(CURRENCY_COOKIE)?.value ?? DEFAULT_CURRENCY

  const [profiles, basket] = await Promise.all([
    await getUserProfiles(serverApiFetcher).catch((err) => {
      console.error(err)
      return []
    }),
    await getBasket(serverApiFetcher).catch((err) => {
      console.error(err)
      return []
    }),
  ])

  const productIds = basket.map((item) => item.productId)

  const [purchases, products] =
    productIds.length > 0
      ? await Promise.all([
          await getPurchases(serverApiFetcher, productIds).catch((err) => {
            console.error(err)
            return []
          }),
          await getProductByIDs(
            serverApiFetcher,
            productIds,
            locale,
            currency,
          ).catch((err) => {
            console.error(err)
            return []
          }),
        ])
      : [[], []]

  return (
    <div className="flex flex-col gap-4">
      <BasketTitle
        className="text-4xl font-extrabold tracking-tight"
        currency={(currency as Currency) ?? DEFAULT_CURRENCY}
      />
      <BasketList
        profiles={profiles}
        purchases={purchases}
        products={products}
        currency={(currency as Currency) ?? DEFAULT_CURRENCY}
      />
    </div>
  )
}
