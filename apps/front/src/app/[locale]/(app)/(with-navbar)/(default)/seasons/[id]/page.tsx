import { AspectRatio } from '@/components/ui/aspect-ratio'
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from '@/components/ui/breadcrumb'
import { getSeasons } from '@/lib/api-endpoints'
import { serverApiFetcher } from '@/lib/server'
import Image from 'next/image'
import { notFound } from 'next/navigation'

type Props = {
  params: Promise<{ locale: string; id: string }>
}

export default async function SeasonPage({ params }: Props) {
  const { id } = await params

  const seasons = await getSeasons(serverApiFetcher)
  const season = seasons.find((season) => season.id === id)
  if (!season) notFound()

  return (
    <div>
      <Breadcrumb>
        <BreadcrumbList>
          <BreadcrumbItem>
            <BreadcrumbLink href="/seasons">Архив сезонов</BreadcrumbLink>
          </BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>
            <BreadcrumbPage>{season.name}</BreadcrumbPage>
          </BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
      <h1 className="mb-6 text-center text-4xl font-bold">{season.name}</h1>
      <div className="flex flex-col gap-4">
        {season.previewImage && (
          <AspectRatio ratio={16 / 9}>
            <Image
              src={season.previewImage}
              alt={season.name}
              fill
              className="rounded-md object-cover transition-transform duration-300 hover:scale-105"
            />
          </AspectRatio>
        )}
      </div>
    </div>
  )
}
