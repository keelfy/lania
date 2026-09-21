import { Button } from '@/components/ui/button'
import PlayerCard from '@/components/ui/player-card'
import RichText from '@/components/ui/rich-text'
import { getTranslations } from 'next-intl/server'
import Link from 'next/link'
import React from 'react'
import { ProfilePageData } from './load-profile-page'
import ProfileSelectWrapper from './profile-select-wrapper'

const SECTIONS = [
  { key: 'overview', path: '/profiles' },
  { key: 'settings', path: '/profiles/settings' },
] as const

type Props = React.PropsWithChildren<{
  locale: string
  active: (typeof SECTIONS)[number]['key']
  data: ProfilePageData
  // The season the access status is shown for. It stays when the section changes.
  seasonId?: string
}>

// Every page of the profiles section renders its own shell: a layout does not render again on navigation.
export default async function ProfileShell({
  locale,
  active,
  data,
  seasonId,
  children,
}: Props) {
  const t = await getTranslations({ locale, namespace: 'profiles' })
  const { profiles, profileId, selectedProfile, freeAccess } = data

  const select = (
    <ProfileSelectWrapper profiles={profiles} selectedProfileId={profileId} />
  )

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-col gap-6 lg:flex-row">
        <div className="flex flex-col items-center gap-2 lg:hidden lg:flex-row lg:justify-between">
          <div className="flex items-center gap-2">
            <p className="text-xl font-bold">{t('title')}</p>
            <p className="text-muted-foreground text-sm font-bold">
              {`(${profiles.length}/2)`}
            </p>
          </div>
          {select}
        </div>
        <PlayerCard
          profileId={profileId}
          username={selectedProfile?.username}
          nameCosmetics={selectedProfile?.cosmetics.name}
          locale={locale}
          className="flex-1"
        />
        <div className="flex flex-1 flex-col gap-4">
          <div className="hidden flex-col items-center gap-2 lg:flex lg:flex-row lg:justify-between">
            <h2 className="mb-2 text-xl font-bold">
              {t('title')}&nbsp;
              <span className="text-muted-foreground text-sm">
                {`(${profiles.length}/2)`}
              </span>
            </h2>
            {select}
          </div>
          {profileId && selectedProfile ? (
            <>
              <nav className="flex gap-2">
                {SECTIONS.map(({ key, path }) => {
                  const query = new URLSearchParams({ id: profileId })
                  if (seasonId) query.set('s', seasonId)
                  return (
                    <Button
                      key={key}
                      asChild
                      variant={key === active ? 'default' : 'outline'}
                      size="sm"
                    >
                      <Link href={`/${locale}${path}?${query}`}>
                        {t(`nav.${key}`)}
                      </Link>
                    </Button>
                  )
                })}
              </nav>
              <div className="flex w-full flex-1 flex-col gap-4">
                {children}
              </div>
            </>
          ) : (
            <div className="flex flex-1 flex-col items-center justify-center gap-2">
              <p className="text-muted-foreground text-center text-base">
                <RichText>
                  {(tags) =>
                    t.rich(freeAccess ? 'selectProfileFree' : 'selectProfile', {
                      ...tags,
                      obtainAccess: (chunks: React.ReactNode) => (
                        <Button variant="link" asChild className="h-auto p-0">
                          <Link href="/obtain-access">{chunks}</Link>
                        </Button>
                      ),
                    })
                  }
                </RichText>
              </p>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
