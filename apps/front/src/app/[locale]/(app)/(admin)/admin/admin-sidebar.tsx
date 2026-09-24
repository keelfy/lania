import DeerIcon from '@/components/icons/DeerIcon'
import { Badge } from '@/components/ui/badge'
import { Separator } from '@/components/ui/separator'
import { cn } from '@/lib/utils'
import { ArrowLeftIcon } from 'lucide-react'
import { getTranslations } from 'next-intl/server'
import Link from 'next/link'
import LanguageDropdownMenu from '../../components/language-dropdown-menu'
import UserDropdownMenu from '../../components/user-dropdown'
import AdminSidebarNav from './admin-sidebar-nav'

type Props = {
  locale: string
  className?: string
}

// Desktop-only: the base classes intentionally leave `display` unset so the
// caller's `hidden lg:flex` decides visibility without fighting this file's classes.
export default async function AdminSidebar({ locale, className }: Props) {
  const t = await getTranslations({ locale, namespace: 'admin.sidebar' })

  return (
    <aside
      className={cn(
        'bg-background sticky top-0 h-svh w-64 shrink-0 flex-col border-r',
        className,
      )}
    >
      <Link href={`/${locale}`} className="flex items-center gap-2 px-4 py-4">
        <DeerIcon className="size-7 shrink-0" />
        <span className="text-base font-bold tracking-widest uppercase">
          Lania
        </span>
        <Badge variant="secondary">{t('badge')}</Badge>
      </Link>
      <Separator />
      <AdminSidebarNav locale={locale} className="flex-1 px-3 py-4" />
      <Separator />
      <div className="flex flex-col gap-3 px-4 py-4">
        <Link
          href={`/${locale}`}
          className="text-muted-foreground hover:text-foreground flex items-center gap-2 text-sm transition-colors"
        >
          <ArrowLeftIcon className="size-4" />
          {t('backToSite')}
        </Link>
        <div className="flex items-center gap-2">
          <UserDropdownMenu />
          <LanguageDropdownMenu currentLocale={locale} />
        </div>
      </div>
    </aside>
  )
}
