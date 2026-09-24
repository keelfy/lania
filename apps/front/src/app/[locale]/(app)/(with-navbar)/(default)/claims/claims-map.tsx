'use client'

import 'leaflet/dist/leaflet.css'
import type { LatLng, Map as LeafletMap } from 'leaflet'
import { useEffect, useRef, useState } from 'react'

const MAP_URL = 'https://survival-map.lania.network'
const WORLD = 'minecraft_overworld'
// squaremap world settings: zoom.max is the zoom where one tile pixel is one block.
const MAX_NATIVE_ZOOM = 3
const EXTRA_ZOOM = 2
const SPAWN = { x: -624, z: -528 }
const CHUNK = 16

// Same projection as squaremap: latlng units are blocks scaled down by 2^maxZoom, north is -z.
const SCALE = 1 / 2 ** MAX_NATIVE_ZOOM
const toLatLng = (x: number, z: number): [number, number] => [
  -z * SCALE,
  x * SCALE,
]
const toBlock = (latlng: LatLng) => ({
  x: latlng.lng / SCALE,
  z: -latlng.lat / SCALE,
})

type Cursor = { x: number; z: number; cx: number; cz: number }

export default function ClaimsMap() {
  const containerRef = useRef<HTMLDivElement>(null)
  const [cursor, setCursor] = useState<Cursor | null>(null)
  const [selected, setSelected] = useState<string[]>([])

  useEffect(() => {
    let map: LeafletMap | undefined
    let cancelled = false

    import('leaflet').then((L) => {
      if (cancelled || !containerRef.current) return

      map = L.map(containerRef.current, {
        crs: L.CRS.Simple,
        preferCanvas: true,
        attributionControl: false,
        minZoom: 0,
        maxZoom: MAX_NATIVE_ZOOM + EXTRA_ZOOM,
      }).setView(toLatLng(SPAWN.x, SPAWN.z), MAX_NATIVE_ZOOM)

      L.tileLayer(`${MAP_URL}/tiles/${WORLD}/{z}/{x}_{y}.png`, {
        tileSize: 512,
        minNativeZoom: 0,
        maxNativeZoom: MAX_NATIVE_ZOOM,
        noWrap: true,
      }).addTo(map)

      // Chunk grid drawn per screen tile; a block spans 2^(zoom - maxZoom) screen pixels.
      class ChunkGrid extends L.GridLayer {
        createTile(coords: L.Coords) {
          const tile = document.createElement('canvas')
          const size = this.getTileSize()
          tile.width = size.x
          tile.height = size.y
          const pxPerBlock = 2 ** (coords.z - MAX_NATIVE_ZOOM)
          const pxPerChunk = CHUNK * pxPerBlock
          if (pxPerChunk < 8) return tile

          const ctx = tile.getContext('2d')!
          ctx.strokeStyle = 'rgba(255,255,255,0.35)'
          ctx.lineWidth = 1
          const originX = coords.x * size.x
          const originY = coords.y * size.y
          const firstX = Math.ceil(originX / pxPerChunk) * pxPerChunk - originX
          const firstY = Math.ceil(originY / pxPerChunk) * pxPerChunk - originY
          ctx.beginPath()
          for (let x = firstX; x < size.x; x += pxPerChunk) {
            ctx.moveTo(x + 0.5, 0)
            ctx.lineTo(x + 0.5, size.y)
          }
          for (let y = firstY; y < size.y; y += pxPerChunk) {
            ctx.moveTo(0, y + 0.5)
            ctx.lineTo(size.x, y + 0.5)
          }
          ctx.stroke()
          return tile
        }
      }
      new ChunkGrid({ tileSize: 256 }).addTo(map)

      const selection = new Map<string, L.Rectangle>()

      map.on('mousemove', (e) => {
        const { x, z } = toBlock(e.latlng)
        const bx = Math.floor(x)
        const bz = Math.floor(z)
        setCursor({
          x: bx,
          z: bz,
          cx: Math.floor(bx / CHUNK),
          cz: Math.floor(bz / CHUNK),
        })
      })

      map.on('click', (e) => {
        const { x, z } = toBlock(e.latlng)
        const cx = Math.floor(x / CHUNK)
        const cz = Math.floor(z / CHUNK)
        const key = `${cx},${cz}`
        const existing = selection.get(key)
        if (existing) {
          existing.remove()
          selection.delete(key)
        } else {
          const rect = L.rectangle(
            [
              toLatLng(cx * CHUNK, cz * CHUNK),
              toLatLng((cx + 1) * CHUNK, (cz + 1) * CHUNK),
            ],
            {
              color: '#f59e0b',
              weight: 1,
              fillOpacity: 0.35,
              interactive: false,
            },
          ).addTo(map!)
          selection.set(key, rect)
        }
        setSelected([...selection.keys()])
      })
    })

    return () => {
      cancelled = true
      map?.remove()
    }
  }, [])

  return (
    <div className="flex flex-col gap-3">
      <div
        ref={containerRef}
        className="h-[70vh] w-full rounded-lg bg-[#1a1a1a]!"
      />
      <div className="text-muted-foreground flex flex-wrap gap-x-6 gap-y-1 font-mono text-sm">
        <span>block: {cursor ? `${cursor.x}, ${cursor.z}` : '—'}</span>
        <span>chunk: {cursor ? `${cursor.cx}, ${cursor.cz}` : '—'}</span>
        <span>selected: {selected.length ? selected.join(' · ') : '—'}</span>
      </div>
    </div>
  )
}
