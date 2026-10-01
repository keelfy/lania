import { getMetadataLocale } from '@/i18n/metadata-locale'
import CurrencySelect from '@/components/ui/currency-select'
import { getProducts } from '@/lib/api-endpoints'
import { Currency, CURRENCY_COOKIE, DEFAULT_CURRENCY } from '@/lib/currency'
import { serverApiFetcher } from '@/lib/server'
import {
  NameColorProductMetadata,
  NamePrefixProductMetadata,
  Product,
  ProductCategory,
  UpgradeProductMetadata,
} from '@/models/product'
import { CoinsIcon } from 'lucide-react'
import { Metadata } from 'next'
import { getTranslations } from 'next-intl/server'
import { cookies } from 'next/headers'
import UsernameColorProductCard from './components/name-color-product-card'
import NamePrefixProductCard from './components/name-prefix-product-card'
import UpgradeProductCard from './components/upgrade-color-product-card'
import { getMockProducts } from './mock-products'
import CatalogCategories from './components/catalog-categories'
import { CatalogTryOn } from './components/catalog-try-on'

type Props = {
  searchParams: Promise<{
    category?: string
    mock?: string
  }>
  params: Promise<{
    locale: string
  }>
}

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const locale = await getMetadataLocale(params)
  const t = await getTranslations({ locale, namespace: 'products.metadata' })

  return {
    title: t('title'),
    description: t('description'),
    openGraph: {
      type: 'website',
      url: 'https://lania.network/products',
      title: t('title'),
      description: t('description'),
      siteName: 'Lania.GG',
    },
  }
}

export default async function ShopPage({ searchParams, params }: Props) {
  const { category: requestedCategory, mock } = await searchParams
  const category = Object.values(ProductCategory).includes(
    requestedCategory as ProductCategory,
  )
    ? requestedCategory!
    : 'all'
  const previewOnly = process.env.NODE_ENV === 'development' && mock === '1'
  const { locale } = await params
  const currency =
    ((await cookies()).get(CURRENCY_COOKIE)?.value as Currency) ??
    DEFAULT_CURRENCY

  const allProducts = previewOnly
    ? getMockProducts(undefined, locale, currency)
    : await getProducts(serverApiFetcher, undefined, locale, currency).catch(
        (err) => {
          console.error(err)
          return []
        },
      )

  const t = await getTranslations({ locale, namespace: 'products' })

  const products =
    category === 'all'
      ? allProducts
      : allProducts.filter((item) => item.category === category)
  const counts: Record<string, number> = { all: allProducts.length }
  for (const item of allProducts)
    counts[item.category] = (counts[item.category] ?? 0) + 1
  const labels = Object.fromEntries(
    ['all', ...Object.values(ProductCategory)].map((id) => [
      id,
      t(`categories.${id}`),
    ]),
  )
  const hasCosmetics = products.some(
    (item) => item.category !== ProductCategory.Upgrade,
  )

  const featuredProducts = products.filter(
    (item) =>
      item.category === ProductCategory.Upgrade &&
      (item.metadata as UpgradeProductMetadata).action === 'season_access',
  )
  const catalogProducts = products.filter(
    (item) => !featuredProducts.includes(item),
  )

  return (
    <div className="flex flex-col gap-6">
      <header className="flex flex-wrap items-baseline gap-x-4 gap-y-2">
        <h1 className="text-4xl font-extrabold tracking-tight">{t('title')}</h1>
        <p className="text-muted-foreground text-lg">
          {labels[category]}{' '}
          <span className="tabular-nums">({products.length})</span>
        </p>
      </header>
      <div className="flex min-w-0 flex-col gap-6 sm:flex-row sm:gap-8">
        <aside className="flex shrink-0 flex-col gap-5 sm:w-56">
          <CatalogCategories
            category={category}
            labels={labels}
            counts={counts}
            previewOnly={previewOnly}
            label={t('categorySelectionTitle')}
          />
          <div className="flex items-center gap-3 sm:flex-col sm:items-stretch sm:gap-2">
            <div className="flex items-center gap-2">
              <CoinsIcon className="size-4" />
              <p className="text-sm font-semibold">{t('currency')}</p>
            </div>
            <CurrencySelect
              currency={currency}
              className="min-w-0 flex-1 sm:w-full"
            />
          </div>
        </aside>
        <div className="flex min-w-0 flex-1 flex-col gap-4">
          <CatalogTryOn
            showControls={hasCosmetics}
            leadingContent={featuredProducts.map((item) => (
              <UpgradeProductCard
                key={item.id}
                item={item as Product<UpgradeProductMetadata>}
                currency={currency}
                previewOnly={previewOnly}
                featured
              />
            ))}
          >
            {renderProducts()}
          </CatalogTryOn>
        </div>
      </div>
    </div>
  )

  function renderProducts() {
    return (
      <div className="grid w-full grid-cols-1 items-stretch gap-4 md:grid-cols-2 lg:grid-cols-3">
        {catalogProducts.map((item) => {
          switch (item.category) {
            case ProductCategory.NameColor:
              return (
                <UsernameColorProductCard
                  key={item.id}
                  item={item as Product<NameColorProductMetadata>}
                  currency={currency}
                  previewOnly={previewOnly}
                />
              )
            case ProductCategory.Upgrade:
              return (
                <UpgradeProductCard
                  key={item.id}
                  item={item as Product<UpgradeProductMetadata>}
                  currency={currency}
                  previewOnly={previewOnly}
                />
              )
            case ProductCategory.NamePrefix:
              return (
                <NamePrefixProductCard
                  key={item.id}
                  item={item as Product<NamePrefixProductMetadata>}
                  currency={currency}
                  previewOnly={previewOnly}
                />
              )
            default:
              return null
          }
        })}
      </div>
    )
  }
}
