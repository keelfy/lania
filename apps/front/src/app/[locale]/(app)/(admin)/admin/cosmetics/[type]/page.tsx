import { requireAdmin } from '@/lib/admin'
import { getAdminCosmetics, getAdminProducts } from '@/lib/api-endpoints'
import { serverApiFetcher } from '@/lib/server'
import { getTranslations } from 'next-intl/server'
import { notFound } from 'next/navigation'
import AdminPageHeader from '../../admin-page-header'
import CosmeticsManager from './cosmetics-manager'

const TYPES = { colors: 'color', prefixes: 'prefix' } as const

type Props = { params: Promise<{ locale: string; type: string }> }

export default async function AdminCosmeticsTypePage({ params }: Props) {
  await requireAdmin()
  const { locale, type: slug } = await params
  if (!(slug in TYPES)) notFound()
  const type = TYPES[slug as keyof typeof TYPES]
  const t = await getTranslations({ locale, namespace: 'admin' })
  const title = t(type === 'color' ? 'nav.nameColors' : 'nav.namePrefixes')
  const [catalog, products] = await Promise.all([
    getAdminCosmetics(serverApiFetcher),
    getAdminProducts(serverApiFetcher),
  ]).catch(() => [undefined, undefined] as const)
  return catalog && products ? (
    <CosmeticsManager
      key={type}
      type={type}
      title={title}
      catalog={catalog}
      products={products}
    />
  ) : (
    <div className="flex flex-col gap-6">
      <AdminPageHeader title={title} />
      <p className="text-destructive py-10 text-center">
        {t('cosmetics.loadFailed')}
      </p>
    </div>
  )
}
