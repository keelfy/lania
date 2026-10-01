import CookieConsent from '@/components/blocks/cookie-consent'
import { BasketProvider } from '@/context/basket'
import { getMetadataLocale } from '@/i18n/metadata-locale'
import { routing } from '@/i18n/routing'
import AuthStoreProvider from '@/providers/auth-store'
import type { Metadata } from 'next'
import { hasLocale, NextIntlClientProvider } from 'next-intl'
import { getTranslations, setRequestLocale } from 'next-intl/server'
import { ThemeProvider } from 'next-themes'
import { Geist, Geist_Mono, Noto_Sans } from 'next/font/google'
import { notFound } from 'next/navigation'
import { NuqsAdapter } from 'nuqs/adapters/next/app'
import React, { Suspense } from 'react'
import { Toaster } from 'sonner'
import './globals.css'
import './typeset.css'
import Viewer from './components/viewer'

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

export function generateStaticParams() {
  return routing.locales.map((locale) => ({ locale }))
}

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const locale = await getMetadataLocale(params)
  const t = await getTranslations({ locale, namespace: 'metadata' })

  return {
    title: t('title'),
    description: t('description'),
    openGraph: {
      type: 'website',
      url: 'https://lania.network',
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
                <AuthStoreProvider>
                  <BasketProvider>
                    <Suspense>
                      <Viewer />
                    </Suspense>
                    {children}
                  </BasketProvider>
                </AuthStoreProvider>
                <Toaster />
              </NuqsAdapter>
              <CookieConsent variant="small" />
            </NextIntlClientProvider>
          </Suspense>
        </ThemeProvider>
      </body>
    </html>
  )
}
