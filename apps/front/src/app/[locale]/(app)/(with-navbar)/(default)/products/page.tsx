import { Button } from '@/components/ui/button'
import CurrencySelect from '@/components/ui/currency-select'
import {
  Drawer,
  DrawerClose,
  DrawerContent,
  DrawerHeader,
  DrawerTitle,
  DrawerTrigger,
} from '@/components/ui/drawer'
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
import {
  ChevronDownIcon,
  ChevronRightIcon,
  CoinsIcon,
  LayoutGridIcon,
  PaletteIcon,
  ShieldIcon,
  TagIcon,
} from 'lucide-react'
import { Metadata } from 'next'
import { getTranslations } from 'next-intl/server'
import { cookies } from 'next/headers'
import Link from 'next/link'
import UsernameColorProductCard from './components/name-color-product-card'
import NamePrefixProductCard from './components/name-prefix-product-card'
import UpgradeProductCard from './components/upgrade-color-product-card'

const CATEGORIES = [
  {
    id: 'all',
    icon: LayoutGridIcon,
  },
  {
    id: 'upgrade',
    icon: ShieldIcon,
  },
  {
    id: 'name-color',
    icon: PaletteIcon,
  },
  {
    id: 'name-prefix',
    icon: TagIcon,
  },
]

type Props = {
  searchParams: Promise<{
    category?: string
  }>
  params: Promise<{
    locale: string
  }>
}

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { locale } = await params
  const t = await getTranslations({ locale, namespace: 'products.metadata' })

  return {
    title: t('title'),
    description: t('description'),
    openGraph: {
      type: 'website',
      url: 'https://lania.gg/products',
      title: t('title'),
      description: t('description'),
      siteName: 'Lania.GG',
    },
  }
}

export default async function ShopPage({ searchParams, params }: Props) {
  const { category } = await searchParams
  const { locale } = await params
  const currency =
    ((await cookies()).get(CURRENCY_COOKIE)?.value as Currency) ??
    DEFAULT_CURRENCY

  const products = await getProducts(
    serverApiFetcher,
    category,
    locale,
    currency,
  ).catch((err) => {
    console.error(err)
    return []
  })

  const t = await getTranslations({ locale, namespace: 'products' })

  const isCategoryActive = (cat: (typeof CATEGORIES)[number]) =>
    cat.id === category || (category === undefined && cat.id === 'all')

  return (
    <div className="flex gap-8">
      <div className="hidden flex-col gap-4 sm:flex">
        {CATEGORIES.map((cat) => (
          <Button
            key={cat.id}
            variant={isCategoryActive(cat) ? 'default' : 'outline'}
            size="lg"
            className="w-full min-w-64 justify-start rounded-md py-6 text-left"
            asChild
          >
            <Link
              href={
                cat.id !== 'all' ? `/products?category=${cat.id}` : '/products'
              }
            >
              <div className="flex items-center justify-start gap-2 text-lg">
                <cat.icon className="size-4 stroke-3" />
                <p className="font-bold tracking-tight">
                  {t(`categories.${cat.id}`)}
                </p>
              </div>
            </Link>
          </Button>
        ))}
        <div className="mt-4 space-y-1.5">
          <div className="flex items-center gap-2">
            <CoinsIcon className="size-4" />
            <p className="text-lg font-semibold tracking-tight">Валюта</p>
          </div>
          <CurrencySelect currency={currency} className="w-full" />
        </div>
      </div>
      <div className="flex w-full flex-1 flex-col gap-4">
        <div className="flex items-center justify-between">
          <p className="text-4xl font-extrabold tracking-tight">
            {t('title')}
            <span className="hidden md:inline">
              &nbsp;→&nbsp;
              {t(`categories.${category ?? 'all'}`)}
              &nbsp;
              <span className="text-muted-foreground text-3xl">
                ({products.length})
              </span>
            </span>
          </p>
          <Drawer>
            <DrawerTrigger asChild>
              <Button variant="secondary" size="default" className="sm:hidden">
                {t(`categories.${category ?? 'all'}`)}
                <ChevronDownIcon className="size-4" />
              </Button>
            </DrawerTrigger>
            <DrawerContent>
              <DrawerHeader>
                <DrawerTitle>{t('categorySelectionTitle')}</DrawerTitle>
              </DrawerHeader>
              <div className="mb-6 flex min-h-32 flex-col gap-2">
                {CATEGORIES.map((cat) => (
                  <DrawerClose key={cat.id} asChild>
                    <Button
                      variant="secondary"
                      size="lg"
                      className="w-full rounded-none"
                      asChild
                    >
                      <Link
                        href={
                          cat.id !== 'all'
                            ? `/products?category=${cat.id}`
                            : '/products'
                        }
                      >
                        {t(`categories.${cat.id}`)}
                        <ChevronRightIcon className="size-4" />
                      </Link>
                    </Button>
                  </DrawerClose>
                ))}
              </div>
            </DrawerContent>
          </Drawer>
        </div>
        <div className="grid w-full flex-1 grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
          {products.map((item) => {
            switch (item.category) {
              case ProductCategory.NameColor:
                return (
                  <UsernameColorProductCard
                    key={item.id}
                    item={item as Product<NameColorProductMetadata>}
                    currency={currency}
                  />
                )
              case ProductCategory.Upgrade:
                return (
                  <UpgradeProductCard
                    key={item.id}
                    item={item as Product<UpgradeProductMetadata>}
                    currency={currency}
                  />
                )
              case ProductCategory.NamePrefix:
                return (
                  <NamePrefixProductCard
                    key={item.id}
                    item={item as Product<NamePrefixProductMetadata>}
                    currency={currency}
                  />
                )
              default:
                return null
            }
          })}
        </div>
      </div>
    </div>
  )
}
