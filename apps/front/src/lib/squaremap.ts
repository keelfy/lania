import { apiFetcher } from '@/lib/fetcher'
import { MapLive, MapWorld } from '@/models/claim'
import { Paginated } from '@/models/types'
import { PublicProfile } from '@/models/profile'
import { Season } from '@/models/season'

// squaremap sends no CORS headers, so everything it serves as JSON is read on the Next server.

type SquaremapSettings = {
  worlds: { name: string; type: MapWorld['type'] }[]
}

type SquaremapWorldSettings = {
  spawn: { x: number; z: number }
  zoom: { max: number; extra: number }
}

type SquaremapMarkerLayer = {
  hide: boolean
  markers: {
    type: string
    icon?: string
    point?: { x: number; z: number }
    tooltip?: string
  }[]
}

type SquaremapPlayers = {
  players: { name: string; uuid: string; world: string; x: number; z: number }[]
}

export const WORLD_NAME = /^[a-z0-9_]{1,64}$/

const worldOrder: MapWorld['type'][] = ['normal', 'nether', 'the_end']

async function fetchJson<T>(url: string, init: RequestInit): Promise<T> {
  const response = await fetch(url, init)
  if (!response.ok) throw new Error(`${url}: ${response.status}`)
  return response.json() as Promise<T>
}

// The primary season when it has a map, otherwise any running season that has one.
export function pickClaimsSeason(seasons: Season[]) {
  const claimable = seasons.filter((season) => season.isActive && season.mapUrl)
  return claimable.find((season) => season.isPrimary) ?? claimable[0]
}

export function mapBaseUrl(season: Season | undefined) {
  return season?.mapUrl?.replace(/\/+$/, '')
}

// The claims season, read without the visitor's cookies so it can be cached between requests.
export async function getClaimsSeason() {
  const seasons = await apiFetcher<Season[]>(
    '/v1/seasons',
    new URLSearchParams(),
    { next: { revalidate: 60 } },
  )
  return pickClaimsSeason(seasons)
}

// Who is online in the season with what they wear, by the in-game UUID squaremap reports.
// Shared by every open map for a while; a player who just joined shows plain until the next refresh.
async function getOnlineProfiles(seasonId: string) {
  const params = new URLSearchParams({
    online: 'true',
    seasonId,
    size: '100',
  })
  const page = await apiFetcher<Paginated<PublicProfile>>(
    '/v1/profiles',
    params,
    { next: { revalidate: 15 } },
  )
  return new Map(page.content.map((profile) => [profile.uuid, profile]))
}

export async function getMapWorlds(mapUrl: string): Promise<MapWorld[]> {
  const cache = { next: { revalidate: 300 } }
  const settings = await fetchJson<SquaremapSettings>(
    `${mapUrl}/tiles/settings.json`,
    cache,
  )
  const worlds = await Promise.all(
    settings.worlds.map(async ({ name, type }) => {
      const world = await fetchJson<SquaremapWorldSettings>(
        `${mapUrl}/tiles/${name}/settings.json`,
        cache,
      )
      return {
        name,
        type,
        maxZoom: world.zoom.max,
        extraZoom: world.zoom.extra,
        spawn: world.spawn,
      }
    }),
  )
  return worlds.sort(
    (a, b) => worldOrder.indexOf(a.type) - worldOrder.indexOf(b.type),
  )
}

// The icon markers of a world (spawn and whatever plugins add) and the players standing in it.
// Shapes such as the world border are left out: at claim scale they are noise.
export async function getMapLive(
  mapUrl: string,
  seasonId: string,
  world: string,
): Promise<MapLive> {
  const [layers, players, profiles] = await Promise.all([
    fetchJson<SquaremapMarkerLayer[]>(`${mapUrl}/tiles/${world}/markers.json`, {
      next: { revalidate: 30 },
    }),
    fetchJson<SquaremapPlayers>(`${mapUrl}/tiles/players.json`, {
      cache: 'no-store',
    }),
    // Players still show, plain, when the API cannot tell who is online.
    getOnlineProfiles(seasonId).catch(() => new Map<string, PublicProfile>()),
  ])
  return {
    markers: layers
      .filter((layer) => !layer.hide)
      .flatMap((layer) => layer.markers)
      .flatMap((marker) =>
        marker.type === 'icon' && marker.icon && marker.point
          ? [
              {
                icon: `${mapUrl}/images/icon/registered/${marker.icon}.png`,
                x: marker.point.x,
                z: marker.point.z,
                // Tooltips are squaremap HTML; the map shows them as text.
                label: (marker.tooltip ?? '').replace(/<[^>]*>/g, '').trim(),
              },
            ]
          : [],
      ),
    players: players.players
      .filter((player) => player.world === world)
      .map(({ name, uuid, x, z }) => {
        const profile = profiles.get(uuid)
        return {
          name,
          uuid,
          x,
          z,
          mojangUuid: profile?.mojangUuid,
          cosmetics: profile?.cosmetics.name,
        }
      }),
  }
}
