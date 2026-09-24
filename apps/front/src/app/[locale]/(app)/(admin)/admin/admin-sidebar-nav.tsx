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
          <div key={item.href} className="flex flex-col gap-1">
            <Link
              href={`/${locale}${item.children?.[0].href ?? item.href}`}
              onClick={onNavigate}
              className={cn(
                'flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors',
                isActive && !item.children
                  ? 'bg-accent text-accent-foreground'
                  : isActive
                    ? 'text-foreground'
                    : 'text-muted-foreground hover:bg-accent/50 hover:text-foreground',
              )}
            >
              <item.icon className="size-4" />
              {t(item.labelKey)}
            </Link>
            {item.children && (
              <div className="ml-5 flex flex-col gap-1 border-l pl-2">
                {item.children.map((child) => (
                  <Link
                    key={child.href}
                    href={`/${locale}${child.href}`}
                    onClick={onNavigate}
                    className={cn(
                      'rounded-md px-3 py-1.5 text-sm transition-colors',
                      pathnameStartsWith(pathname, child.href)
                        ? 'bg-accent text-accent-foreground font-medium'
                        : 'text-muted-foreground hover:bg-accent/50 hover:text-foreground',
                    )}
                  >
                    {t(child.labelKey)}
                  </Link>
                ))}
              </div>
            )}
          </div>
        )
      })}
    </nav>
  )
}
