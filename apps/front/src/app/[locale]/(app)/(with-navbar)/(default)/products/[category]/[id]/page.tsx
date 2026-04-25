import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from '@/components/ui/breadcrumb'
import { getProduct } from '@/lib/api-endpoints'
import {
  Currency,
  CURRENCY_COOKIE,
  CURRENCY_SYMBOLS,
  DEFAULT_CURRENCY,
} from '@/lib/currency'
import { serverApiFetcher } from '@/lib/server'
import {
  NameColorProductMetadata,
  NamePrefixProductMetadata,
  Product,
  ProductCategory,
  ProductMetadata,
  UpgradeProductMetadata,
} from '@/models/product'
import { AlertTriangleIcon, ClockIcon, ShoppingBagIcon } from 'lucide-react'
import { getTranslations } from 'next-intl/server'
import { cookies } from 'next/headers'
import { notFound } from 'next/navigation'
import AddToBasketForm from './add-to-basket-form'
import NameColorProductDetails from './name-color-details'
import NamePrefixProductDetails from './name-prefix-details'
import UpgradeProductDetails from './upgrade-product-details'

type Props = {
  params: Promise<{
    id: string
    category: string
    locale: string
  }>
}

export default async function ProductPage({ params }: Props) {
  const { id, category, locale } = await params
  const t = await getTranslations({ locale, namespace: 'products' })
  const currency =
    ((await cookies()).get(CURRENCY_COOKIE)?.value as Currency) ??
    DEFAULT_CURRENCY

  const item = await getProduct<ProductMetadata>(
    serverApiFetcher,
    id,
    locale,
    currency,
  ).catch((err) => {
    console.error(err)
    return undefined
  })

  if (!item) {
    return notFound()
  }

  const getItemComponent = () => {
    switch (item.category) {
      case ProductCategory.NameColor:
        return (
          <NameColorProductDetails
            item={item as Product<NameColorProductMetadata>}
          />
        )
      case ProductCategory.NamePrefix:
        return (
          <NamePrefixProductDetails
            item={item as Product<NamePrefixProductMetadata>}
          />
        )
      case ProductCategory.Upgrade:
        return (
          <UpgradeProductDetails
            item={item as Product<UpgradeProductMetadata>}
          />
        )
      default:
        return null
    }
  }

  return (
    <div className="flex flex-col gap-4">
      <Breadcrumb>
        <BreadcrumbList>
          <BreadcrumbItem>
            <BreadcrumbLink
              href="/products"
              className="flex items-center gap-2"
            >
              <ShoppingBagIcon className="size-3" />
              Магазин
            </BreadcrumbLink>
          </BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>
            <BreadcrumbLink href={`/products?category=${category}`}>
              {t('categories.' + category)}
            </BreadcrumbLink>
          </BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>
            <BreadcrumbPage>{item.name}</BreadcrumbPage>
          </BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
      <div className="mt-4 flex flex-col gap-8 lg:flex-row">
        {getItemComponent()}
        <div className="flex flex-col gap-4">
          <div className="bg-card flex h-fit w-full flex-col rounded-lg p-8 lg:w-max lg:max-w-xs">
            <h3 className="text-4xl font-bold text-nowrap">
              {item.price} {CURRENCY_SYMBOLS[currency]} / 1 шт.
            </h3>
            <p className="text-muted-foreground mt-4 text-sm text-wrap">
              <AlertTriangleIcon className="mr-1 inline-block size-4 align-text-bottom text-yellow-500" />
              Вы покупаете товар на один сезон сервера. Минимальная длительность
              сезона — <span className="font-bold">3 месяца</span>.
            </p>
            <AddToBasketForm productId={id} className="mt-10" />
          </div>
          <p className="bg-card h-fit w-full rounded-lg px-6 py-4 text-sm lg:w-max lg:max-w-xs">
            <ClockIcon className="mr-1 inline-block size-4 align-text-bottom" />
            Товар будет доступен на игровом сервере Lania максимум через час
            после завершения оплаты.
          </p>
        </div>
      </div>
    </div>
  )
}
