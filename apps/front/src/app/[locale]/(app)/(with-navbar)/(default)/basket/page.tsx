import {
  getBasket,
  getSeasons,
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

  const seasonIds = [...new Set(basket.map((item) => item.seasonId))]

  const seasons = await getSeasons(serverApiFetcher).catch((err) => {
    console.error(err)
    return []
  })

  const [purchases, products] =
    productIds.length > 0
      ? await Promise.all([
          // Purchases are looked up per season.
          Promise.all(
            seasonIds.map((seasonId) =>
              getPurchases(serverApiFetcher, productIds, seasonId).catch(
                (err) => {
                  console.error(err)
                  return []
                },
              ),
            ),
          ).then((bySeason) => bySeason.flat()),
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
        seasons={seasons}
        currency={(currency as Currency) ?? DEFAULT_CURRENCY}
      />
    </div>
  )
}
