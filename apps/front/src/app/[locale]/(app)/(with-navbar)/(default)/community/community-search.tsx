'use client'

import { Input } from '@/components/ui/input'
import { useDebouncedState } from '@/lib/use-debounced-state'
import { useTranslations } from 'next-intl'
import { useRouter } from 'next/navigation'
import React from 'react'
import { CommunityFilters, communityHref } from './community-href'

type Props = CommunityFilters & {
  defaultValue?: string
  sort?: string
  locale: string
}

// Longest Minecraft username.
const MAX_SEARCH_LENGTH = 16

export default function CommunitySearch({
  defaultValue,
  sort,
  locale,
  online,
  staff,
  season,
}: Props) {
  const t = useTranslations('community')
  const router = useRouter()
  const [value, setValue] = React.useState(defaultValue ?? '')
  const search = useDebouncedState(value.trim(), 300)

  React.useEffect(() => {
    if (search === (defaultValue ?? '')) return
    router.replace(
      communityHref({ locale, sort, search, online, staff, season }),
    )
  }, [search, defaultValue, sort, locale, online, staff, season, router])

  return (
    <Input
      type="search"
      value={value}
      onChange={(e) => setValue(e.target.value)}
      maxLength={MAX_SEARCH_LENGTH}
      placeholder={t('searchPlaceholder')}
      aria-label={t('searchPlaceholder')}
      className="w-full sm:w-48"
    />
  )
}
