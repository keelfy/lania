import { getSelectableSeasons } from '@/lib/seasons'
import { notFound } from 'next/navigation'
import { loadProfilePage } from '../load-profile-page'
import ProfilePasswordCard from './profile-password-card'
import ProfileUsernameCard from './profile-username-card'

type Props = {
  searchParams: Promise<{
    id: string | undefined
  }>
}

// What can be changed on a game profile besides its look.
export default async function ProfileSettingsPage({ searchParams }: Props) {
  const { id } = await searchParams
  const { seasons, selectedProfile: profile } = await loadProfilePage(id)
  if (!profile) notFound()

  // A licensed profile logs in with its Minecraft account and has no password in game.
  const licensed = !!profile.mojangUuid && profile.mojangUuid === profile.mcUuid
  const passwordSeasons = getSelectableSeasons(seasons).filter((season) =>
    profile.accesses.some(
      (access) => access.seasonId === season.id && access.status === 'active',
    ),
  )

  return (
    <>
      <ProfileUsernameCard key={profile.id} profile={profile} />
      {!licensed && (
        <ProfilePasswordCard
          key={`password-${profile.id}`}
          profile={profile}
          seasons={passwordSeasons}
        />
      )}
    </>
  )
}
