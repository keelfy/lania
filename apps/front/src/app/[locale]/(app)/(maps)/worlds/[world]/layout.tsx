import { Locale } from '@/lib/locale'
import Footer from '../../../components/footer'
import Navbar from '../../../components/navbar'

type Props = {
  params: Promise<{
    world: string
    locale: string
  }>
}

// The map takes the whole screen under the navbar; the footer is past it.
export default async function MapsLayout({
  children,
  params,
}: React.PropsWithChildren<Props>) {
  const { locale } = await params
  return (
    <>
      <Navbar id="header" className="w-full" currentLocale={locale as Locale} />
      <main className="h-[calc(100svh-4rem)] w-full">{children}</main>
      <Footer locale={locale} />
    </>
  )
}
