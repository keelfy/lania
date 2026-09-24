import DeerIcon from '@/components/icons/DeerIcon'
import { Badge } from '@/components/ui/badge'
import { requireAdmin } from '@/lib/admin'
import type { Metadata } from 'next'
import { getTranslations } from 'next-intl/server'
import Link from 'next/link'
import React from 'react'
import AdminMobileNav from './admin-mobile-nav'
import AdminSidebar from './admin-sidebar'

type Props = {
  children: React.ReactNode
  params: Promise<{ locale: string }>
}

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { locale } = await params
  const t = await getTranslations({ locale, namespace: 'admin' })
  return {
    title: t('metadata.title'),
    robots: { index: false, follow: false },
  }
}

// Every admin page still calls requireAdmin() itself: this layout does not
// re-render on client-side navigation between admin pages, so on its own it
// would not re-guard a section change (see lib/admin.ts).
export default async function AdminLayout({ children, params }: Props) {
  await requireAdmin()
  const { locale } = await params
  const t = await getTranslations({ locale, namespace: 'admin.sidebar' })

  return (
    <div className="flex min-h-svh">
      <AdminSidebar locale={locale} className="hidden lg:flex" />
      <div className="flex min-h-svh flex-1 flex-col">
        <div className="flex items-center justify-between border-b px-4 py-3 lg:hidden">
          <Link href={`/${locale}`} className="flex items-center gap-2">
            <DeerIcon className="size-6 shrink-0" />
            <span className="text-sm font-bold tracking-widest uppercase">
              Lania
            </span>
            <Badge variant="secondary">{t('badge')}</Badge>
          </Link>
          <AdminMobileNav locale={locale} />
        </div>
        <main className="flex-1 px-4 py-6 lg:px-8 lg:py-8">
          <div className="mx-auto w-full max-w-7xl">{children}</div>
        </main>
      </div>
    </div>
  )
}
