import ProfileResyncCard from '@/components/profile-resync-card'
import { requireAdmin } from '@/lib/admin'
import {
  getAdminCosmetics,
  getAdminGrants,
  getAdminProfile,
  getProfileMerges,
  getSeasons,
  getProducts,
} from '@/lib/api-endpoints'
import { serverApiFetcher } from '@/lib/server'
import { getTranslations } from 'next-intl/server'
import { notFound } from 'next/navigation'
import AdminPageHeader from '../../admin-page-header'
import GrantsCard from './grants-card'
import MergeCard from './merge-card'
import MergedFromCard from './merged-from-card'
import OwnerCard from './owner-card'
import RoleCard from './role-card'

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

  const [grants, seasons, products, catalog, merges] = await Promise.all([
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
    getProfileMerges(serverApiFetcher, profileId).catch((error) => {
      console.error(error)
      return []
    }),
  ])

  return (
    <div className="flex flex-col gap-6">
      <AdminPageHeader
        backHref={`/${locale}/admin/profiles`}
        backLabel={t('back')}
        title={profile.username}
        description={<span className="font-mono text-xs">{profile.id}</span>}
      />
      <OwnerCard profileId={profile.id} owner={profile.owner} locale={locale} />
      <RoleCard profileId={profile.id} role={profile.role} />
      <ProfileResyncCard profileId={profile.id} asAdmin />
      {merges.length > 0 && <MergedFromCard merges={merges} locale={locale} />}
      <MergeCard
        profileId={profile.id}
        profileUsername={profile.username}
        locale={locale}
      />
      <GrantsCard
        profileId={profile.id}
        grants={grants}
        seasons={seasons}
        products={products}
        catalog={catalog}
      />
    </div>
  )
}
