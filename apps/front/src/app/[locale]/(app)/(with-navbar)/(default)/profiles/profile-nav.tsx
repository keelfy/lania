'use client'

import { Button } from '@/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Profile } from '@/models/profile'
import { ArrowUpRightIcon } from 'lucide-react'
import { useLocale, useTranslations } from 'next-intl'
import Link from 'next/link'
import { usePathname, useRouter, useSearchParams } from 'next/navigation'
import { useSelectedProfile } from './use-selected-profile'

const SECTIONS = [
  { key: 'overview', path: '' },
  { key: 'cosmetics', path: '/cosmetics' },
  { key: 'settings', path: '/settings' },
] as const

type Section = (typeof SECTIONS)[number]['key']

type Props = {
  profiles: Profile[]
}

// The tabs of the section. The profile and the season stay in the query when the tab changes.
// A phone gets a select instead, so the tabs do not wrap.
export default function ProfileNav({ profiles }: Props) {
  const t = useTranslations('profiles.nav')
  const locale = useLocale()
  const router = useRouter()
  const profile = useSelectedProfile(profiles)
  const pathname = usePathname()
  const query = useSearchParams().toString()

  const base = pathname.replace(/\/(cosmetics|settings)$/, '')
  const active =
    SECTIONS.find(({ path }) => path && pathname.endsWith(path))?.key ??
    'overview'
  const hrefOf = (key: Section) =>
    `${base}${SECTIONS.find((section) => section.key === key)!.path}${query && `?${query}`}`

  return (
    <nav className="flex items-center justify-between gap-2">
      <Select
        value={active}
        onValueChange={(key) => router.push(hrefOf(key as Section))}
      >
        <SelectTrigger className="min-w-0 flex-1 sm:hidden">
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          {SECTIONS.map(({ key }) => (
            <SelectItem key={key} value={key}>
              {t(key)}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
      {/* Every tab is its own page, so the triggers are links and no panel is rendered here. */}
      <Tabs value={active} className="hidden sm:flex">
        <TabsList>
          {SECTIONS.map(({ key }) => (
            <TabsTrigger key={key} value={key} asChild className="px-3">
              <Link href={hrefOf(key)}>{t(key)}</Link>
            </TabsTrigger>
          ))}
        </TabsList>
      </Tabs>
      {profile && (
        <Button asChild variant="ghost" size="sm" className="shrink-0">
          <Link href={`/${locale}/community/${profile.username}`}>
            {t('viewPublic')}
            <ArrowUpRightIcon />
          </Link>
        </Button>
      )}
    </nav>
  )
}
