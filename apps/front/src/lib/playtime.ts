export type PlaytimeUnit = 'hours' | 'minutes' | 'seconds'

// Picks the largest unit that fits, playtime is in milliseconds.
export function formatPlaytime(playtime: number): {
  value: number
  unit: PlaytimeUnit
} {
  const seconds = playtime / 1000
  const minutes = seconds / 60
  const hours = minutes / 60
  if (minutes >= 60) return { value: Math.floor(hours), unit: 'hours' }
  if (seconds >= 60) return { value: Math.floor(minutes), unit: 'minutes' }
  return { value: Math.floor(seconds), unit: 'seconds' }
}
