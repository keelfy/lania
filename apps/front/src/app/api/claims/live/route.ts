import {
  WORLD_NAME,
  getClaimsSeason,
  getMapLive,
  mapBaseUrl,
} from '@/lib/squaremap'
import { NextRequest, NextResponse } from 'next/server'

// Relays squaremap's markers and players to the claims map, which cannot read them across origins.
// Only the claims season's map is read, so this is no open proxy.
export async function GET(request: NextRequest) {
  const world = request.nextUrl.searchParams.get('world') ?? ''
  if (!WORLD_NAME.test(world)) {
    return NextResponse.json({ error: 'bad world' }, { status: 400 })
  }
  try {
    const season = await getClaimsSeason()
    const mapUrl = mapBaseUrl(season)
    if (!season || !mapUrl) {
      return NextResponse.json({ error: 'no map' }, { status: 404 })
    }
    return NextResponse.json(await getMapLive(mapUrl, season.id, world), {
      headers: { 'Cache-Control': 'no-store' },
    })
  } catch (error) {
    console.error('squaremap live data: ', error)
    return NextResponse.json({ error: 'map unavailable' }, { status: 502 })
  }
}
