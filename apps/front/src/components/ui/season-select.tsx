'use client'

import { cn } from '@/lib/utils'
import { Season } from '@/models/season'
import { useTranslations } from 'next-intl'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from './select'

type Props = React.ComponentProps<typeof SelectTrigger> & {
  seasons: Season[]
  selectedSeasonId: string | undefined
  onSelectSeasonId: (id: string) => void
}

export default function SeasonSelect({
  seasons,
  selectedSeasonId,
  onSelectSeasonId,
  ...props
}: Props) {
  const t = useTranslations('seasonSelect')
  return (
    <Select value={selectedSeasonId} onValueChange={onSelectSeasonId}>
      <SelectTrigger {...props}>
        <SelectValue placeholder={t('placeholder')} />
      </SelectTrigger>
      <SelectContent>
        {seasons.map((season) => (
          <SelectItem key={season.id} value={season.id}>
            <span
              role="img"
              aria-label={t(season.isActive ? 'active' : 'inactive')}
              title={t(season.isActive ? 'active' : 'inactive')}
              className={cn(
                'size-2 shrink-0 rounded-full',
                season.isActive ? 'bg-green-500' : 'bg-muted-foreground/40',
              )}
            />
            {season.name}
            {season.isPrimary ? ` (${t('primary')})` : ''}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  )
}
