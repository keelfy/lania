import { Season } from '@/models/season'
import { apiFetcher } from './fetcher'

// A season that has its own version of the wiki: a content folder named after its slug.
export type WikiSeason = {
  slug: string
  name: string
  isPrimary: boolean
}

// "Lania V" -> "lania-v". The wiki folder of a season must be named with this slug.
export function seasonSlug(name: string) {
  return name
    .normalize('NFKD')
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
}

// Seasons whose slug matches one of the wiki folders, newest first. The primary season of the API is the
// current wiki; when the API is down, the newest folder stands in for it.
export async function getWikiSeasons(folders: string[]): Promise<WikiSeason[]> {
  const seasons = await apiFetcher<Season[]>(
    '/v1/seasons',
    new URLSearchParams(),
    { next: { revalidate: 300 } },
  ).catch(() => [] as Season[])

  const wikiSeasons = seasons
    .map((season) => ({
      slug: seasonSlug(season.name),
      name: season.name,
      isPrimary: season.isPrimary,
    }))
    .filter((season) => folders.includes(season.slug))

  if (wikiSeasons.length > 0 && !wikiSeasons.some((s) => s.isPrimary)) {
    wikiSeasons[0].isPrimary = true
  }
  return wikiSeasons
}
