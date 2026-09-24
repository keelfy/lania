import { requireAdmin } from '@/lib/admin'
import { getAdminCosmetics, getAdminProducts } from '@/lib/api-endpoints'
import { serverApiFetcher } from '@/lib/server'
import { getTranslations } from 'next-intl/server'
import AdminPageHeader from '../admin-page-header'
import CosmeticsManager from './cosmetics-manager'

type Props = { params: Promise<{ locale: string }> }

export default async function AdminCosmeticsPage({ params }: Props) {
  await requireAdmin()
  const { locale } = await params
  const t = await getTranslations({ locale, namespace: 'admin' })
  const [catalog, products] = await Promise.all([
    getAdminCosmetics(serverApiFetcher),
    getAdminProducts(serverApiFetcher),
  ]).catch(() => [undefined, undefined] as const)
  return (
    <div className="flex flex-col gap-6">
      <AdminPageHeader title={t('nav.cosmetics')} />
      {catalog && products ? (
        <CosmeticsManager catalog={catalog} products={products} />
      ) : (
        <p className="text-destructive py-10 text-center">
          {t('cosmetics.loadFailed')}
        </p>
      )}
    </div>
  )
}
