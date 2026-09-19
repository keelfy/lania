import type { Metadata } from 'next'
import { notFound } from 'next/navigation'
import { hasLocale } from 'next-intl'
import { routing } from '@/i18n/routing'
import { AdminShell } from '@/components/admin/shell'
import { AdminIssue } from '@/components/admin/common'
import { adminSession } from '@/lib/admin/server'
import { Toaster } from '@/components/ui/sonner'
import '../(app)/globals.css'

export const metadata: Metadata = { title: 'Управление · Lania', robots: { index: false, follow: false } }

export default async function AdminLayout({children,params}:{children:React.ReactNode;params:Promise<{locale:string}>}) {
 const {locale}=await params
 if(!hasLocale(routing.locales,locale)) notFound()
 const base=`/${locale}/admin`
 let content=children
 try { await adminSession() } catch(error) { content=<AdminIssue error={error} base={base}/> }
 return <html lang="ru"><body className="font-sans antialiased"><AdminShell base={base}>{content}</AdminShell><Toaster/></body></html>
}
