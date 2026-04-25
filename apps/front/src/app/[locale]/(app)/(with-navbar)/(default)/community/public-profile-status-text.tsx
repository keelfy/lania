'use client'
import { PROFILE_STATUS_COLORS } from '@/components/ui/player-card'
import { cn } from '@/lib/utils'
import { DotIcon } from 'lucide-react'
import { useTranslations } from 'next-intl'

type Props = React.ComponentProps<'div'> & {
  status: 'online' | 'offline' | 'banned'
}

export default function PublicProfileStatusText({
  status,
  className,
  ...props
}: Props) {
  const t = useTranslations('playerCard')
  return (
    <div className={cn('flex items-center gap-0', className)} {...props}>
      <DotIcon
        className={cn('size-6', status === 'online' && 'animate-pulse')}
        style={{ color: PROFILE_STATUS_COLORS[status] }}
      />
      <p
        className="font-medium"
        style={{ color: PROFILE_STATUS_COLORS[status] }}
      >
        {t(`statuses.${status}`)}
      </p>
    </div>
  )
}
