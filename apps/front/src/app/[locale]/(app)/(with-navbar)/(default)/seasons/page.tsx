import { AspectRatio } from '@/components/ui/aspect-ratio'
import { Button } from '@/components/ui/button'
import { ClockIcon } from 'lucide-react'
import { getTranslations } from 'next-intl/server'
import Image from 'next/image'
import Link from 'next/link'

const seasons = [
  {
    seasonNumber: 3,
    startDate: '2025-04-01',
    endDate: '2025-06-10',
    image: 's3://lania-web-134312503254-eu-central-1-an/2025-08-19_11.15.29.png',
  },
  {
    seasonNumber: 2,
    startDate: '2024-07-12',
    endDate: '2024-09-03',
    image: 's3://lania-web-134312503254-eu-central-1-an/2025-08-29_21.23.23.png',
  },
  {
    seasonNumber: 1,
    startDate: '2023-12-23',
    endDate: '2024-02-23',
    image: 's3://lania-web-134312503254-eu-central-1-an/2025-08-29_21.26.21.png',
  },
]

type Props = {
  params: Promise<{ locale: string }>
}

export default async function SeasonsPage({ params }: Props) {
  const { locale } = await params
  const t = await getTranslations({ locale, namespace: 'seasons' })

  return (
    <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
      {seasons.map((season) => {
        return (
          <div
            key={season.seasonNumber}
            className="bg-card flex max-w-sm flex-col items-start justify-between gap-2 justify-self-center rounded-md p-6 shadow-md"
          >
            <div className="flex flex-col gap-4">
              <AspectRatio ratio={16 / 9}>
                <Image
                  src={season.image}
                  alt="Season 1"
                  fill
                  className="rounded-md object-cover transition-transform duration-300 hover:scale-105"
                />
              </AspectRatio>
              <h2 className="-translate-y-0.5 text-2xl font-bold">
                {t(`${season.seasonNumber}.title`)}
              </h2>
              <div>
                <h3 className="text-sm">
                  {new Date(season.startDate).toLocaleDateString(locale, {
                    day: 'numeric',
                    month: 'long',
                    year: 'numeric',
                  })}
                  &nbsp;&mdash;&nbsp;
                  {new Date(season.endDate).toLocaleDateString(locale, {
                    day: 'numeric',
                    month: 'long',
                    year: 'numeric',
                  })}
                </h3>
                <div className="flex items-center gap-2 text-sm">
                  <ClockIcon className="text-muted-foreground size-3" />
                  <p>{t(`${season.seasonNumber}.length`)}</p>
                </div>
              </div>
              <p className="text-muted-foreground text-sm">
                {t(`${season.seasonNumber}.description`)}
              </p>
            </div>
            <Button variant="link" className="hidden h-auto p-0" asChild>
              <Link href={`/${locale}/seasons/${season.seasonNumber}`}>
                {t('readMore')}
              </Link>
            </Button>
          </div>
        )
      })}
    </div>
  )
}
