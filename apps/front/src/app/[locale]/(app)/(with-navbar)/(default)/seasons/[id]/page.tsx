import { AspectRatio } from '@/components/ui/aspect-ratio'
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from '@/components/ui/breadcrumb'
import { seasonNumbers } from '@/lib/season-number'
import Image from 'next/image'

type Props = {
  params: Promise<{ locale: string; id: string }>
}

const season = {
  seasonNumber: 1,
  startDate: '2023-12-23',
  endDate: '2024-02-23',
  description: 'Первый сезон был запущен для игры со зрителями keelfy.',
  image: '/screenshot-1.jpg',
}

export default async function SeasonPage({ params }: Props) {
  const { id, locale } = await params

  return (
    <div>
      <Breadcrumb>
        <BreadcrumbList>
          <BreadcrumbItem>
            <BreadcrumbLink href="/seasons">Архив сезонов</BreadcrumbLink>
          </BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>
            <BreadcrumbPage>
              ʟᴀɴɪᴀ {seasonNumbers[Number(id)]} {locale}
            </BreadcrumbPage>
          </BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
      <h1 className="mb-6 text-center text-4xl font-bold">
        ʟᴀɴɪᴀ &mdash; Сезон {seasonNumbers[Number(id)]}
      </h1>
      <div className="flex flex-col gap-4">
        <AspectRatio ratio={16 / 9}>
          <Image
            src={season.image}
            alt={`Lania ${seasonNumbers[Number(id)]}`}
            fill
            className="rounded-md object-cover transition-transform duration-300 hover:scale-105"
            unoptimized
          />
        </AspectRatio>
      </div>
    </div>
  )
}
