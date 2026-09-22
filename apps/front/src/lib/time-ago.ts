export type TimeAgoUnit =
  | 'now'
  | 'minutes'
  | 'hours'
  | 'days'
  | 'weeks'
  | 'months'
  | 'years'

// Picks the largest unit that fits, mirrors formatPlaytime's short unit style.
export function formatTimeAgo(
  date: number | string | Date,
  now: number = Date.now(),
): { value: number; unit: TimeAgoUnit } {
  const timestamp = typeof date === 'number' ? date : new Date(date).getTime()
  const seconds = Math.max(0, (now - timestamp) / 1000)
  const minutes = seconds / 60
  const hours = minutes / 60
  const days = hours / 24
  const weeks = days / 7
  const months = days / 30
  const years = days / 365

  if (years >= 1) return { value: Math.floor(years), unit: 'years' }
  if (months >= 1) return { value: Math.floor(months), unit: 'months' }
  if (weeks >= 1) return { value: Math.floor(weeks), unit: 'weeks' }
  if (days >= 1) return { value: Math.floor(days), unit: 'days' }
  if (hours >= 1) return { value: Math.floor(hours), unit: 'hours' }
  if (minutes >= 1) return { value: Math.floor(minutes), unit: 'minutes' }
  return { value: 0, unit: 'now' }
}
