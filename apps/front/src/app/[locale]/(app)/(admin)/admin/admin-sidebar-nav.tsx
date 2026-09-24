'use client'

import { cn } from '@/lib/utils'
import { pathnameStartsWith, usePathname } from '@/i18n/navigation'
import { useTranslations } from 'next-intl'
import Link from 'next/link'
import { ADMIN_NAV_ITEMS } from './admin-nav-items'

type Props = {
  locale: string
  className?: string
  // Closes the mobile sheet after a link is picked; unused on the desktop sidebar.
  onNavigate?: () => void
}

export default function AdminSidebarNav({
  locale,
  className,
  onNavigate,
}: Props) {
  const t = useTranslations('admin.nav')
  const pathname = usePathname()

  return (
    <nav className={cn('flex flex-col gap-1', className)}>
      {ADMIN_NAV_ITEMS.map((item) => {
        const isActive = pathnameStartsWith(pathname, item.href)
        return (
          <Link
            key={item.href}
            href={`/${locale}${item.href}`}
            onClick={onNavigate}
            className={cn(
              'flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors',
              isActive
                ? 'bg-accent text-accent-foreground'
                : 'text-muted-foreground hover:bg-accent/50 hover:text-foreground',
            )}
          >
            <item.icon className="size-4" />
            {t(item.labelKey)}
          </Link>
        )
      })}
    </nav>
  )
}
