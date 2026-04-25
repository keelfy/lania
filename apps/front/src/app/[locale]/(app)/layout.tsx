import CookieConsent from '@/components/blocks/cookie-consent'
import { BasketProvider } from '@/context/basket'
import { routing } from '@/i18n/routing'
import { getBasket } from '@/lib/api-endpoints'
import { getCurrentSession } from '@/lib/get-current-session'
import { serverApiFetcher } from '@/lib/server'
import AuthStoreProvider from '@/providers/auth-store'
import ModalStoreProvider from '@/providers/modal'
import type { Metadata } from 'next'
import { hasLocale, NextIntlClientProvider } from 'next-intl'
import { getTranslations, setRequestLocale } from 'next-intl/server'
import { ThemeProvider } from 'next-themes'
import { Geist, Geist_Mono } from 'next/font/google'
import { notFound } from 'next/navigation'
import { NuqsAdapter } from 'nuqs/adapters/next/app'
import React, { Suspense } from 'react'
import { Toaster } from 'sonner'
import './globals.css'

const geistSans = Geist({
  variable: '--font-geist-sans',
  subsets: ['latin'],
})

const geistMono = Geist_Mono({
  variable: '--font-geist-mono',
  subsets: ['latin'],
})

type Props = {
  params: Promise<{
    locale: string
  }>
}

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { locale } = await params
  const t = await getTranslations({ locale, namespace: 'metadata' })

  return {
    title: t('title'),
    description: t('description'),
    openGraph: {
      type: 'website',
      url: 'https://lania.gg',
      title: t('title'),
      description: t('description'),
      siteName: 'Lania Network',
    },
  }
}

export default async function RootLayout({
  children,
  params,
}: Readonly<React.PropsWithChildren<Props>>) {
  const { locale } = await params
  const session = await getCurrentSession()
  const basket = await getBasket(serverApiFetcher).catch((err) => {
    console.error(err)
    return []
  })

  if (!hasLocale(routing.locales, locale)) {
    notFound()
  }

  // Enable static rendering
  setRequestLocale(locale)

  return (
    <html lang={locale} dir="ltr" suppressHydrationWarning>
      <body
        className={`${geistSans.variable} ${geistMono.variable} antialiased`}
      >
        <ThemeProvider attribute="class" defaultTheme="dark" enableSystem>
          <Suspense>
            <NextIntlClientProvider locale={locale}>
              <NuqsAdapter>
                <ModalStoreProvider>
                  <AuthStoreProvider session={session}>
                    <BasketProvider initialBasket={basket}>
                      {children}
                    </BasketProvider>
                  </AuthStoreProvider>
                  <Toaster />
                </ModalStoreProvider>
              </NuqsAdapter>
              <CookieConsent variant="small" />
            </NextIntlClientProvider>
          </Suspense>
        </ThemeProvider>
      </body>
    </html>
  )
}
