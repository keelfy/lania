import { getProfilesStats } from '@/lib/api-endpoints'
import { serverApiFetcher } from '@/lib/server'
import { getTranslations } from 'next-intl/server'

type Props = {
  locale: string
  // The season the online count is of, missing for the primary one.
  season?: string
}

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
    { key: 'total', value: stats.total },
    { key: 'online', value: stats.online },
    { key: 'newLastWeek', value: stats.newLastWeek },
  ] as const

  return (
    <dl className="grid grid-cols-3 gap-2">
      {items.map(
        ({ key, value }) =>
          value !== undefined && (
            <div key={key} className="rounded-md border p-3">
              <dd className="text-2xl font-bold tabular-nums">
                {value.toLocaleString(locale)}
              </dd>
              <dt className="text-muted-foreground text-sm">{t(key)}</dt>
            </div>
          ),
      )}
    </dl>
  )
}
