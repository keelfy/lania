'use client'

import { getImageProps } from 'next/image'
import { useEffect, useState } from 'react'

const SIZE = 32
const cache = new Map<string, string | null>()

// The average color of the opaque pixels. The image is read through the Next.js optimizer, so the
// canvas stays untainted without CORS headers on the image host.
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
    image.onload = () => {
      const canvas = document.createElement('canvas')
      canvas.width = canvas.height = SIZE
      const context = canvas.getContext('2d', { willReadFrequently: true })
      if (!context) return resolve(null)
      context.drawImage(image, 0, 0, SIZE, SIZE)
      const { data } = context.getImageData(0, 0, SIZE, SIZE)
      let r = 0
      let g = 0
      let b = 0
      let weight = 0
      for (let i = 0; i < data.length; i += 4) {
        const alpha = data[i + 3]
        r += data[i] * alpha
        g += data[i + 1] * alpha
        b += data[i + 2] * alpha
        weight += alpha
      }
      resolve(
        weight === 0
          ? null
          : `rgb(${Math.round(r / weight)} ${Math.round(g / weight)} ${Math.round(b / weight)})`,
      )
    }
    image.onerror = () => resolve(null)
    image.src = props.src
  })
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
