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
