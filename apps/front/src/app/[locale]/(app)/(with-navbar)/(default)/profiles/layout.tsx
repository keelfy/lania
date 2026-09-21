import Link from 'next/link'
import { Button } from '@/components/ui/button'
import RichText from '@/components/ui/rich-text'
import { cn } from '@/lib/utils'
import { getTranslations } from 'next-intl/server'
import React from 'react'
import { getProfiles, isFreeAccessSeason } from './load-profile-page'
import ProfileNav from './profile-nav'
import ProfilePlayerCard from './profile-player-card'
import ProfileSelectWrapper from './profile-select-wrapper'

type Props = React.PropsWithChildren<{
  params: Promise<{ locale: string }>
}>

// The frame of the profiles section. It stays on the screen while the tab changes, so the skin is not drawn again.
// Only the part next to the player card is rendered by the pages.
export default async function ProfilesLayout({ children, params }: Props) {
  const { locale } = await params
  const t = await getTranslations({ locale, namespace: 'profiles' })
  const [profiles, freeAccess] = await Promise.all([
    getProfiles(),
    isFreeAccessSeason(),
  ])
  const hasProfiles = profiles.length > 0

  return (
    <div className="flex flex-col gap-4 lg:grid lg:grid-cols-[24rem_minmax(0,1fr)] lg:grid-rows-[auto_auto_1fr] lg:gap-x-6">
      <div className="flex flex-col items-center gap-2 lg:col-start-2 lg:row-start-1 lg:flex-row lg:justify-between">
        <h2 className="text-xl font-bold">
          {t('title')}&nbsp;
          <span className="text-muted-foreground text-sm">
            {`(${profiles.length}/2)`}
          </span>
        </h2>
        <ProfileSelectWrapper profiles={profiles} />
      </div>
      <ProfilePlayerCard
        profiles={profiles}
        locale={locale}
        className="lg:col-start-1 lg:row-span-3 lg:row-start-1"
      />
      {hasProfiles && (
        <div className="lg:col-start-2 lg:row-start-2">
          <ProfileNav profiles={profiles} />
        </div>
      )}
      <div
        className={cn(
          'flex flex-col gap-4 lg:col-start-2',
          hasProfiles ? 'lg:row-start-3' : 'lg:row-span-2 lg:row-start-2',
        )}
      >
        {hasProfiles ? (
          children
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
  )
}
