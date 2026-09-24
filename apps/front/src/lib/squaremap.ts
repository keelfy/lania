import { apiFetcher } from '@/lib/fetcher'
import { MapLive, MapWorld } from '@/models/claim'
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

// The map of the claims season, read without the visitor's cookies so it can be cached between requests.
export async function getClaimsMapUrl() {
  const seasons = await apiFetcher<Season[]>(
    '/v1/seasons',
    new URLSearchParams(),
    { next: { revalidate: 60 } },
  )
  return mapBaseUrl(pickClaimsSeason(seasons))
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
  world: string,
): Promise<MapLive> {
  const [layers, players] = await Promise.all([
    fetchJson<SquaremapMarkerLayer[]>(`${mapUrl}/tiles/${world}/markers.json`, {
      next: { revalidate: 30 },
    }),
    fetchJson<SquaremapPlayers>(`${mapUrl}/tiles/players.json`, {
      cache: 'no-store',
    }),
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
      .map(({ name, uuid, x, z }) => ({ name, uuid, x, z })),
  }
}
