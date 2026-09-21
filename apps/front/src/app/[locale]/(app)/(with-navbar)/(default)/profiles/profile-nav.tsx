'use client'

import { Button } from '@/components/ui/button'
import { Profile } from '@/models/profile'
import { ArrowUpRightIcon } from 'lucide-react'
import { useLocale, useTranslations } from 'next-intl'
import Link from 'next/link'
import { usePathname, useSearchParams } from 'next/navigation'
import { useSelectedProfile } from './use-selected-profile'

const SECTIONS = [
  { key: 'overview', path: '' },
  { key: 'settings', path: '/settings' },
] as const

type Props = {
  profiles: Profile[]
}

// The tabs of the section. The profile and the season stay in the query when the tab changes.
export default function ProfileNav({ profiles }: Props) {
  const t = useTranslations('profiles.nav')
  const locale = useLocale()
  const profile = useSelectedProfile(profiles)
  const pathname = usePathname()
  const query = useSearchParams().toString()

  const base = pathname.replace(/\/settings$/, '')
  const active = pathname === base ? 'overview' : 'settings'

  return (
    <nav className="flex flex-wrap items-center justify-between gap-2">
      <div className="flex gap-2">
        {SECTIONS.map(({ key, path }) => (
          <Button
            key={key}
            asChild
            variant={key === active ? 'default' : 'outline'}
            size="sm"
          >
            <Link href={`${base}${path}${query && `?${query}`}`}>{t(key)}</Link>
          </Button>
        ))}
      </div>
      {profile && (
        <Button asChild variant="ghost" size="sm">
          <Link href={`/${locale}/community/${profile.username}`}>
            {t('viewPublic')}
            <ArrowUpRightIcon />
          </Link>
        </Button>
      )}
    </nav>
  )
}
