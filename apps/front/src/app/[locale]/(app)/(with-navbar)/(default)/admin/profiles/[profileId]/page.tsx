import { Button } from '@/components/ui/button'
import { requireAdmin } from '@/lib/admin'
import {
  getAdminCosmetics,
  getAdminGrants,
  getAdminProfile,
  getSeasons,
  getProducts,
} from '@/lib/api-endpoints'
import { serverApiFetcher } from '@/lib/server'
import { getTranslations } from 'next-intl/server'
import Link from 'next/link'
import { notFound } from 'next/navigation'
import AdminShell from '../../admin-shell'
import GrantsCard from './grants-card'
import OwnerCard from './owner-card'

type Props = {
  params: Promise<{
    locale: string
    profileId: string
  }>
}

export default async function AdminProfilePage({ params }: Props) {
  await requireAdmin()
  const { locale, profileId } = await params
  const t = await getTranslations({ locale, namespace: 'admin.profiles' })

  const profile = await getAdminProfile(serverApiFetcher, profileId).catch(
    (error) => {
      console.error(error)
      return undefined
    },
  )
  if (!profile) notFound()

  const [grants, seasons, products, catalog] = await Promise.all([
    getAdminGrants(serverApiFetcher, profileId).catch((error) => {
      console.error(error)
      return undefined
    }),
    getSeasons(serverApiFetcher).catch((error) => {
      console.error(error)
      return []
    }),
    getProducts(serverApiFetcher, undefined, locale).catch((error) => {
      console.error(error)
      return []
    }),
    getAdminCosmetics(serverApiFetcher).catch((error) => {
      console.error(error)
      return undefined
    }),
  ])

  return (
    <AdminShell locale={locale} active="profiles">
      <Button asChild variant="link" className="w-fit p-0">
        <Link href={`/${locale}/admin/profiles`}>{t('back')}</Link>
      </Button>
      <div>
        <h2 className="text-2xl font-bold">{profile.username}</h2>
        <p className="text-muted-foreground font-mono text-xs">{profile.id}</p>
      </div>
      <OwnerCard profileId={profile.id} owner={profile.owner} locale={locale} />
      <GrantsCard
        profileId={profile.id}
        grants={grants}
        seasons={seasons}
        products={products}
        catalog={catalog}
      />
    </AdminShell>
  )
}
