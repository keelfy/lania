import { Button } from '@/components/ui/button'
import { getTranslations } from 'next-intl/server'
import Link from 'next/link'
import React from 'react'

type Props = React.PropsWithChildren<{
  locale: string
  active: 'users' | 'profiles' | 'seasons' | 'cosmetics' | 'products'
}>

// Every admin page renders its own shell: a layout does not render again on navigation, so it cannot guard the section.
export default async function AdminShell({ locale, active, children }: Props) {
  const t = await getTranslations({ locale, namespace: 'admin' })

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <h1 className="text-4xl font-extrabold tracking-tight">{t('title')}</h1>
        <nav className="flex gap-2">
          {(
            ['users', 'profiles', 'seasons', 'cosmetics', 'products'] as const
          ).map((section) => (
            <Button
              key={section}
              asChild
              variant={section === active ? 'default' : 'outline'}
              size="sm"
            >
              <Link href={`/${locale}/admin/${section}`}>
                {t(`nav.${section}`)}
              </Link>
            </Button>
          ))}
        </nav>
      </div>
      {children}
    </div>
  )
}
