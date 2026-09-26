import { requireAdmin } from '@/lib/admin'
import { getAdminCosmetics, getAdminProducts } from '@/lib/api-endpoints'
import { serverApiFetcher } from '@/lib/server'
import { getTranslations } from 'next-intl/server'
import AdminPageHeader from '../admin-page-header'
import { EdCredentialsProvider } from './ed-credentials-context'
import ProductsManager, { DefaultDescriptions } from './products-manager'

type Props = { params: Promise<{ locale: string }> }

export default async function AdminProductsPage({ params }: Props) {
  await requireAdmin()
  const { locale } = await params
  const t = await getTranslations({ locale, namespace: 'admin' })
  // A product carries texts in both languages, so the defaults of both are needed whatever the admin's locale.
  const [ru, en] = await Promise.all(
    (['ru', 'en'] as const).map((textLocale) =>
      getTranslations({
        locale: textLocale,
        namespace: 'admin.products.defaultDescriptions',
      }),
    ),
  )
  const defaultDescriptions: DefaultDescriptions = {
    'name-color': { ru: ru('name-color'), en: en('name-color') },
    'name-prefix': { ru: ru('name-prefix'), en: en('name-prefix') },
  }
  const [products, catalog] = await Promise.all([
    getAdminProducts(serverApiFetcher),
    getAdminCosmetics(serverApiFetcher),
  ]).catch(() => [undefined, undefined] as const)
  return products && catalog ? (
    <EdCredentialsProvider>
      <ProductsManager
        title={t('nav.products')}
        products={products}
        catalog={catalog}
        defaultDescriptions={defaultDescriptions}
      />
    </EdCredentialsProvider>
  ) : (
    <div className="flex flex-col gap-6">
      <AdminPageHeader title={t('nav.products')} />
      <p className="text-destructive py-10 text-center">
        {t('products.loadFailed')}
      </p>
    </div>
  )
}
