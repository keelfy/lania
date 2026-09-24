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
