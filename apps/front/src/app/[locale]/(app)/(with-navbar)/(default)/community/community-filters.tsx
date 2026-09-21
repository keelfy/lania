'use client'

import { Button } from '@/components/ui/button'
import { RadioIcon, ShieldIcon } from 'lucide-react'
import { useTranslations } from 'next-intl'
import { useRouter } from 'next/navigation'
import React from 'react'
import { CommunityFilters as Filters, communityHref } from './community-href'

type Props = Filters & {
  sort: string
  search: string
  locale: string
  // The online filter is hidden for a season that has no running server.
  onlineAvailable: boolean
}

export default function CommunityFilters({
  sort,
  search,
  locale,
  online,
  staff,
  season,
  onlineAvailable,
}: Props) {
  const t = useTranslations('community.filters')
  const router = useRouter()
  const [isPending, startTransition] = React.useTransition()

  const toggle = (filters: Filters) => {
    startTransition(() => {
      router.push(
        communityHref({
          locale,
          sort,
          search,
          online,
          staff,
          season,
          ...filters,
        }),
      )
    })
  }

  return (
    <div className="flex flex-wrap items-center gap-2">
      {onlineAvailable && (
        <Button
          variant={online ? 'default' : 'outline'}
          size="sm"
          aria-pressed={!!online}
          disabled={isPending}
          onClick={() => toggle({ online: !online })}
        >
          <RadioIcon className="size-4" />
          {t('online')}
        </Button>
      )}
      <Button
        variant={staff ? 'default' : 'outline'}
        size="sm"
        aria-pressed={!!staff}
        disabled={isPending}
        onClick={() => toggle({ staff: !staff })}
      >
        <ShieldIcon className="size-4" />
        {t('staff')}
      </Button>
    </div>
  )
}
