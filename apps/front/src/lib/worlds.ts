import { apiFetcher } from '@/lib/fetcher'
import { Season, SeasonWorld } from '@/models/season'

// Public data read without the visitor's cookies, so it is cached between requests.
const cached = { next: { revalidate: 60 } }

export type ActiveSeasonWorlds = { season: Season; worlds: SeasonWorld[] }

// Every running season with its worlds, the primary season first.
export async function getActiveSeasonWorlds(): Promise<ActiveSeasonWorlds[]> {
  const seasons = await apiFetcher<Season[]>(
    '/v1/seasons',
    new URLSearchParams(),
    cached,
  )
  const active = seasons
    .filter((season) => season.isActive)
    .sort((a, b) => Number(b.isPrimary) - Number(a.isPrimary))
  return Promise.all(
    active.map(async (season) => ({
      season,
      worlds: await apiFetcher<SeasonWorld[]>(
        `/v1/seasons/${season.id}/worlds`,
        new URLSearchParams(),
        cached,
      ).catch(() => []),
    })),
  )
}

// The world with the slug in a running season; the primary season wins when two share a slug.
export async function findActiveWorld(slug: string) {
  for (const { season, worlds } of await getActiveSeasonWorlds()) {
    const world = worlds.find((world) => world.slug === slug)
    if (world) return { season, world }
  }
  return undefined
}

export function getWorld(worldId: string) {
  return apiFetcher<SeasonWorld>(
    `/v1/worlds/${encodeURIComponent(worldId)}`,
    new URLSearchParams(),
    cached,
  )
}
