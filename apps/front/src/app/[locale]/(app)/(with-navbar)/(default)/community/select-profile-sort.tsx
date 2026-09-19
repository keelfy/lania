'use client'

import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  ArrowDownAzIcon,
  ArrowDownNarrowWideIcon,
  ArrowUpAzIcon,
  ArrowUpNarrowWideIcon,
} from 'lucide-react'
import { useTranslations } from 'next-intl'
import { useRouter } from 'next/navigation'
import React from 'react'
import { communityHref, DEFAULT_COMMUNITY_SORT } from './community-href'

type Props = {
  defaultValue?: string
  search?: string
  locale: string
}

const SORT_OPTIONS = [
  {
    value: 'created_at.asc',
    labelKey: 'createdAtAsc',
    icon: ArrowUpNarrowWideIcon,
  },
  {
    value: 'created_at.desc',
    labelKey: 'createdAtDesc',
    icon: ArrowDownNarrowWideIcon,
  },
  {
    value: 'username.asc',
    labelKey: 'usernameAsc',
    icon: ArrowUpAzIcon,
  },
  {
    value: 'username.desc',
    labelKey: 'usernameDesc',
    icon: ArrowDownAzIcon,
  },
  {
    value: 'first_seen_at.asc',
    labelKey: 'firstSeenAtAsc',
    icon: ArrowUpNarrowWideIcon,
  },
  {
    value: 'first_seen_at.desc',
    labelKey: 'firstSeenAtDesc',
    icon: ArrowDownNarrowWideIcon,
  },
  {
    value: 'last_seen_at.desc',
    labelKey: 'lastSeenAtDesc',
    icon: ArrowDownNarrowWideIcon,
  },
  {
    value: 'last_seen_at.asc',
    labelKey: 'lastSeenAtAsc',
    icon: ArrowUpNarrowWideIcon,
  },
  {
    value: 'playtime.desc',
    labelKey: 'playtimeDesc',
    icon: ArrowDownNarrowWideIcon,
  },
  {
    value: 'playtime.asc',
    labelKey: 'playtimeAsc',
    icon: ArrowUpNarrowWideIcon,
  },
] as const

export default function SelectCommunitySort({
  defaultValue,
  search,
  locale,
}: Props) {
  const t = useTranslations('community.sort')
  const [sort, setSort] = React.useState(defaultValue ?? DEFAULT_COMMUNITY_SORT)
  const [isSortChanging, startSortChange] = React.useTransition()
  const router = useRouter()

  const onSortChange = (value: string) => {
    setSort(value)
    startSortChange(() => {
      router.push(communityHref({ locale, sort: value, search }))
    })
  }

  return (
    <Select value={sort} onValueChange={onSortChange} disabled={isSortChanging}>
      <SelectTrigger>
        <SelectValue placeholder={t('placeholder')} />
      </SelectTrigger>
      <SelectContent>
        {SORT_OPTIONS.map((option) => (
          <SelectItem key={option.value} value={option.value}>
            <option.icon className="size-4" />
            {t(option.labelKey)}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  )
}
