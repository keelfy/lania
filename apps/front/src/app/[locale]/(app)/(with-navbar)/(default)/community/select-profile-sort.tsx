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
import { useRouter } from 'next/navigation'
import React from 'react'

type Props = {
  defaultValue?: string
  locale: string
}

const SORT_OPTIONS = [
  {
    value: 'created_at.asc',
    label: 'Старые',
    icon: ArrowUpNarrowWideIcon,
  },
  {
    value: 'created_at.desc',
    label: 'Новые',
    icon: ArrowDownNarrowWideIcon,
  },
  {
    value: 'username.asc',
    label: 'А-я',
    icon: ArrowUpAzIcon,
  },
  {
    value: 'username.desc',
    label: 'Я-а',
    icon: ArrowDownAzIcon,
  },
]

export default function SelectCommunitySort({ defaultValue, locale }: Props) {
  const [sort, setSort] = React.useState(defaultValue ?? 'created_at.asc')
  const [isSortChanging, startSortChange] = React.useTransition()
  const router = useRouter()

  React.useEffect(() => {
    if (sort === defaultValue) return
    startSortChange(async () => {
      router.push(`/${locale}/community?sort=${sort}&page=0`)
    })
  }, [sort, locale, router, defaultValue])

  return (
    <Select
      defaultValue={sort}
      onValueChange={setSort}
      disabled={isSortChanging}
    >
      <SelectTrigger>
        <SelectValue placeholder="Сортировать по" />
      </SelectTrigger>
      <SelectContent>
        {SORT_OPTIONS.map((option) => (
          <SelectItem key={option.value} value={option.value}>
            <option.icon className="size-4" />
            {option.label}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  )
}
