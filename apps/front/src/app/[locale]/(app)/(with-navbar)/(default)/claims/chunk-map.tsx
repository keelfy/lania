'use client'

import 'leaflet/dist/leaflet.css'
import type { Coords, LayerGroup, Map as LeafletMap } from 'leaflet'
import { useEffect, useRef, useState } from 'react'
import { MapWorld } from '@/models/claim'

const CHUNK = 16

export const chunkKey = (x: number, z: number) => `${x},${z}`

export type ChunkHover = { x: number; z: number; cx: number; cz: number }

export type SelectionKind = 'claim' | 'release' | 'adminRelease'

const selectionColors: Record<SelectionKind, string> = {
  claim: '#f59e0b',
  release: '#ef4444',
  adminRelease: '#a855f7',
}

type Props = {
  mapUrl: string
  world: MapWorld
  // Fill colour of every claimed chunk, by chunkKey.
  claimColors: Map<string, string>
  selection: Map<string, { x: number; z: number; kind: SelectionKind }>
  onChunkClick: (cx: number, cz: number, shift: boolean) => void
  onChunkHover: (hover: ChunkHover | undefined) => void
}

// Draws squaremap tiles with a chunk grid, claimed chunks and the selection over them.
// The projection is squaremap's own: latlng units are blocks scaled down by 2^maxZoom, north is -z.
export default function ChunkMap({
  mapUrl,
  world,
  claimColors,
  selection,
  onChunkClick,
  onChunkHover,
}: Props) {
  const containerRef = useRef<HTMLDivElement>(null)
  const layersRef = useRef<{
    L: typeof import('leaflet')
    map: LeafletMap
    claims: import('leaflet').GridLayer
    selection: LayerGroup
  }>(undefined)
  const [ready, setReady] = useState(false)

  // Leaflet handlers live as long as the map, so they read the latest props through refs.
  const claimColorsRef = useRef(claimColors)
  const clickRef = useRef(onChunkClick)
  const hoverRef = useRef(onChunkHover)
  useEffect(() => {
    clickRef.current = onChunkClick
    hoverRef.current = onChunkHover
  }, [onChunkClick, onChunkHover])
  useEffect(() => {
    claimColorsRef.current = claimColors
    layersRef.current?.claims.redraw()
  }, [claimColors])

  useEffect(() => {
    let cancelled = false
    let map: LeafletMap | undefined
    const scale = 1 / 2 ** world.maxZoom
    const toLatLng = (x: number, z: number): [number, number] => [
      -z * scale,
      x * scale,
    ]

    void import('leaflet').then((L) => {
      if (cancelled || !containerRef.current) return

      map = L.map(containerRef.current, {
        crs: L.CRS.Simple,
        preferCanvas: true,
        attributionControl: false,
        boxZoom: false,
        minZoom: 0,
        maxZoom: world.maxZoom + world.extraZoom,
      }).setView(toLatLng(world.spawn.x, world.spawn.z), world.maxZoom)

      L.tileLayer(`${mapUrl}/tiles/${world.name}/{z}/{x}_{y}.png`, {
        tileSize: 512,
        minNativeZoom: 0,
        maxNativeZoom: world.maxZoom,
        noWrap: true,
      }).addTo(map)

      // One canvas per tile: claimed chunks filled with the owner's colour, owner borders, and the grid.
      class ClaimsLayer extends L.GridLayer {
        createTile(coords: Coords) {
          const tile = document.createElement('canvas')
          const size = this.getTileSize()
          tile.width = size.x
          tile.height = size.y
          const ctx = tile.getContext('2d')!

          // At zoom z one block is 2^(z - maxZoom) screen pixels.
          const pxPerChunk = CHUNK * 2 ** (coords.z - world.maxZoom)
          const originX = coords.x * size.x
          const originY = coords.y * size.y
          const firstX = Math.floor(originX / pxPerChunk)
          const lastX = Math.floor((originX + size.x - 1) / pxPerChunk)
          const firstZ = Math.floor(originY / pxPerChunk)
          const lastZ = Math.floor((originY + size.y - 1) / pxPerChunk)

          const colors = claimColorsRef.current
          if (colors.size > 0) {
            for (let cx = firstX; cx <= lastX; cx++) {
              for (let cz = firstZ; cz <= lastZ; cz++) {
                const color = colors.get(chunkKey(cx, cz))
                if (!color) continue
                const left = cx * pxPerChunk - originX
                const top = cz * pxPerChunk - originY
                ctx.globalAlpha = 0.4
                ctx.fillStyle = color
                ctx.fillRect(left, top, pxPerChunk, pxPerChunk)

                // A solid edge where the neighbour belongs to someone else, so claims read as areas.
                if (pxPerChunk < 4) continue
                ctx.globalAlpha = 1
                ctx.strokeStyle = color
                ctx.lineWidth = 2
                ctx.beginPath()
                if (colors.get(chunkKey(cx, cz - 1)) !== color) {
                  ctx.moveTo(left, top + 1)
                  ctx.lineTo(left + pxPerChunk, top + 1)
                }
                if (colors.get(chunkKey(cx, cz + 1)) !== color) {
                  ctx.moveTo(left, top + pxPerChunk - 1)
                  ctx.lineTo(left + pxPerChunk, top + pxPerChunk - 1)
                }
                if (colors.get(chunkKey(cx - 1, cz)) !== color) {
                  ctx.moveTo(left + 1, top)
                  ctx.lineTo(left + 1, top + pxPerChunk)
                }
                if (colors.get(chunkKey(cx + 1, cz)) !== color) {
                  ctx.moveTo(left + pxPerChunk - 1, top)
                  ctx.lineTo(left + pxPerChunk - 1, top + pxPerChunk)
                }
                ctx.stroke()
              }
            }
          }

          // The grid would be a solid wash when chunks are only a few pixels wide.
          if (pxPerChunk < 8) return tile
          ctx.globalAlpha = 1
          ctx.strokeStyle = 'rgba(255,255,255,0.3)'
          ctx.lineWidth = 1
          ctx.beginPath()
          for (let cx = firstX; cx <= lastX + 1; cx++) {
            const x = cx * pxPerChunk - originX + 0.5
            ctx.moveTo(x, 0)
            ctx.lineTo(x, size.y)
          }
          for (let cz = firstZ; cz <= lastZ + 1; cz++) {
            const y = cz * pxPerChunk - originY + 0.5
            ctx.moveTo(0, y)
            ctx.lineTo(size.x, y)
          }
          ctx.stroke()
          return tile
        }
      }
      const claims = new ClaimsLayer({ tileSize: 256 }).addTo(map)
      const selectionLayer = L.layerGroup().addTo(map)

      const chunkAt = (latlng: import('leaflet').LatLng) => {
        const x = Math.floor(latlng.lng / scale)
        const z = Math.floor(-latlng.lat / scale)
        return {
          x,
          z,
          cx: Math.floor(x / CHUNK),
          cz: Math.floor(z / CHUNK),
        }
      }
      map.on('mousemove', (e) => hoverRef.current(chunkAt(e.latlng)))
      map.on('mouseout', () => hoverRef.current(undefined))
      map.on('click', (e) => {
        const { cx, cz } = chunkAt(e.latlng)
        clickRef.current(cx, cz, e.originalEvent.shiftKey)
      })

      layersRef.current = { L, map, claims, selection: selectionLayer }
      setReady(true)
    })

    return () => {
      cancelled = true
      layersRef.current = undefined
      setReady(false)
      map?.remove()
    }
  }, [mapUrl, world])

  useEffect(() => {
    const layers = layersRef.current
    if (!ready || !layers) return
    const scale = 1 / 2 ** world.maxZoom
    layers.selection.clearLayers()
    for (const { x, z, kind } of selection.values()) {
      layers.L.rectangle(
        [
          [-z * CHUNK * scale, x * CHUNK * scale],
          [-(z + 1) * CHUNK * scale, (x + 1) * CHUNK * scale],
        ],
        {
          color: selectionColors[kind],
          weight: 1,
          fillOpacity: 0.45,
          interactive: false,
        },
      ).addTo(layers.selection)
    }
  }, [ready, selection, world.maxZoom])

  return (
    <div
      ref={containerRef}
      className="h-[70vh] w-full cursor-crosshair rounded-lg bg-[#1a1a1a]!"
    />
  )
}
