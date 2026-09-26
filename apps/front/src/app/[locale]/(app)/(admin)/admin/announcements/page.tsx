import { requireAdmin } from '@/lib/admin'
import { getTranslations } from 'next-intl/server'
import AdminPageHeader from '../admin-page-header'
import AnnouncementForm from './announcement-form'

type Props = {
  params: Promise<{ locale: string }>
}

export default async function AdminAnnouncementsPage({ params }: Props) {
  await requireAdmin()
  const { locale } = await params
  const t = await getTranslations({ locale, namespace: 'admin' })

  return (
    <div className="flex flex-col gap-6">
      <AdminPageHeader
        title={t('nav.announcements')}
        description={t('announcements.description')}
      />
      <AnnouncementForm />
    </div>
  )
}
