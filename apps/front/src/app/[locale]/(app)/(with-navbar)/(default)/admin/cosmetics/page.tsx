import { requireAdmin } from '@/lib/admin'
import { getAdminCosmetics, getAdminProducts } from '@/lib/api-endpoints'
import { serverApiFetcher } from '@/lib/server'
import { getTranslations } from 'next-intl/server'
import AdminShell from '../admin-shell'
import CosmeticsManager from './cosmetics-manager'

type Props = { params: Promise<{ locale: string }> }

export default async function AdminCosmeticsPage({ params }: Props) {
  await requireAdmin()
  const { locale } = await params
  const t = await getTranslations({ locale, namespace: 'admin.cosmetics' })
  const [catalog, products] = await Promise.all([
    getAdminCosmetics(serverApiFetcher),
    getAdminProducts(serverApiFetcher),
  ]).catch(() => [undefined, undefined] as const)
  return (
    <AdminShell locale={locale} active="cosmetics">
      {catalog && products ? (
        <CosmeticsManager catalog={catalog} products={products} />
      ) : (
        <p className="text-destructive py-10 text-center">{t('loadFailed')}</p>
      )}
    </AdminShell>
  )
}
