// A chunk index, the block coordinate divided by 16 and rounded down, as F3 shows it.
export type ChunkPos = [x: number, z: number]

export type ChunkClaim = {
  x: number
  z: number
  profileId: string
  claimedAt: number
}

export type ChunkClaimProfile = {
  id: string
  username: string
}

// Every claim that holds in one world; each claiming profile is listed once.
export type ChunkClaims = {
  claims: ChunkClaim[]
  profiles: ChunkClaimProfile[]
}

// A squaremap world as the claims map draws it, read from the map's settings.json.
export type MapWorld = {
  name: string
  type: 'normal' | 'nether' | 'the_end'
  // The zoom where one tile pixel is one block.
  maxZoom: number
  // Zoom levels past maxZoom that upscale the tiles.
  extraZoom: number
  spawn: { x: number; z: number }
}

// The most chunks one request may claim or release, the API's MaxClaimBatch.
export const MAX_CLAIM_BATCH = 256

// An icon squaremap shows on the world, like the spawn.
export type MapMarker = {
  icon: string
  x: number
  z: number
  label: string
}

export type MapPlayer = {
  name: string
  uuid: string
  x: number
  z: number
}

// What changes on the map while it is open: markers and the players online in the world.
export type MapLive = {
  markers: MapMarker[]
  players: MapPlayer[]
}
