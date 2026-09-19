import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from '@/components/ui/breadcrumb'
import PlayerCard from '@/components/ui/player-card'
import { getProfileDetailsByUsername } from '@/lib/api-endpoints'
import { serverApiFetcher } from '@/lib/server'
import { Metadata } from 'next'
import { getTranslations } from 'next-intl/server'
import { notFound } from 'next/navigation'

type Props = {
  params: Promise<{
    locale: string
    username: string
  }>
}

function getProfile(username: string) {
  return getProfileDetailsByUsername(serverApiFetcher, username).catch(
    (err) => {
      console.error(err)
      return undefined
    },
  )
}

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { locale, username } = await params
  const t = await getTranslations({ locale, namespace: 'community.profile' })
  const profile = await getProfile(decodeURIComponent(username))
  if (!profile) return {}

  const title = t('metadata.title', { username: profile.username })
  return {
    title,
    description: t('metadata.description', { username: profile.username }),
    openGraph: {
      type: 'profile',
      url: `${process.env.NEXT_PUBLIC_DOMAIN}/community/${profile.username}`,
      title,
      siteName: 'Lania Network',
    },
  }
}

export default async function CommunityProfilePage({ params }: Props) {
  const { locale, username } = await params
  const t = await getTranslations({ locale, namespace: 'community' })
  const profile = await getProfile(decodeURIComponent(username))
  if (!profile) return notFound()

  return (
    <div className="flex flex-col gap-4">
      <Breadcrumb>
        <BreadcrumbList>
          <BreadcrumbItem>
            <BreadcrumbLink href={`/${locale}/community`}>
              {t('title')}
            </BreadcrumbLink>
          </BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>
            <BreadcrumbPage>{profile.username}</BreadcrumbPage>
          </BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
      <div className="flex justify-center">
        <PlayerCard
          profileId={profile.id}
          nameCosmetics={profile.cosmetics.name}
          locale={locale}
        />
      </div>
    </div>
  )
}
