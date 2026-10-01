import Link from 'next/link'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import { getTranslations } from 'next-intl/server'
import React from 'react'
import {
  getAllSeasons,
  getProfiles,
  isFreeAccessSeason,
} from './load-profile-page'
import ProfileLicenseCard from './profile-license-card'
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
  const [profiles, freeAccess, seasons] = await Promise.all([
    getProfiles(),
    isFreeAccessSeason(),
    getAllSeasons(),
  ])
  const hasProfiles = profiles.length > 0

  return (
    <div className="flex flex-col gap-4 lg:grid lg:grid-cols-[22rem_minmax(0,1fr)] lg:grid-rows-[auto_auto_1fr] lg:gap-x-6">
      <div
        className={cn(
          'flex flex-col items-center gap-2 lg:col-start-2 lg:row-start-1 xl:flex-row xl:justify-between',
          !hasProfiles && 'lg:col-span-2 lg:col-start-1',
        )}
      >
        <h2 className="text-xl font-bold">
          {t('title')}&nbsp;
          <span className="text-muted-foreground text-sm">
            {`(${profiles.length}/2)`}
          </span>
        </h2>
        <ProfileSelectWrapper profiles={profiles} />
      </div>
      {hasProfiles && (
        <div className="flex flex-col gap-4 lg:col-start-1 lg:row-span-3 lg:row-start-1">
          <ProfilePlayerCard
            profiles={profiles}
            locale={locale}
            className="sm:max-w-none sm:min-w-0"
          />
          <ProfileLicenseCard
            profiles={profiles}
            serverAddress={
              seasons.find((season) => season.isPrimary)?.publicAddress
            }
            className="w-full"
          />
        </div>
      )}
      {hasProfiles && (
        <div className="lg:col-start-2 lg:row-start-2">
          <ProfileNav profiles={profiles} />
        </div>
      )}
      <div
        className={cn(
          'flex flex-col gap-4 lg:col-start-2',
          hasProfiles
            ? 'lg:row-start-3'
            : 'lg:col-span-2 lg:col-start-1 lg:row-span-2 lg:row-start-2',
        )}
      >
        {hasProfiles ? (
          children
        ) : (
          <div className="bg-card flex min-h-64 flex-col items-center justify-center gap-3 rounded-xl border p-6 text-center">
            <h3 className="text-lg font-semibold">{t('empty.title')}</h3>
            <p className="text-muted-foreground max-w-sm text-sm">
              {t('empty.description')}
            </p>
            <Button asChild>
              <Link href="/obtain-access">
                {t(
                  freeAccess
                    ? 'accessStatus.obtainFree'
                    : 'accessStatus.obtain',
                )}
              </Link>
            </Button>
          </div>
        )}
      </div>
    </div>
  )
}
