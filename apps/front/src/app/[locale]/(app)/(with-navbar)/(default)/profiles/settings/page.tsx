import { ProfileResyncButton } from '@/components/profile-resync-card'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardAction,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { CannabisIcon, PaletteIcon } from 'lucide-react'
import { getTranslations } from 'next-intl/server'
import Link from 'next/link'
import { notFound } from 'next/navigation'
import { loadCosmeticOptions, loadProfilePage } from '../load-profile-page'
import NameColorOptionSelect from '../name-color-option-select'
import NameGlythOptionSelect from '../name-glyth-option-select'

type Props = {
  searchParams: Promise<{
    id: string | undefined
  }>
  params: Promise<{
    locale: string
  }>
}

// What can be changed on a game profile.
export default async function ProfileSettingsPage({
  searchParams,
  params,
}: Props) {
  const { id } = await searchParams
  const { locale } = await params
  const t = await getTranslations({ locale, namespace: 'profiles' })
  const { selectedProfile } = await loadProfilePage(id)
  if (!selectedProfile) notFound()
  const cosmeticOptions = await loadCosmeticOptions(selectedProfile.id)

  return (
    <>
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">{t('cosmetics.title')}</CardTitle>
          <CardDescription>{t('cosmetics.description')}</CardDescription>
          <CardAction>
            <ProfileResyncButton profileId={selectedProfile.id} />
          </CardAction>
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
                      query: { u: selectedProfile.username },
                    }}
                  >
                    {t('cosmetics.buyInStore')}
                  </Link>
                </Button>
              </p>
            </div>
          </div>
        </CardContent>
      </Card>
    </>
  )
}
