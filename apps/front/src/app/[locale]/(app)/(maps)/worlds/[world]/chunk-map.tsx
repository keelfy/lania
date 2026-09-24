'use client'

import 'leaflet/dist/leaflet.css'
import type {
  Coords,
  LatLngBoundsExpression,
  LayerGroup,
  Map as LeafletMap,
  Marker,
} from 'leaflet'
import { useEffect, useRef, useState } from 'react'
import { usernameColorStyle } from '@/components/ui/mc-username'
import imgproxyImageLoader from '@/lib/imgproxyImageLoader'
import { ChunkPos, MapMarker, MapPlayer, MapDimension } from '@/models/claim'

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
  dimension: MapDimension
  // Fill colour of every claimed chunk, by chunkKey.
  claimColors: Map<string, string>
  selection: Map<string, { x: number; z: number; kind: SelectionKind }>
  // The chunk whose details are shown, outlined on the map.
  focused?: ChunkPos
  markers: MapMarker[]
  players: MapPlayer[]
  // A left click gives the same chunk twice; a left drag gives the corners of the dragged area.
  onChunkArea: (from: ChunkPos, to: ChunkPos) => void
  onChunkHover: (hover: ChunkHover | undefined) => void
}

// Draws squaremap tiles with a chunk grid, claimed chunks and the selection over them.
// The projection is squaremap's own: latlng units are blocks scaled down by 2^maxZoom, north is -z.
// The left button selects chunks; the map pans with the middle button (Leaflet's own drag) or a finger.
export default function ChunkMap({
  mapUrl,
  dimension,
  claimColors,
  selection,
  focused,
  markers,
  players,
  onChunkArea,
  onChunkHover,
}: Props) {
  const containerRef = useRef<HTMLDivElement>(null)
  const layersRef = useRef<{
    L: typeof import('leaflet')
    map: LeafletMap
    claims: import('leaflet').GridLayer
    selection: LayerGroup
    markers: LayerGroup
    players: LayerGroup
  }>(undefined)
  // Each player's marker with the look it was drawn with, so a changed look redraws it.
  const playerMarkersRef = useRef(
    new Map<string, { marker: Marker; look: string }>(),
  )
  const [ready, setReady] = useState(false)

  // Leaflet handlers live as long as the map, so they read the latest props through refs.
  const claimColorsRef = useRef(claimColors)
  const areaRef = useRef(onChunkArea)
  const hoverRef = useRef(onChunkHover)
  useEffect(() => {
    areaRef.current = onChunkArea
    hoverRef.current = onChunkHover
  }, [onChunkArea, onChunkHover])
  useEffect(() => {
    claimColorsRef.current = claimColors
    layersRef.current?.claims.redraw()
  }, [claimColors])

  useEffect(() => {
    let cancelled = false
    let map: LeafletMap | undefined
    let stopSelecting: (() => void) | undefined
    const container = containerRef.current
    const scale = 1 / 2 ** dimension.maxZoom
    const toLatLng = (x: number, z: number): [number, number] => [
      -z * scale,
      x * scale,
    ]
    const chunkBounds = (
      [ax, az]: ChunkPos,
      [bx, bz]: ChunkPos,
    ): [[number, number], [number, number]] => [
      toLatLng(Math.min(ax, bx) * CHUNK, Math.min(az, bz) * CHUNK),
      toLatLng((Math.max(ax, bx) + 1) * CHUNK, (Math.max(az, bz) + 1) * CHUNK),
    ]

    const onMouseDown = (e: MouseEvent) => {
      const layers = layersRef.current
      if (!layers) return
      // The middle button is Leaflet's drag; without this the browser would start autoscroll.
      if (e.button === 1) {
        e.preventDefault()
        return
      }
      if (e.button !== 0 || (e.target as Element).closest('.leaflet-control'))
        return
      // Runs in the capture phase, so Leaflet never sees the left button and does not pan.
      e.stopPropagation()
      e.preventDefault()

      const chunkOf = (event: MouseEvent): ChunkPos => {
        const latlng = layers.map.mouseEventToLatLng(event)
        return [
          Math.floor(Math.floor(latlng.lng / scale) / CHUNK),
          Math.floor(Math.floor(-latlng.lat / scale) / CHUNK),
        ]
      }
      const from = chunkOf(e)
      let to = from
      const area = layers.L.rectangle(chunkBounds(from, to), {
        color: '#ffffff',
        weight: 1,
        dashArray: '4 4',
        fillOpacity: 0.1,
        interactive: false,
      })

      const onMove = (event: MouseEvent) => {
        to = chunkOf(event)
        area.setBounds(layers.L.latLngBounds(chunkBounds(from, to)))
        if (to[0] !== from[0] || to[1] !== from[1]) area.addTo(layers.map)
      }
      const onUp = (event: MouseEvent) => {
        if (event.button !== 0) return
        stopSelecting?.()
        areaRef.current(from, to)
      }
      stopSelecting = () => {
        window.removeEventListener('mousemove', onMove)
        window.removeEventListener('mouseup', onUp)
        area.remove()
        stopSelecting = undefined
      }
      window.addEventListener('mousemove', onMove)
      window.addEventListener('mouseup', onUp)
    }

    void import('leaflet').then((L) => {
      if (cancelled || !container) return

      map = L.map(container, {
        crs: L.CRS.Simple,
        preferCanvas: true,
        attributionControl: false,
        boxZoom: false,
        doubleClickZoom: false,
        minZoom: 0,
        maxZoom: dimension.maxZoom + dimension.extraZoom,
      }).setView(
        toLatLng(dimension.spawn.x, dimension.spawn.z),
        dimension.maxZoom,
      )

      L.tileLayer(`${mapUrl}/tiles/${dimension.name}/{z}/{x}_{y}.png`, {
        tileSize: 512,
        minNativeZoom: 0,
        maxNativeZoom: dimension.maxZoom,
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
          const pxPerChunk = CHUNK * 2 ** (coords.z - dimension.maxZoom)
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
      const markerLayer = L.layerGroup().addTo(map)
      const playerLayer = L.layerGroup().addTo(map)

      map.on('mousemove', (e) => {
        const x = Math.floor(e.latlng.lng / scale)
        const z = Math.floor(-e.latlng.lat / scale)
        hoverRef.current({
          x,
          z,
          cx: Math.floor(x / CHUNK),
          cz: Math.floor(z / CHUNK),
        })
      })
      map.on('mouseout', () => hoverRef.current(undefined))
      // A tap on a touch screen arrives as a left click too.
      container.addEventListener('mousedown', onMouseDown, true)

      layersRef.current = {
        L,
        map,
        claims,
        selection: selectionLayer,
        markers: markerLayer,
        players: playerLayer,
      }
      setReady(true)
    })

    const playerMarkers = playerMarkersRef.current
    return () => {
      cancelled = true
      stopSelecting?.()
      container?.removeEventListener('mousedown', onMouseDown, true)
      layersRef.current = undefined
      playerMarkers.clear()
      setReady(false)
      map?.remove()
    }
  }, [mapUrl, dimension])

  useEffect(() => {
    const layers = layersRef.current
    if (!ready || !layers) return
    const scale = 1 / 2 ** dimension.maxZoom
    const chunkRect = (x: number, z: number): LatLngBoundsExpression => [
      [-z * CHUNK * scale, x * CHUNK * scale],
      [-(z + 1) * CHUNK * scale, (x + 1) * CHUNK * scale],
    ]
    layers.selection.clearLayers()
    for (const { x, z, kind } of selection.values()) {
      layers.L.rectangle(chunkRect(x, z), {
        color: selectionColors[kind],
        weight: 1,
        fillOpacity: 0.45,
        interactive: false,
      }).addTo(layers.selection)
    }
    if (focused) {
      layers.L.rectangle(chunkRect(...focused), {
        color: '#ffffff',
        weight: 2,
        fill: false,
        interactive: false,
      }).addTo(layers.selection)
    }
  }, [ready, selection, focused, dimension.maxZoom])

  useEffect(() => {
    const layers = layersRef.current
    if (!ready || !layers) return
    const scale = 1 / 2 ** dimension.maxZoom
    layers.markers.clearLayers()
    for (const marker of markers) {
      const icon = layers.L.marker([-marker.z * scale, marker.x * scale], {
        icon: layers.L.icon({
          iconUrl: marker.icon,
          iconSize: [16, 16],
          iconAnchor: [8, 8],
        }),
        // Interactive only so the label shows on hover; a left press still selects the chunk under it.
        keyboard: false,
      })
      if (marker.label) {
        icon.bindTooltip(textElement(marker.label), {
          direction: 'top',
          offset: [0, -8],
        })
      }
      icon.addTo(layers.markers)
    }
  }, [ready, markers, dimension.maxZoom])

  // Players are moved in place, so their nameplates do not blink on every poll.
  useEffect(() => {
    const layers = layersRef.current
    if (!ready || !layers) return
    const scale = 1 / 2 ** dimension.maxZoom
    const current = playerMarkersRef.current
    const online = new Set(players.map((player) => player.uuid))
    for (const [uuid, { marker }] of current) {
      if (online.has(uuid)) continue
      marker.remove()
      current.delete(uuid)
    }
    for (const player of players) {
      const latlng: [number, number] = [-player.z * scale, player.x * scale]
      const look = JSON.stringify([player.mojangUuid, player.cosmetics])
      const existing = current.get(player.uuid)
      if (existing?.look === look) {
        existing.marker.setLatLng(latlng)
        continue
      }
      existing?.marker.remove()
      const face = document.createElement('img')
      // Skins are looked up by the Mojang UUID; the in-game one only matches it on an online-mode server.
      face.src = `https://crafatar-pub.neodium.fr/avatars/${player.mojangUuid ?? player.uuid}?size=16&overlay`
      face.alt = player.name
      face.className = 'size-4 rounded-sm shadow ring-1 ring-black/60'
      const marker = layers.L.marker(latlng, {
        icon: layers.L.divIcon({
          html: face,
          className: '',
          iconSize: [16, 16],
          iconAnchor: [8, 8],
        }),
        interactive: false,
        keyboard: false,
        zIndexOffset: 1000,
      })
        .bindTooltip(nameplate(player), {
          permanent: true,
          direction: 'right',
          offset: [8, 0],
          // A dark plate like the in-game nameplate, so every name colour reads on it.
          className:
            'rounded-sm! border-none! bg-black/65! px-1.5! py-0.5! text-white! shadow-none! before:hidden',
        })
        .addTo(layers.players)
      current.set(player.uuid, { marker, look })
    }
  }, [ready, players, dimension.maxZoom])

  return (
    <div
      ref={containerRef}
      className="h-full w-full cursor-crosshair bg-[#1a1a1a]!"
    />
  )
}

// Leaflet writes a string tooltip as HTML; an element keeps names and labels as plain text.
function textElement(text: string) {
  const span = document.createElement('span')
  span.textContent = text
  return span
}

// The player's name as the game chat shows it: glyth and special prefixes, then the name in its colours.
// Built as DOM because Leaflet owns the tooltip; it mirrors NamePrefixes and McUsername.
function nameplate(player: MapPlayer) {
  const plate = document.createElement('span')
  plate.className = 'flex items-center gap-1'
  const cosmetics = player.cosmetics
  for (const prefix of [cosmetics?.glythPrefix, cosmetics?.specialPrefix]) {
    if (!prefix?.image) continue
    const image = document.createElement('img')
    // Prefix images are s3:// keys that only resolve through imgproxy, as next/image does elsewhere.
    image.src = imgproxyImageLoader({ src: prefix.image, width: 32 })
    image.alt = prefix.name
    image.className = 'size-4 [image-rendering:pixelated]'
    plate.append(image)
  }
  const name = textElement(player.name)
  const { isGradient, style } = usernameColorStyle(cosmetics?.colors?.colors)
  name.className = isGradient
    ? 'font-minecraft tracking-mc bg-clip-text text-transparent'
    : 'font-minecraft tracking-mc'
  if (style.backgroundImage) name.style.backgroundImage = style.backgroundImage
  if (style.color) name.style.color = style.color
  plate.append(name)
  return plate
}
