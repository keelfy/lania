import { getProfilesStats } from '@/lib/api-endpoints'
import { serverApiFetcher } from '@/lib/server'
import { cn } from '@/lib/utils'
import { RadioIcon, UserPlusIcon, UsersIcon } from 'lucide-react'
import { getTranslations } from 'next-intl/server'

type Props = {
  locale: string
  // The season the online count is of, missing for the primary one.
  season?: string
}

const ICONS = {
  total: UsersIcon,
  online: RadioIcon,
  newLastWeek: UserPlusIcon,
} as const

export default async function CommunityStats({ locale, season }: Props) {
  const t = await getTranslations({ locale, namespace: 'community.stats' })
  const stats = await getProfilesStats(serverApiFetcher, season).catch(
    (err) => {
      console.error(err)
      return null
    },
  )
  if (!stats) return null

  const items = [
    { key: 'total', value: stats.total, live: false },
    { key: 'online', value: stats.online, live: true },
    { key: 'newLastWeek', value: stats.newLastWeek, live: false },
  ] as const

  return (
    <dl className="grid grid-cols-3 gap-2">
      {items.map(({ key, value, live }) => {
        if (value === undefined) return null
        const Icon = ICONS[key]
        return (
          <div
            key={key}
            className="bg-card flex items-center gap-3 rounded-lg border p-3 shadow-sm"
          >
            <div className="bg-muted text-muted-foreground flex size-9 shrink-0 items-center justify-center rounded-md">
              <Icon
                className={cn('size-4.5', live && !!value && 'text-teal-400')}
              />
            </div>
            <div className="flex min-w-0 flex-col">
              <dd className="text-2xl font-bold tabular-nums">
                {value.toLocaleString(locale)}
              </dd>
              <dt className="text-muted-foreground truncate text-sm">
                {t(key)}
              </dt>
            </div>
          </div>
        )
      })}
    </dl>
  )
}
