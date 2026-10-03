import { requireAdmin } from '@/lib/admin'
import { getAdminPrivileges, getAdminProducts } from '@/lib/api-endpoints'
import { serverApiFetcher } from '@/lib/server'
import { getTranslations } from 'next-intl/server'
import AdminPageHeader from '../admin-page-header'
import PrivilegesManager from './privileges-manager'

type Props = { params: Promise<{ locale: string }> }

export default async function AdminPrivilegesPage({ params }: Props) {
  await requireAdmin()
  const { locale } = await params
  const t = await getTranslations({ locale, namespace: 'admin' })
  const [privileges, products] = await Promise.all([
    getAdminPrivileges(serverApiFetcher),
    getAdminProducts(serverApiFetcher),
  ]).catch(() => [undefined, undefined] as const)
  return privileges && products ? (
    <PrivilegesManager
      title={t('nav.privileges')}
      privileges={privileges}
      products={products}
    />
  ) : (
    <div className="flex flex-col gap-6">
      <AdminPageHeader title={t('nav.privileges')} />
      <p className="text-destructive py-10 text-center">
        {t('privileges.loadFailed')}
      </p>
    </div>
  )
}
