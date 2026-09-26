import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from '@/components/ui/breadcrumb'
import { getProduct, getProducts } from '@/lib/api-endpoints'
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
import {
  AlertTriangleIcon,
  ArrowRightIcon,
  ClockIcon,
  ShoppingBagIcon,
} from 'lucide-react'
import { getTranslations } from 'next-intl/server'
import { cookies } from 'next/headers'
import { notFound } from 'next/navigation'
import { Button } from '@/components/ui/button'
import { Link } from '@/i18n/navigation'
import UsernameColorProductCard from '../../components/name-color-product-card'
import NamePrefixProductCard from '../../components/name-prefix-product-card'
import UpgradeProductCard from '../../components/upgrade-color-product-card'
import AddToBasketForm from './add-to-basket-form'
import ProductProfileProvider from './product-profile-context'
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

// How many other products of the category are offered under the product.
const RELATED_COUNT = 3

// The category whose products can be tried on together with a product of the category.
const PAIRED_CATEGORY: Partial<Record<ProductCategory, ProductCategory>> = {
  [ProductCategory.NameColor]: ProductCategory.NamePrefix,
  [ProductCategory.NamePrefix]: ProductCategory.NameColor,
}

function listProducts(
  category: ProductCategory | undefined,
  locale: string,
  currency: Currency,
) {
  if (!category) return Promise.resolve([])
  return getProducts(serverApiFetcher, category, locale, currency).catch(
    (err) => {
      console.error(err)
      return []
    },
  )
}

export default async function ProductPage({ params }: Props) {
  const { id, locale } = await params
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

  const [sameCategory, pairings] = await Promise.all([
    listProducts(item.category, locale, currency),
    listProducts(PAIRED_CATEGORY[item.category], locale, currency),
  ])
  const related = sameCategory
    .filter((product) => product.id !== item.id)
    .slice(0, RELATED_COUNT)

  const getItemComponent = () => {
    switch (item.category) {
      case ProductCategory.NameColor:
        return (
          <NameColorProductDetails
            item={item as Product<NameColorProductMetadata>}
            pairings={pairings as Product<NamePrefixProductMetadata>[]}
          />
        )
      case ProductCategory.NamePrefix:
        return (
          <NamePrefixProductDetails
            item={item as Product<NamePrefixProductMetadata>}
            pairings={pairings as Product<NameColorProductMetadata>[]}
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

  const getProductCard = (product: Product<ProductMetadata>) => {
    switch (product.category) {
      case ProductCategory.NameColor:
        return (
          <UsernameColorProductCard
            key={product.id}
            item={product as Product<NameColorProductMetadata>}
            currency={currency}
          />
        )
      case ProductCategory.NamePrefix:
        return (
          <NamePrefixProductCard
            key={product.id}
            item={product as Product<NamePrefixProductMetadata>}
            currency={currency}
          />
        )
      case ProductCategory.Upgrade:
        return (
          <UpgradeProductCard
            key={product.id}
            item={product as Product<UpgradeProductMetadata>}
            currency={currency}
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
            <BreadcrumbLink asChild>
              <Link href="/products" className="flex items-center gap-2">
                <ShoppingBagIcon className="size-3" />
                {t('title')}
              </Link>
            </BreadcrumbLink>
          </BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>
            <BreadcrumbLink asChild>
              <Link href={`/products?category=${item.category}`}>
                {t(`categories.${item.category}`)}
              </Link>
            </BreadcrumbLink>
          </BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>
            <BreadcrumbPage>{item.name}</BreadcrumbPage>
          </BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
      <ProductProfileProvider productId={item.id}>
        <div className="mt-4 flex flex-col gap-8 lg:flex-row lg:items-start">
          <div className="min-w-0 flex-1">{getItemComponent()}</div>
          <div className="flex flex-col gap-4 lg:sticky lg:top-8 lg:w-xs">
            <div className="bg-card flex h-fit w-full flex-col rounded-lg p-8">
              <h3 className="text-4xl font-bold text-nowrap">
                {t('page.price', {
                  price: `${item.price} ${CURRENCY_SYMBOLS[currency]}`,
                })}
              </h3>
              <p className="text-muted-foreground mt-4 text-sm text-wrap">
                <AlertTriangleIcon className="mr-1 inline-block size-4 align-text-bottom text-yellow-500" />
                {t.rich('page.seasonNotice', {
                  b: (chunks) => <span className="font-bold">{chunks}</span>,
                })}
              </p>
              <AddToBasketForm
                productId={item.id}
                category={item.category}
                className="mt-10"
              />
            </div>
            <p className="bg-card h-fit w-full rounded-lg px-6 py-4 text-sm">
              <ClockIcon className="mr-1 inline-block size-4 align-text-bottom" />
              {t('page.deliveryNotice')}
            </p>
          </div>
        </div>
      </ProductProfileProvider>
      {related.length > 0 && (
        <section className="mt-12 flex flex-col gap-4">
          <div className="flex items-center justify-between gap-4">
            <h2 className="text-2xl font-semibold tracking-tight">
              {t('page.related')}
            </h2>
            <Button variant="ghost" asChild>
              <Link href={`/products?category=${item.category}`}>
                {t('page.allInCategory')}
                <ArrowRightIcon className="size-4" />
              </Link>
            </Button>
          </div>
          <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
            {related.map(getProductCard)}
          </div>
        </section>
      )}
    </div>
  )
}
