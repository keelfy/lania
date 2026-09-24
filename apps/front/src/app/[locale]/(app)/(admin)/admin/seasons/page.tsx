import { requireAdmin } from '@/lib/admin'
import { getAdminSeasons } from '@/lib/api-endpoints'
import { serverApiFetcher } from '@/lib/server'
import { getTranslations } from 'next-intl/server'
import AdminPageHeader from '../admin-page-header'
import SeasonManager from './season-manager'

type Props = {
  params: Promise<{ locale: string }>
}

export default async function AdminSeasonsPage({ params }: Props) {
  await requireAdmin()
  const { locale } = await params
  const t = await getTranslations({ locale, namespace: 'admin' })
  const seasons = await getAdminSeasons(serverApiFetcher).catch((error) => {
    console.error(error)
    return undefined
  })

  return (
    <div className="flex flex-col gap-6">
      <AdminPageHeader title={t('nav.seasons')} />
      {!seasons ? (
        <p className="text-destructive py-10 text-center">
          {t('seasons.loadFailed')}
        </p>
      ) : (
        <SeasonManager seasons={seasons} locale={locale} />
      )}
    </div>
  )
}
