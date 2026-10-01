'use client'

import { Input } from '@/components/ui/input'
import { useDebouncedState } from '@/lib/use-debounced-state'
import { useTranslations } from 'next-intl'
import { useRouter } from 'next/navigation'
import React from 'react'
import { LoaderCircleIcon, SearchIcon } from 'lucide-react'
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
  const [lastSubmitted, setLastSubmitted] = React.useState(defaultValue ?? '')
  const [previousDefault, setPreviousDefault] = React.useState(defaultValue)
  const [isPending, startTransition] = React.useTransition()
  // Keep browser history and external resets in sync without remounting the focused input.
  if (previousDefault !== defaultValue) {
    setPreviousDefault(defaultValue)
    if (defaultValue !== lastSubmitted) setValue(defaultValue ?? '')
  }
  const search = useDebouncedState(value.trim(), 300)

  React.useEffect(() => {
    if (search !== value.trim() || search === (defaultValue ?? '')) return
    startTransition(() => {
      setLastSubmitted(search)
      router.replace(
        communityHref({ locale, sort, search, online, staff, season }),
        { scroll: false },
      )
    })
  }, [
    search,
    value,
    defaultValue,
    sort,
    locale,
    online,
    staff,
    season,
    router,
    startTransition,
  ])

  return (
    <div
      className="relative min-w-48 flex-1 basis-full sm:basis-56"
      aria-busy={isPending}
    >
      <SearchIcon
        aria-hidden
        className="text-muted-foreground pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2"
      />
      <Input
        type="search"
        value={value}
        onChange={(e) => setValue(e.target.value)}
        maxLength={MAX_SEARCH_LENGTH}
        placeholder={t('searchPlaceholder')}
        aria-label={t('searchPlaceholder')}
        className="w-full pr-9 pl-9"
      />
      {(isPending || value.trim() !== (defaultValue ?? '')) && (
        <span
          role="status"
          className="text-muted-foreground pointer-events-none absolute top-1/2 right-3 -translate-y-1/2"
        >
          <LoaderCircleIcon
            aria-hidden
            className="size-4 animate-spin motion-reduce:animate-none"
          />
          <span className="sr-only">{t('searching')}</span>
        </span>
      )}
    </div>
  )
}
