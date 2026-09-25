import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from '@/components/ui/breadcrumb'
import McUsername from '@/components/ui/mc-username'
import NamePrefixes from '@/components/ui/name-prefixes'
import PlayerFace from '@/components/ui/player-face'
import VerifiedBadge from '@/components/ui/verified-badge'
import ProfileSeasonStats from '@/components/profile-season-stats'
import { getProfileDetailsByUsername } from '@/lib/api-endpoints'
import { PROFILE_ROLE_COLORS } from '@/lib/profile-colors'
import { serverApiFetcher } from '@/lib/server'
import { Metadata } from 'next'
import { getTranslations } from 'next-intl/server'
import { notFound } from 'next/navigation'
import PublicProfileStatusText from '../public-profile-status-text'
import CopyUsernameButton from './copy-username-button'
import ProfileSkinStage from './profile-skin-stage'
import SeenAt from './seen-at'

type Props = {
  params: Promise<{
    locale: string
    username: string
  }>
  searchParams: Promise<{
    // The season the cosmetics, the last seen date and the online status are of, the primary one when missing.
    season?: string
  }>
}

function getProfile(username: string, season?: string) {
  return getProfileDetailsByUsername(serverApiFetcher, username, season).catch(
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

export default async function CommunityProfilePage({
  params,
  searchParams,
}: Props) {
  const { locale, username } = await params
  const { season } = await searchParams
  const t = await getTranslations({ locale, namespace: 'community' })
  const tCard = await getTranslations({ locale, namespace: 'playerCard' })
  const profile = await getProfile(decodeURIComponent(username), season)
  if (!profile) return notFound()

  const nameColors = profile.cosmetics.name.colors?.colors ?? []
  const status = profile.isOnline ? 'online' : 'offline'

  return (
    <div className="flex flex-col gap-4">
      <Breadcrumb>
        <BreadcrumbList>
          <BreadcrumbItem>
            <BreadcrumbLink
              href={`/${locale}/community${season ? `?season=${season}` : ''}`}
            >
              {t('title')}
            </BreadcrumbLink>
          </BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>
            <BreadcrumbPage>{profile.username}</BreadcrumbPage>
          </BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>

      <div className="grid gap-8 lg:grid-cols-[22rem_minmax(0,1fr)] lg:items-start">
        <ProfileSkinStage
          mojangUuid={profile.mojangUuid}
          isSlimModel={profile.isSlimModel}
          colors={nameColors}
          className="h-[26rem] lg:sticky lg:top-24"
        />

        <div className="flex max-w-2xl flex-col gap-8">
          <header className="flex flex-col gap-3">
            <div className="flex items-center gap-3">
              {profile.mojangUuid && (
                <PlayerFace player={profile} className="size-10 rounded-sm" />
              )}
              <NamePrefixes cosmetics={profile.cosmetics.name} size={32} />
              <McUsername
                username={profile.username}
                colors={nameColors}
                className="text-4xl sm:text-5xl"
              />
              {profile.verified && <VerifiedBadge className="size-7" />}
              <CopyUsernameButton username={profile.username} />
            </div>
            <div className="flex flex-wrap items-center gap-x-4 gap-y-1">
              <p
                className="font-medium"
                style={{ color: PROFILE_ROLE_COLORS[profile.role] }}
              >
                {tCard(`roles.${profile.role}`)}
              </p>
              <PublicProfileStatusText status={status} className="-ml-2" />
            </div>
            <dl className="grid grid-cols-1 gap-x-8 gap-y-1 text-sm sm:grid-cols-2">
              <div>
                <dt className="text-muted-foreground">
                  {tCard('lastSeenAt.title')}
                </dt>
                <dd>
                  {profile.isOnline ? (
                    tCard('lastSeenAt.now')
                  ) : (
                    <SeenAt date={profile.lastSeenAt} locale={locale} />
                  )}
                </dd>
              </div>
              <div>
                <dt className="text-muted-foreground">
                  {tCard('firstSeenAt.title')}
                </dt>
                <dd>
                  {profile.isOnline && !profile.firstSeenAt ? (
                    tCard('firstSeenAt.now')
                  ) : (
                    <SeenAt date={profile.firstSeenAt} locale={locale} />
                  )}
                </dd>
              </div>
            </dl>
          </header>

          <ProfileSeasonStats
            profileId={profile.id}
            colors={nameColors}
            locale={locale}
          />
        </div>
      </div>
    </div>
  )
}
