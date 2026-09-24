import { DIMENSION_NAME, getMapLive, mapBaseUrl } from '@/lib/squaremap'
import { getWorld } from '@/lib/worlds'
import { NextRequest, NextResponse } from 'next/server'

// Relays squaremap's markers and players to the world map, which cannot read them across origins.
// Only the map a world names is read, so this is no open proxy.
export async function GET(
  request: NextRequest,
  { params }: { params: Promise<{ worldId: string }> },
) {
  const { worldId } = await params
  const dimension = request.nextUrl.searchParams.get('dimension') ?? ''
  if (!DIMENSION_NAME.test(dimension)) {
    return NextResponse.json({ error: 'bad dimension' }, { status: 400 })
  }
  const world = await getWorld(worldId).catch(() => undefined)
  const mapUrl = mapBaseUrl(world?.mapUrl)
  if (!world || !mapUrl) {
    return NextResponse.json({ error: 'no map' }, { status: 404 })
  }
  try {
    return NextResponse.json(
      await getMapLive(mapUrl, world.seasonId, dimension),
      { headers: { 'Cache-Control': 'no-store' } },
    )
  } catch (error) {
    console.error('squaremap live data: ', error)
    return NextResponse.json({ error: 'map unavailable' }, { status: 502 })
  }
}
