import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import PlayerCard from '@/components/ui/player-card'
import RichText from '@/components/ui/rich-text'
import { getProfileCosmeticOptions, getUserProfiles } from '@/lib/api-endpoints'
import { getCurrentSession } from '@/lib/get-current-session'
import { serverApiFetcher } from '@/lib/server'
import { cn } from '@/lib/utils'
import { Profile, ProfileCosmeticOptions } from '@/models/profile'
import {
  ActivityIcon,
  CannabisIcon,
  CheckIcon,
  ClockIcon,
  PaletteIcon,
  ShieldCheckIcon,
  ShieldIcon,
  ShoppingBagIcon,
  XIcon,
} from 'lucide-react'
import { getTranslations } from 'next-intl/server'
import Link from 'next/link'
import NameColorOptionSelect from './name-color-option-select'
import NameGlythOptionSelect from './name-glyth-option-select'
import ProfileSelectWrapper from './profile-select-wrapper'

const accessStatusColors = {
  active: 'text-primary',
  inactive: 'text-destructive',
  expired: 'text-orange-500',
}

const accessStatusIcons = {
  active: CheckIcon,
  inactive: XIcon,
  expired: ClockIcon,
}

const accessStatusIconColors = {
  active: 'text-green-500',
  inactive: 'text-destructive',
  expired: 'text-orange-500',
}

const DEFAULT_COSMETIC_OPTIONS: ProfileCosmeticOptions = {
  name: { colors: [], glythPrefixes: [], specialPrefixes: [] },
}

type Props = {
  searchParams: Promise<{
    id: string | undefined
  }>
  params: Promise<{
    locale: string
  }>
}

const fetchProfiles = async (
  profileId: string | undefined,
  userId: string | undefined,
): Promise<[Profile[], ProfileCosmeticOptions]> => {
  if (profileId) {
    return await Promise.all([
      getUserProfiles(serverApiFetcher, userId).catch((error) => {
        console.error(error)
        return []
      }),
      getProfileCosmeticOptions(serverApiFetcher, profileId).catch((error) => {
        console.error(error)
        return DEFAULT_COSMETIC_OPTIONS
      }),
    ])
  }

  const profiles = await getUserProfiles(serverApiFetcher, userId).catch(
    (error) => {
      console.error(error)
      return []
    },
  )

  let cosmeticOptions: ProfileCosmeticOptions = {
    name: DEFAULT_COSMETIC_OPTIONS.name,
  }

  if (profiles.length > 0) {
    cosmeticOptions = await getProfileCosmeticOptions(
      serverApiFetcher,
      profiles[0].id,
    ).catch((error) => {
      console.error(error)
      return DEFAULT_COSMETIC_OPTIONS
    })
  }

  return [profiles, cosmeticOptions]
}

