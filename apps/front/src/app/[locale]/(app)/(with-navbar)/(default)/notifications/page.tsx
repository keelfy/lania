import { getMetadataLocale } from '@/i18n/metadata-locale'
import { getCurrentSession } from '@/lib/get-current-session'
import { getTranslations } from 'next-intl/server'
import { redirect } from 'next/navigation'
import NotificationFeed from './notification-feed'

type Props = {
  params: Promise<{
    locale: string
  }>
}

export async function generateMetadata({ params }: Props) {
  const locale = await getMetadataLocale(params)
  const t = await getTranslations({ locale, namespace: 'notifications' })
  return { title: t('title') }
}

export default async function NotificationsPage({ params }: Props) {
  const { locale } = await params
  const session = await getCurrentSession()
  const t = await getTranslations({ locale, namespace: 'notifications' })

  if (session?.active !== true) {
    return redirect('/auth/login')
  }

  return (
    <div className="mx-auto flex w-full max-w-2xl flex-col gap-4">
      <NotificationFeed title={t('title')} />
    </div>
  )
}
