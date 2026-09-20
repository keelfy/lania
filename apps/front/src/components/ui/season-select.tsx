'use client'

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
            {season.name}
            {season.isPrimary ? ` (${t('primary')})` : ''}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  )
}
