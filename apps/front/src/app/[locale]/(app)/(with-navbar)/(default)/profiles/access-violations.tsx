import ProfileViolationList, {
  loadProfileViolations,
} from '@/components/profile-violations'
import { cn } from '@/lib/utils'
import {
  ShieldAlertIcon,
  ShieldCheckIcon,
  ShieldIcon,
  ShieldQuestionIcon,
} from 'lucide-react'
import { getTranslations } from 'next-intl/server'

type Props = {
  profileId: string
  // The season of the access card. The primary one when the profile has no access anywhere.
  seasonId?: string
  locale: string
}

// The value cell of the violations row, and the list under the row when there is one.
export default async function AccessViolations({
  profileId,
  seasonId,
  locale,
}: Props) {
  const t = await getTranslations({ locale, namespace: 'profiles.violations' })
  const { available, violations, loadedAt } = await loadProfileViolations(
    profileId,
    seasonId,
  )
  const active = violations.filter((v) => v.status === 'active').length

  let text: string
  let color: string
  let Icon = ShieldIcon
  if (!available) {
    text = t('noData')
    color = 'text-muted-foreground'
    Icon = ShieldQuestionIcon
  } else if (violations.length === 0) {
    text = t('noViolations')
    color = 'text-primary'
    Icon = ShieldCheckIcon
  } else if (active > 0) {
    text = t('active', { count: active })
    color = 'text-destructive'
    Icon = ShieldAlertIcon
  } else {
    text = t('count', { count: violations.length })
    color = 'text-orange-500'
  }

  return (
    <>
      <div className="flex min-h-9 flex-nowrap items-center justify-end gap-2">
        <p
          className={cn('text-right text-sm font-semibold sm:text-base', color)}
        >
          {text}
        </p>
        <Icon
          className={cn(
            'size-4',
            violations.length === 0 && available ? 'text-green-500' : color,
          )}
        />
      </div>
      {violations.length > 0 && (
        <ProfileViolationList
          violations={violations}
          now={loadedAt}
          locale={locale}
          className="col-span-2"
        />
      )}
    </>
  )
}

export function AccessViolationsFallback() {
  return (
    <div className="flex min-h-9 items-center justify-end">
      <div className="bg-muted h-4 w-28 animate-pulse rounded-sm" />
    </div>
  )
}
