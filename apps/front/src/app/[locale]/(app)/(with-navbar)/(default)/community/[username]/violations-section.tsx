import ProfileViolationList, {
  loadProfileViolations,
} from '@/components/profile-violations'
import { getTranslations } from 'next-intl/server'

type Props = {
  profileId: string
  // The season of the page, the primary one when missing.
  seasonId?: string
  locale: string
}

// Left out when the server of the season is over or cannot be reached: nothing is known then.
export default async function ViolationsSection({
  profileId,
  seasonId,
  locale,
}: Props) {
  const t = await getTranslations({ locale, namespace: 'violations' })
  const { available, violations, loadedAt } = await loadProfileViolations(
    profileId,
    seasonId,
  )
  if (!available) return null

  return (
    <section className="flex flex-col gap-2">
      <h2 className="text-xl font-bold tracking-tight">{t('title')}</h2>
      {violations.length === 0 ? (
        <p className="text-muted-foreground py-4 text-sm">{t('empty')}</p>
      ) : (
        <ProfileViolationList
          violations={violations}
          now={loadedAt}
          locale={locale}
        />
      )}
    </section>
  )
}