export default async function ProfilePage({ searchParams, params }: Props) {
  const { id } = await searchParams
  const { locale } = await params
  const t = await getTranslations({ locale, namespace: 'profiles' })
  const session = await getCurrentSession()

  const [profiles, cosmeticOptions] = await fetchProfiles(
    id,
    session?.identity?.id,
  )

  const profileId = id ?? (profiles.length > 0 ? profiles[0].id : undefined)

  const selectedProfile = profiles.find((profile) => profile.id === profileId)
  const accessStatus = selectedProfile?.accessStatus ?? 'inactive'
  const accessStatusColor = accessStatusColors[accessStatus]
  const Icon = accessStatusIcons[accessStatus]

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
          <ProfileSelectWrapper
            profiles={profiles}
            selectedProfileId={profileId}
          />
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
            <ProfileSelectWrapper
              profiles={profiles}
              selectedProfileId={profileId}
            />
          </div>
          <div className="flex w-full flex-1 flex-col gap-4">
            {profileId && profiles.length > 0 && selectedProfile ? (
              <>
                <Card>
                  <CardHeader>
                    <CardTitle className="text-lg">
                      {t('accessStatus.title')}
                    </CardTitle>
                    <CardDescription className="text-muted-foreground text-sm">
                      {t('accessStatus.description')}
                    </CardDescription>
                  </CardHeader>
                  <CardContent>
                    <div className="grid grid-cols-2 gap-2">
                      <div className="flex flex-nowrap items-center gap-2">
                        <ActivityIcon className="text-muted-foreground size-4 stroke-3" />
                        <p className="text-md font-semibold tracking-tight">
                          {t('accessStatus.status')}
                        </p>
                      </div>
                      <div className="flex flex-nowrap items-center justify-end gap-2">
                        <p className={cn('font-semibold', accessStatusColor)}>
                          {t(`accessStatus.names.${accessStatus}`)}
                        </p>
                        <Icon
                          className={cn(
                            'size-4',
                            accessStatusIconColors[accessStatus],
                          )}
                        />
                      </div>
                      {accessStatus !== 'active' && (
                        <Button
                          variant="outline"
                          asChild
                          size="sm"
                          className="col-span-2 mt-2"
                        >
                          <Link
                            href={{
                              pathname: '/obtain-access',
                              query: {
                                u: selectedProfile?.username,
                              },
                            }}
                          >
                            <ShoppingBagIcon className="size-4" />
                            {t('accessStatus.obtain')}
                          </Link>
                        </Button>
                      )}
                    </div>
                  </CardContent>
                </Card>
                <Card>
                  <CardHeader>
                    <CardTitle className="text-lg">
                      {t('violations.title')}
                    </CardTitle>
                    <CardDescription>
                      {t('violations.description')}
                    </CardDescription>
                  </CardHeader>
                  <CardContent>
                    <div className="flex items-center justify-between gap-2">
                      <div className="flex flex-nowrap items-center gap-2">
                        <ShieldIcon className="text-muted-foreground size-4 stroke-3" />
                        <p className="text-md font-semibold tracking-tight">
                          {t('violations.status')}
                        </p>
                      </div>
                      <div className="flex flex-nowrap items-center gap-2">
                        <p className={cn('text-primary font-semibold')}>
                          {t('violations.noViolations')}
                        </p>
                        <ShieldCheckIcon className="size-4 text-green-500" />
                      </div>
                    </div>
                  </CardContent>
                </Card>
                <Card>
                  <CardHeader>
                    <CardTitle className="text-lg">
                      {t('cosmetics.title')}
                    </CardTitle>
                    <CardDescription>
                      {t('cosmetics.description')}
                    </CardDescription>
                  </CardHeader>
                  <CardContent>
                    <div>
                      <div className="flex items-center justify-between gap-2">
                        <div className="flex flex-nowrap items-center gap-2">
                          <PaletteIcon className="text-muted-foreground size-4 stroke-3" />
                          <p className="text-md font-semibold tracking-tight">
                            {t('cosmetics.nameColor.title')}
                          </p>
                        </div>
                        <div className="flex w-1/3 items-center justify-end gap-2">
                          <NameColorOptionSelect
                            selectedProfile={selectedProfile}
                            cosmeticOptions={cosmeticOptions}
                          />
                        </div>
                      </div>
                      <div className="mt-4 flex items-center justify-between gap-2">
                        <div className="flex flex-nowrap items-center gap-2">
                          <CannabisIcon className="text-muted-foreground size-4 stroke-3" />
                          <p className="text-md font-semibold tracking-tight">
                            {t('cosmetics.glyth.title')}
                          </p>
                        </div>
                        <div className="flex w-1/3 items-center justify-end gap-2">
                          <NameGlythOptionSelect
                            selectedProfile={selectedProfile}
                            cosmeticOptions={cosmeticOptions.name}
                          />
                        </div>
                      </div>
                      <div className="text-muted-foreground mt-4 flex min-h-20 flex-1 flex-col items-center justify-center gap-2">
                        <p className="text-center text-sm">
                          {t('cosmetics.wantToStandOut')}
                          <br />
                          <Button variant="link" asChild className="h-auto p-0">
                            <Link
                              href={{
                                pathname: '/products',
                                query: {
                                  u: selectedProfile?.username,
                                },
                              }}
                            >
                              {t('cosmetics.buyInStore')}
                            </Link>
                          </Button>
                        </p>
                      </div>
                      {/* <div className="grid w-full flex-1 grid-cols-2 gap-2 self-center lg:grid-cols-4">
                    <div className="flex h-fit w-fit flex-col items-center justify-start gap-2 justify-self-center rounded-sm border px-4 py-2">
                      <PaperclipIcon className="size-16" />
                      <p className={cn('text-primary font-semibold')}>
                        Бумажка 1
                      </p>
                    </div>
                    <div className="flex h-fit w-fit flex-col items-center justify-start gap-2 justify-self-center rounded-sm border px-4 py-2">
                      <PaperclipIcon className="size-16" />
                      <p className={cn('text-primary font-semibold')}>
                        Бумажка 2
                      </p>
                    </div>
                    <div className="flex h-fit w-fit flex-col items-center justify-start gap-2 justify-self-center rounded-sm border px-4 py-2">
                      <PaperclipIcon className="size-16" />
                      <p className={cn('text-primary font-semibold')}>
                        Бумажка 3
                      </p>
                    </div>
                    <div className="flex h-fit w-fit flex-col items-center justify-start gap-2 justify-self-center rounded-sm border px-4 py-2">
                      <PaperclipIcon className="size-16" />
                      <p className={cn('text-primary font-semibold')}>
                        Бумажка 4
                      </p>
                    </div>
                  </div> */}
                    </div>
                  </CardContent>
                </Card>
              </>
            ) : (
              <div className="flex flex-1 flex-col items-center justify-center gap-2">
                <p className="text-muted-foreground text-center text-base">
                  <RichText>
                    {(tags) =>
                      t.rich('selectProfile', {
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
    </div>
  )
}
