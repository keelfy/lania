'use client'

import { getImageProps } from 'next/image'
import { useEffect, useState } from 'react'

const SIZE = 32
const cache = new Map<string, string | null>()

// Keeps the hue, but lifts the color so a glow of it reads on the dark card.
function toAccent(r: number, g: number, b: number) {
  const max = Math.max(r, g, b) / 255
  const min = Math.min(r, g, b) / 255
  const lightness = (max + min) / 2
  const delta = max - min
  const saturation = delta === 0 ? 0 : delta / (1 - Math.abs(2 * lightness - 1))
  let hue = 0
  if (delta !== 0) {
    if (max * 255 === r) hue = ((g - b) / 255 / delta) % 6
    else if (max * 255 === g) hue = (b - r) / 255 / delta + 2
    else hue = (r - g) / 255 / delta + 4
  }
  const h = Math.round((hue * 60 + 360) % 360)
  const s = Math.round(Math.min(1, Math.max(saturation, 0.35)) * 100)
  const l = Math.round(Math.min(0.7, Math.max(lightness, 0.55)) * 100)
  return `hsl(${h} ${s}% ${l}%)`
}

function readAverage(
  src: string,
  unoptimized: boolean,
): Promise<string | null> {
  const { props } = getImageProps({
    src,
    alt: '',
    width: SIZE,
    height: SIZE,
    unoptimized,
  })
  return new Promise((resolve) => {
    const image = new window.Image()
    // Images come from imgproxy on another origin; without CORS the canvas cannot be read.
    image.crossOrigin = 'anonymous'
    image.onload = () => {
      try {
        resolve(averageOf(image))
      } catch {
        resolve(null)
      }
    }
    image.onerror = () => resolve(null)
    image.src = props.src
  })
}

// The average color of the vivid pixels, so dark outlines and gray shading do not turn the accent muddy.
function averageOf(image: HTMLImageElement) {
  const canvas = document.createElement('canvas')
  canvas.width = canvas.height = SIZE
  const context = canvas.getContext('2d', { willReadFrequently: true })
  if (!context) return null
  context.drawImage(image, 0, 0, SIZE, SIZE)
  const { data } = context.getImageData(0, 0, SIZE, SIZE)
  let r = 0
  let g = 0
  let b = 0
  let weight = 0
  for (let i = 0; i < data.length; i += 4) {
    const alpha = data[i + 3] / 255
    const max = Math.max(data[i], data[i + 1], data[i + 2])
    const min = Math.min(data[i], data[i + 1], data[i + 2])
    if (alpha < 0.5 || max < 40) continue
    const saturation = (max - min) / max
    const w = alpha * (0.02 + saturation * saturation)
    r += data[i] * w
    g += data[i + 1] * w
    b += data[i + 2] * w
    weight += w
  }
  return weight === 0 ? null : toAccent(r / weight, g / weight, b / weight)
}

// Undefined until the color is known, or when the image cannot be read.
export function useImageAccent(src?: string, unoptimized = false) {
  const [loaded, setLoaded] = useState<{ src: string; color: string | null }>()

  useEffect(() => {
    if (!src || cache.has(src)) return
    let active = true
    readAverage(src, unoptimized).then((color) => {
      cache.set(src, color)
      if (active) setLoaded({ src, color })
    })
    return () => {
      active = false
    }
  }, [src, unoptimized])

  if (!src) return undefined
  const color = cache.has(src) ? cache.get(src) : loaded?.color
  return color ?? undefined
}
