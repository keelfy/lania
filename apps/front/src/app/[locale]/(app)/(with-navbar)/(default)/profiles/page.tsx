import ProfileSeasonStats from '@/components/profile-season-stats'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { getAccessMode, isFreeAccess } from '@/lib/access-mode'
import { pickSeason } from '@/lib/seasons'
import { cn } from '@/lib/utils'
import {
  ActivityIcon,
  CalendarIcon,
  CheckIcon,
  ClockIcon,
  ShieldCheckIcon,
  ShieldIcon,
  ShoppingBagIcon,
  XIcon,
  ZapIcon,
} from 'lucide-react'
import { getTranslations } from 'next-intl/server'
import Link from 'next/link'
import { notFound } from 'next/navigation'
import AccessSeasonSelect from './access-season-select'
import { loadProfilePage } from './load-profile-page'

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

type Props = {
  searchParams: Promise<{
    id: string | undefined
    // The season the access status is shown for.
    s: string | undefined
  }>
  params: Promise<{
    locale: string
  }>
}

// What a game profile is: its access, its violations and what it did on the servers.
// What can be changed is on the settings page.
export default async function ProfilePage({ searchParams, params }: Props) {
  const { id, s: seasonParam } = await searchParams
  const { locale } = await params
  const t = await getTranslations({ locale, namespace: 'profiles' })
  const { seasons, selectedProfile } = await loadProfilePage(id)
  if (!selectedProfile) notFound()

  // The profile lists a status for every running season and every ended season it has access to.
  const accessSeasons = seasons.filter((season) =>
    selectedProfile.accesses.some((access) => access.seasonId === season.id),
  )
  const accessSeason = pickSeason(accessSeasons, seasonParam)
  const accessStatus =
    selectedProfile.accesses.find(
      (access) => access.seasonId === accessSeason?.id,
    )?.status ?? 'inactive'
  // An ended season cannot be bought or registered for any more.
  const canObtainAccess = accessSeason?.isActive ?? false
  const accessSeasonFree = isFreeAccess(getAccessMode(accessSeason))
  const accessStatusColor = accessStatusColors[accessStatus]
  const Icon = accessStatusIcons[accessStatus]

  return (
    <>
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">{t('accessStatus.title')}</CardTitle>
          <CardDescription className="text-muted-foreground text-sm">
            {t('accessStatus.description')}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-[auto_minmax(0,1fr)] gap-x-2 gap-y-1 sm:grid-cols-2">
            {accessSeasons.length > 1 && (
              <>
                <div className="flex min-h-9 flex-nowrap items-center gap-2">
                  <CalendarIcon className="text-muted-foreground size-4 stroke-3" />
                  <p className="text-md font-semibold tracking-tight">
                    {t('accessStatus.season')}
                  </p>
                </div>
                <AccessSeasonSelect
                  seasons={accessSeasons}
                  selectedSeasonId={accessSeason?.id}
                  profileId={selectedProfile.id}
                />
              </>
            )}
            <div className="flex min-h-9 flex-nowrap items-center gap-2">
              <ActivityIcon className="text-muted-foreground size-4 stroke-3" />
              <p className="text-md font-semibold tracking-tight">
                {t('accessStatus.status')}
              </p>
            </div>
            <div className="flex min-h-9 flex-nowrap items-center justify-end gap-2">
              <p
                className={cn(
                  'text-right text-sm font-semibold sm:text-base',
                  accessStatusColor,
                )}
              >
                {t(`accessStatus.names.${accessStatus}`)}
              </p>
              <Icon
                className={cn('size-4', accessStatusIconColors[accessStatus])}
              />
            </div>
            <div className="flex min-h-9 flex-nowrap items-center gap-2">
              <ShieldIcon className="text-muted-foreground size-4 stroke-3" />
              <p className="text-md font-semibold tracking-tight">
                {t('violations.title')}
              </p>
            </div>
            {/* Violations are not tracked yet, so there is never a list to show. */}
            <div className="flex min-h-9 flex-nowrap items-center justify-end gap-2">
              <p className="text-primary text-right text-sm font-semibold sm:text-base">
                {t('violations.noViolations')}
              </p>
              <ShieldCheckIcon className="size-4 text-green-500" />
            </div>
            {accessStatus !== 'active' && canObtainAccess && (
              <Button
                variant="outline"
                asChild
                size="sm"
                className="col-span-2 mt-3"
              >
                <Link
                  href={{
                    pathname: '/obtain-access',
                    query: {
                      u: selectedProfile.username,
                      s: accessSeason?.id,
                    },
                  }}
                >
                  {accessSeasonFree ? (
                    <ZapIcon className="size-4" />
                  ) : (
                    <ShoppingBagIcon className="size-4" />
                  )}
                  {t(
                    accessSeasonFree
                      ? 'accessStatus.obtainFree'
                      : 'accessStatus.obtain',
                  )}
                </Link>
              </Button>
            )}
          </div>
        </CardContent>
      </Card>
      <Card className="px-6">
        <ProfileSeasonStats
          profileId={selectedProfile.id}
          colors={selectedProfile.cosmetics.name.colors.colors}
          locale={locale}
        />
      </Card>
    </>
  )
}
