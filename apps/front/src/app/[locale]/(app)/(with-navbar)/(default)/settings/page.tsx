import { getMetadataLocale } from '@/i18n/metadata-locale'
import { getTranslations } from 'next-intl/server'
import { Suspense } from 'react'
import AccountSettings from './account-settings'

type Props = {
  params: Promise<{
    locale: string
  }>
}

export async function generateMetadata({ params }: Props) {
  const locale = await getMetadataLocale(params)
  const t = await getTranslations({ locale, namespace: 'accountSettings' })
  return { title: t('metadata.title') }
}

// Kratos sends the browser here with ?flow=<id> after every change of the account.
export default async function AccountSettingsPage({ params }: Props) {
  const { locale } = await params
  const t = await getTranslations({ locale, namespace: 'accountSettings' })

  return (
    <div className="flex flex-col gap-4">
      <h1 className="text-4xl font-extrabold tracking-tight">{t('title')}</h1>
      <Suspense>
        <AccountSettings />
      </Suspense>
    </div>
  )
}
