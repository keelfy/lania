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

// All seasons with wiki content can be selected. The API owns the primary season.
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

  return wikiSeasons
}
