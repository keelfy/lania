import { AspectRatio } from '@/components/ui/aspect-ratio'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { getSeasons } from '@/lib/api-endpoints'
import { serverApiFetcher } from '@/lib/server'
import { Season } from '@/models/season'
import { CalendarIcon, ClockIcon } from 'lucide-react'
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
      <h1 className="col-span-full text-4xl font-extrabold tracking-tight">
        {t('title')}
      </h1>
      {seasons.map((season) => {
        return (
          <div
            key={season.id}
            className="bg-card flex w-full max-w-sm flex-col items-start justify-between gap-4 justify-self-center rounded-md border p-6 shadow-sm"
          >
            <div className="flex w-full flex-col gap-4">
              <AspectRatio
                ratio={16 / 9}
                className="relative w-full overflow-hidden rounded-md"
              >
                <Image
                  src={season.previewImage}
                  alt={season.name}
                  fill
                  sizes="(max-width: 768px) 100vw, (max-width: 1024px) 50vw, 33vw"
                  className="object-cover transition-transform duration-300 hover:scale-105"
                />
                {season.isActive && (
                  <Badge className="absolute top-2 right-2 border-transparent bg-teal-500 text-white">
                    <span className="size-1.5 rounded-full bg-white" />
                    {t('active')}
                  </Badge>
                )}
              </AspectRatio>
              <h2
                className={`text-xl font-bold uppercase ${notoSans.className} antialiased`}
              >
                {season.name}
              </h2>
              <div className="flex flex-wrap items-center gap-3">
                <div className="flex items-center gap-1.5">
                  <div className="bg-muted text-muted-foreground flex size-6 shrink-0 items-center justify-center rounded-md">
                    <CalendarIcon className="size-3" />
                  </div>
                  <span className="text-xs">
                    {season.isActive
                      ? `${t('since')} ${toLocalDate(season.startDate)}`
                      : `${toLocalDate(season.startDate)} — ${toLocalDate(season.endDate)}`}
                  </span>
                </div>
                {!season.isActive && (
                  <div className="flex items-center gap-1.5">
                    <div className="bg-muted text-muted-foreground flex size-6 shrink-0 items-center justify-center rounded-md">
                      <ClockIcon className="size-3" />
                    </div>
                    <span className="text-xs">{t(`${season.id}.length`)}</span>
                  </div>
                )}
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
