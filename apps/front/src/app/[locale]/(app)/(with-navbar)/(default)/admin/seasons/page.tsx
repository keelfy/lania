import { requireAdmin } from '@/lib/admin'
import { getAdminSeasons } from '@/lib/api-endpoints'
import { serverApiFetcher } from '@/lib/server'
import { getTranslations } from 'next-intl/server'
import AdminShell from '../admin-shell'
import SeasonManager from './season-manager'

type Props = {
  params: Promise<{ locale: string }>
}

export default async function AdminSeasonsPage({ params }: Props) {
  await requireAdmin()
  const { locale } = await params
  const t = await getTranslations({ locale, namespace: 'admin.seasons' })
  const seasons = await getAdminSeasons(serverApiFetcher).catch((error) => {
    console.error(error)
    return undefined
  })

  return (
    <AdminShell locale={locale} active="seasons">
      {!seasons ? (
        <p className="text-destructive py-10 text-center">{t('loadFailed')}</p>
      ) : (
        <SeasonManager seasons={seasons} locale={locale} />
      )}
    </AdminShell>
  )
}
