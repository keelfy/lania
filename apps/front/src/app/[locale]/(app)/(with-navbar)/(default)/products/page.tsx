import { getMetadataLocale } from '@/i18n/metadata-locale'
import CurrencySelect from '@/components/ui/currency-select'
import { Currency, CURRENCY_COOKIE, DEFAULT_CURRENCY } from '@/lib/currency'
import { getCachedProductCatalog } from '@/lib/public-data'
import {
  Product,
  ProductCatalog,
  ProductCategory,
  ProductMetadata,
  UpgradeProductMetadata,
} from '@/models/product'
import { CoinsIcon } from 'lucide-react'
import { Metadata } from 'next'
import { getTranslations } from 'next-intl/server'
import { cookies } from 'next/headers'
import ProductCards from './components/product-cards'
import ProductsFeed from './components/products-feed'
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

  const allMockProducts = previewOnly
    ? getMockProducts(undefined, locale, currency)
    : []
  const page: ProductCatalog = previewOnly
    ? {
        content: allMockProducts.filter(
          (item) => category === 'all' || item.category === category,
        ),
        counts: countByCategory(allMockProducts),
      }
    : await getCachedProductCatalog(
        category === 'all' ? undefined : category,
        undefined,
        locale,
        currency,
      ).catch((err) => {
        console.error(err)
        return { content: [], counts: {} }
      })

  const t = await getTranslations({ locale, namespace: 'products' })

  const counts: Record<string, number> = {
    ...page.counts,
    all: Object.values(page.counts).reduce((sum, count) => sum + count, 0),
  }
  const labels = Object.fromEntries(
    ['all', ...Object.values(ProductCategory)].map((id) => [
      id,
      t(`categories.${id}`),
    ]),
  )
  const hasCosmetics = [ProductCategory.NameColor, ProductCategory.NamePrefix]
    .filter((id) => category === 'all' || category === id)
    .some((id) => (counts[id] ?? 0) > 0)

  // The season access is pinned above the grid. Only the first page is searched, upgrades sort first.
  const featuredProducts = page.content.filter(
    (item) =>
      item.category === ProductCategory.Upgrade &&
      (item.metadata as UpgradeProductMetadata).action === 'season_access',
  )
  const catalogProducts = page.content.filter(
    (item) => !featuredProducts.includes(item),
  )

  return (
    <div className="flex flex-col gap-6">
      <header className="flex flex-wrap items-baseline gap-x-4 gap-y-2">
        <h1 className="text-4xl font-extrabold tracking-tight">{t('title')}</h1>
        <p className="text-muted-foreground text-lg">
          {labels[category]}{' '}
          <span className="tabular-nums">({counts[category] ?? 0})</span>
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
            <ProductsFeed
              key={`${category}:${currency}`}
              category={category}
              initialCursor={page.nextCursor}
              initialCards={
                <ProductCards
                  items={catalogProducts}
                  currency={currency}
                  previewOnly={previewOnly}
                />
              }
            />
          </CatalogTryOn>
        </div>
      </div>
    </div>
  )
}

function countByCategory(items: Product<ProductMetadata>[]) {
  const counts: Record<string, number> = {}
  for (const item of items)
    counts[item.category] = (counts[item.category] ?? 0) + 1
  return counts
}
