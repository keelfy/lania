import { requireAdmin } from '@/lib/admin'
import { getAdminCosmetics, getAdminProducts } from '@/lib/api-endpoints'
import { serverApiFetcher } from '@/lib/server'
import { getTranslations } from 'next-intl/server'
import AdminPageHeader from '../admin-page-header'
import { EdCredentialsProvider } from './ed-credentials-context'
import ProductsManager from './products-manager'

type Props = { params: Promise<{ locale: string }> }

export default async function AdminProductsPage({ params }: Props) {
  await requireAdmin()
  const { locale } = await params
  const t = await getTranslations({ locale, namespace: 'admin' })
  const [products, catalog] = await Promise.all([
    getAdminProducts(serverApiFetcher),
    getAdminCosmetics(serverApiFetcher),
  ]).catch(() => [undefined, undefined] as const)
  return (
    <div className="flex flex-col gap-6">
      <AdminPageHeader title={t('nav.products')} />
      {products && catalog ? (
        <EdCredentialsProvider>
          <ProductsManager products={products} catalog={catalog} />
        </EdCredentialsProvider>
      ) : (
        <p className="text-destructive py-10 text-center">
          {t('products.loadFailed')}
        </p>
      )}
    </div>
  )
}
