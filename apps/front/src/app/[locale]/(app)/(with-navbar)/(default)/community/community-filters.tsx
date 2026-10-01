'use client'

import { Button } from '@/components/ui/button'
import { LoaderCircleIcon, RadioIcon, ShieldIcon } from 'lucide-react'
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
        { scroll: false },
      )
    })
  }

  return (
    <div className="flex flex-wrap items-center gap-1" aria-busy={isPending}>
      {onlineAvailable && (
        <Button
          variant={online ? 'secondary' : 'ghost'}
          className={
            online
              ? 'ring-primary/25 text-foreground ring-1'
              : 'text-muted-foreground'
          }
          aria-pressed={!!online}
          disabled={isPending}
          onClick={() => toggle({ online: !online })}
        >
          <RadioIcon
            aria-hidden
            className={online ? 'size-4 text-emerald-500' : 'size-4'}
          />
          {t('online')}
        </Button>
      )}
      <Button
        variant={staff ? 'secondary' : 'ghost'}
        className={
          staff
            ? 'ring-primary/25 text-foreground ring-1'
            : 'text-muted-foreground'
        }
        aria-pressed={!!staff}
        disabled={isPending}
        onClick={() => toggle({ staff: !staff })}
      >
        <ShieldIcon aria-hidden className="size-4" />
        {t('staff')}
      </Button>
      {isPending && (
        <LoaderCircleIcon
          aria-hidden
          className="text-muted-foreground size-4 animate-spin motion-reduce:animate-none"
        />
      )}
    </div>
  )
}
