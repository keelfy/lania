import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from '@/components/ui/breadcrumb'
import { Locale } from '@/lib/locale'
import { MapIcon } from 'lucide-react'
import { Metadata } from 'next'
import { getTranslations } from 'next-intl/server'
import Footer from '../../../components/footer'
import Navbar from '../../../components/navbar'

type Props = {
  params: Promise<{
    world: string
    locale: string
  }>
}

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { world, locale } = await params
  const t = await getTranslations({ locale, namespace: 'worlds.maps' })

  const mapTitle = t('mapNames.' + world)
  return {
    title: t('metadata.title').replace('<world/>', mapTitle),
    description: t('metadata.description').replace('<world/>', mapTitle),
    openGraph: {
      type: 'website',
      url: 'https://lania.gg/worlds',
      title: t('metadata.title').replace('<world/>', mapTitle),
      description: t('metadata.description').replace('<world/>', mapTitle),
      siteName: 'Lania Network',
    },
  }
}

export default async function MapsLayout({
  children,
  params,
}: React.PropsWithChildren<Props>) {
  const { world, locale } = await params
  const t = await getTranslations({ locale, namespace: 'worlds.maps' })
  return (
    <>
      <div className="min-h-svh">
        <Navbar
          id="header"
          className="w-full"
          currentLocale={locale as Locale}
        />
        <main className="w-full flex-1 flex-col">
          <div className="mx-auto flex h-8 max-w-5xl justify-center">
            <Breadcrumb>
              <BreadcrumbList>
                <BreadcrumbItem>
                  <BreadcrumbLink
                    href="/worlds"
                    className="flex items-center gap-2"
                  >
                    <MapIcon className="size-3" />
                    {t('breadcrumbs.title')}
                  </BreadcrumbLink>
                </BreadcrumbItem>
                <BreadcrumbSeparator />
                <BreadcrumbItem>
                  <BreadcrumbPage>{t('mapNames.' + world)}</BreadcrumbPage>
                </BreadcrumbItem>
              </BreadcrumbList>
            </Breadcrumb>
          </div>
          <div className="h-[calc(100vh-96px)]">{children}</div>
        </main>
      </div>
      <Footer locale={locale} />
    </>
  )
}
