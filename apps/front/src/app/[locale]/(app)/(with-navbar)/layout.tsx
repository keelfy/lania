import { Locale } from '@/lib/locale'
import React from 'react'
import Footer from '../components/footer'
import Navbar from '../components/navbar'

type Props = {
  params: Promise<{
    locale: string
  }>
}

export default async function DefaultLayout({
  children,
  params,
}: React.PropsWithChildren<Props>) {
  const { locale } = await params
  return (
    <div className="space-y-8">
      <div className="min-h-svh space-y-8">
        <Navbar currentLocale={locale as Locale} />
        {children}
      </div>
      <Footer locale={locale} />
    </div>
  )
}
