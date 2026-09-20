import { AspectRatio } from '@/components/ui/aspect-ratio'
import { Button } from '@/components/ui/button'
import { getSeasons } from '@/lib/api-endpoints'
import { serverApiFetcher } from '@/lib/server'
import { Season } from '@/models/season'
import { ClockIcon } from 'lucide-react'
import { getTranslations } from 'next-intl/server'
import Image from 'next/image'
import { Noto_Sans } from 'next/font/google'
import Link from 'next/link'

const notoSans = Noto_Sans({
  subsets: ['latin'],
})

type Props = {
  params: Promise<{ locale: string }>
}

export default async function SeasonsPage({ params }: Props) {
  const { locale } = await params
  const t = await getTranslations({ locale, namespace: 'seasons' })

  // A season without a preview is not part of the public list.
  const seasons = (
    await getSeasons(serverApiFetcher).catch(() => [] as Season[])
  ).filter(
    (season): season is Season & { previewImage: string } =>
      !!season.previewImage,
  )

  const toLocalDate = (millis?: number) => {
    if (millis === undefined) return '?'
    return new Date(millis).toLocaleDateString(locale, {
      day: 'numeric',
      month: 'long',
      year: 'numeric',
      timeZone: 'UTC',
    })
  }

  return (
    <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
      {seasons.map((season) => {
        return (
          <div
            key={season.id}
            className="bg-card flex w-full max-w-sm flex-col items-start justify-between gap-2 justify-self-center rounded-md p-6 shadow-md"
          >
            <div className="flex flex-col gap-4">
              <AspectRatio ratio={16 / 9}>
                <Image
                  src={season.previewImage}
                  alt={season.name}
                  width={690}
                  height={480}
                  className="rounded-md object-cover transition-transform duration-300 hover:scale-105"
                />
              </AspectRatio>
              <h2
                className={`text-xl font-bold uppercase ${notoSans.className} antialiased`}
              >
                {season.name}
              </h2>
              <div>
                <h3 className="text-sm">
                  {toLocalDate(season.startDate)}
                  &nbsp;&mdash;&nbsp;
                  {toLocalDate(season.endDate)}
                </h3>
                <div className="flex items-center gap-2 text-sm">
                  <ClockIcon className="text-muted-foreground size-3" />
                  <p>{t(`${season.id}.length`)}</p>
                </div>
              </div>
              <p className="text-muted-foreground text-sm">
                {t(`${season.id}.description`)}
              </p>
            </div>
            <Button variant="link" className="hidden h-auto p-0" asChild>
              <Link href={`/${locale}/seasons/${season.id}`}>
                {t('readMore')}
              </Link>
            </Button>
          </div>
        )
      })}
    </div>
  )
}
