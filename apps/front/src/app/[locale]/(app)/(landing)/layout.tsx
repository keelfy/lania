import SwirlBackground from '@/components/SwirlBackground'
import { Locale } from '@/lib/locale'
import React from 'react'
import Footer from '../components/footer'
import Navbar from '../components/navbar'
import { LandingNavigationProvider } from './components/landing-navigation-ctx'

type Props = {
  params: Promise<{
    locale: string
  }>
}

export default async function LandingLayout({
  children,
  params,
}: React.PropsWithChildren<Props>) {
  const { locale } = await params
  return (
    <div className="relative min-h-screen">
      <Navbar
        id="header"
        className="pointer-events-auto fixed inset-x-0 top-0 z-50 w-full"
        currentLocale={locale as Locale}
      />
      <SwirlBackground />
      <main
        id="landing-scroll-root"
        className="flex flex-1 snap-y snap-mandatory overflow-y-auto scroll-smooth"
      >
        <LandingNavigationProvider>{children}</LandingNavigationProvider>
      </main>
      <Footer locale={locale} />
    </div>
  )
}
