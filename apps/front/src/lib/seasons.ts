import { Season } from '@/models/season'

// Seasons a player can get access to: the ones that are running.
export function getSelectableSeasons(seasons: Season[]) {
  return seasons.filter((season) => season.isActive)
}

// The requested season if it can be picked, otherwise the primary one.
export function pickSeason(seasons: Season[], seasonId?: string | null) {
  return (
    seasons.find((season) => season.id === seasonId) ??
    seasons.find((season) => season.isPrimary) ??
    seasons[0]
  )
}

// Calendar-correct month/day arithmetic between two dates. undefined endDate means the season is still
// running, the caller falls back to an "active" label instead of a length.
export function formatSeasonDuration(startDate: number, endDate?: number) {
  if (endDate === undefined) return undefined
  const start = new Date(startDate)
  const end = new Date(endDate)

  let months =
    (end.getUTCFullYear() - start.getUTCFullYear()) * 12 +
    (end.getUTCMonth() - start.getUTCMonth())
  let days = end.getUTCDate() - start.getUTCDate()
  if (days < 0) {
    months -= 1
    // Days in the month before `end`.
    days += new Date(
      Date.UTC(end.getUTCFullYear(), end.getUTCMonth(), 0),
    ).getUTCDate()
  }
  return { months, days }
}
