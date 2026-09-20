import { AspectRatio } from '@/components/ui/aspect-ratio'
import { Button } from '@/components/ui/button'
import { ClockIcon } from 'lucide-react'
import { getTranslations } from 'next-intl/server'
import Image from 'next/image'
import { Noto_Sans } from 'next/font/google'
import Link from 'next/link'

const notoSans = Noto_Sans({
  subsets: ['latin'],
})

const seasons = [
  {
    seasonNumber: 5,
    name: "LANIA V",
    startDate: '2026-10-09',
    endDate: '-',
    image: 's3://lania-web-134312503254-eu-central-1-an/lania-5-preview.jpg',
  },
  {
    seasonNumber: 10,
    name: "LANIA SPINOFF I",
    startDate: '2026-04-26',
    endDate: '-',
    image: 's3://lania-web-134312503254-eu-central-1-an/lania-spinoff-1-preview.png',
  },
  {
    seasonNumber: 4,
    name: "LANIA IV",
    startDate: '2025-09-26',
    endDate: '2025-12-10',
    image: 's3://lania-web-134312503254-eu-central-1-an/lania-4-castle.png',
  },
  {
    seasonNumber: 3,
    name: "LANIA III",
    startDate: '2025-04-01',
    endDate: '2025-06-10',
    image: 's3://lania-web-134312503254-eu-central-1-an/2025-08-19_11.15.29.png',
  },
  {
    seasonNumber: 2,
    name: "LANIA II",
    startDate: '2024-07-12',
    endDate: '2024-09-03',
    image: 's3://lania-web-134312503254-eu-central-1-an/2025-08-29_21.23.23.png',
  },
  {
    seasonNumber: 1,
    name: "LANIA I",
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

  const toLocalDate = (date: string) => {
    const d = new Date(date);
    if (isNaN(Date.parse(date))) return '?';
    return d.toLocaleDateString(locale, {
                    day: 'numeric',
                    month: 'long',
                    year: 'numeric',
                  })
  }

  return (
    <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
      {seasons.map((season) => {
        return (
          <div
            key={season.seasonNumber}
            className="bg-card w-full flex max-w-sm flex-col items-start justify-between gap-2 justify-self-center rounded-md p-6 shadow-md"
          >
            <div className="flex flex-col gap-4">
              <AspectRatio ratio={16 / 9}>
                <Image
                  src={season.image}
                  alt={season.name}
                  width={690}
                  height={480}
                  className="rounded-md object-cover transition-transform duration-300 hover:scale-105"
                />
              </AspectRatio>
              <h2 className={`text-xl font-bold ${notoSans.className} antialiased`}>
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
