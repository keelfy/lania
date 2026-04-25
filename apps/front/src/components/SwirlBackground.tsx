'use client'

import { detectPowerSavingMode } from '@/lib/power-saving'
import { cn } from '@/lib/utils'
import React, { useCallback, useEffect, useRef } from 'react'
import { createNoise3D } from 'simplex-noise'

// Utility functions
const { PI, cos, sin, abs, random } = Math
const TAU = 2 * PI
const rand = (n: number) => n * random()
const randRange = (n: number) => n - rand(2 * n)
const fadeInOut = (t: number, m: number) => {
  const hm = 0.5 * m
  return abs(((t + hm) % m) - hm) / hm
}
const lerp = (n1: number, n2: number, speed: number) =>
  (1 - speed) * n1 + speed * n2

interface CanvasRefs {
  a: HTMLCanvasElement
  b: HTMLCanvasElement
}

interface ContextRefs {
  a: CanvasRenderingContext2D
  b: CanvasRenderingContext2D
}

export default function SwirlBackground({
  className,
  ...props
}: React.ComponentProps<'div'>) {
  const containerRef = useRef<HTMLDivElement>(null)
  const canvasRef = useRef<CanvasRefs | null>(null)
  const ctxRef = useRef<ContextRefs | null>(null)
  const centerRef = useRef<[number, number]>([0, 0])
  const tickRef = useRef(0)
  const particlePropsRef = useRef<Float32Array | null>(null)
  const animationFrameRef = useRef<number | null>(null)
  const noise3DRef = useRef<ReturnType<typeof createNoise3D> | null>(null)
  const [powerSavingMode, setPowerSavingMode] = React.useState<boolean>()

  // Constants
  const particleCount = 700
  const particlePropCount = 9
  const particlePropsLength = particleCount * particlePropCount
  const rangeY = 200
  const baseTTL = 100
  const rangeTTL = 200
  const baseSpeed = 0.03
  const rangeSpeed = 0.7
  const baseRadius = 1
  const rangeRadius = 4
  const baseHue = 160
  const rangeHue = 40
  const noiseSteps = 8
  const xOff = 0.00125
  const yOff = 0.00125
  const zOff = 0.0002 // Reduced from 0.0005

  // Initialize noise3D once
  if (!noise3DRef.current) {
    noise3DRef.current = createNoise3D(Math.random)
  }

  const backgroundColor = React.useMemo(
    () =>
      typeof document !== 'undefined'
        ? getComputedStyle(document.documentElement).getPropertyValue(
            '--background',
          )
        : 'transparent',
    [],
  )

  const createCanvas = useCallback(() => {
    if (!containerRef.current) return

    const canvas = {
      a: document.createElement('canvas'),
      b: document.createElement('canvas'),
    }

    canvas.b.style.cssText = `
      position: fixed;
      top: 0;
      left: 0;
      width: 100%;
      height: 100%;
      pointer-events: none;
      z-index: -1;
    `

    containerRef.current.appendChild(canvas.b)

    const ctx = {
      a: canvas.a.getContext('2d')!,
      b: canvas.b.getContext('2d')!,
    }

    canvasRef.current = canvas
    ctxRef.current = ctx
    centerRef.current = [0, 0]
  }, [])

  const resize = useCallback(() => {
    if (!canvasRef.current || !ctxRef.current) return

    const { innerWidth, innerHeight } = window
    const canvas = canvasRef.current
    const ctx = ctxRef.current

    canvas.a.width = innerWidth
    canvas.a.height = innerHeight

    ctx.a.drawImage(canvas.b, 0, 0)

    canvas.b.width = innerWidth
    canvas.b.height = innerHeight

    ctx.b.drawImage(canvas.a, 0, 0)

    centerRef.current[0] = 0.5 * canvas.a.width
    centerRef.current[1] = 0.5 * canvas.a.height
  }, [])

  const initParticle = useCallback(
    (i: number) => {
      if (!canvasRef.current || !particlePropsRef.current) return

      const canvas = canvasRef.current
      const particleProps = particlePropsRef.current
      const center = centerRef.current

      const x = rand(canvas.a.width)
      const y = center[1] + randRange(rangeY)
      const vx = 0
      const vy = 0
      const life = 0
      const ttl = baseTTL + rand(rangeTTL)
      const speed = baseSpeed + rand(rangeSpeed)
      const radius = baseRadius + rand(rangeRadius)
      const hue = baseHue + rand(rangeHue)

      particleProps.set([x, y, vx, vy, life, ttl, speed, radius, hue], i)
    },
    [
      baseTTL,
      rangeTTL,
      baseSpeed,
      rangeSpeed,
      baseRadius,
      rangeRadius,
      baseHue,
      rangeHue,
      rangeY,
    ],
  )

  const initParticles = useCallback(() => {
    tickRef.current = 0
    particlePropsRef.current = new Float32Array(particlePropsLength)

    for (let i = 0; i < particlePropsLength; i += particlePropCount) {
      initParticle(i)
    }
  }, [particlePropsLength, particlePropCount, initParticle])

  const checkBounds = useCallback((x: number, y: number) => {
    if (!canvasRef.current) return false

    return (
      x > canvasRef.current.a.width ||
      x < 0 ||
      y > canvasRef.current.a.height ||
      y < 0
    )
  }, [])

  const drawParticle = useCallback(
    (
      x: number,
      y: number,
      x2: number,
      y2: number,
      life: number,
      ttl: number,
      radius: number,
      hue: number,
    ) => {
      if (!ctxRef.current) return

      const ctx = ctxRef.current

      ctx.a.save()
      ctx.a.lineCap = 'round'
      ctx.a.lineWidth = radius
      ctx.a.strokeStyle = `hsla(${hue},100%,60%,${fadeInOut(life, ttl)})`
      ctx.a.beginPath()
      ctx.a.moveTo(x, y)
      ctx.a.lineTo(x2, y2)
      ctx.a.stroke()
      ctx.a.closePath()
      ctx.a.restore()
    },
    [],
  )

  const updateParticle = useCallback(
    (i: number) => {
      if (!particlePropsRef.current || !ctxRef.current || !noise3DRef.current)
        return

      const particleProps = particlePropsRef.current
      const noise3D = noise3DRef.current

      const i2 = 1 + i,
        i3 = 2 + i,
        i4 = 3 + i,
        i5 = 4 + i,
        i6 = 5 + i,
        i7 = 6 + i,
        i8 = 7 + i,
        i9 = 8 + i

      const x = particleProps[i]
      const y = particleProps[i2]
      const n =
        noise3D(x * xOff, y * yOff, tickRef.current * zOff) * noiseSteps * TAU
      const vx = lerp(particleProps[i3], cos(n), 0.5)
      const vy = lerp(particleProps[i4], sin(n), 0.5)
      let life = particleProps[i5]
      const ttl = particleProps[i6]
      const speed = particleProps[i7]
      const x2 = x + vx * speed
      const y2 = y + vy * speed
      const radius = particleProps[i8]
      const hue = particleProps[i9]

      drawParticle(x, y, x2, y2, life, ttl, radius, hue)

      life++

      particleProps[i] = x2
      particleProps[i2] = y2
      particleProps[i3] = vx
      particleProps[i4] = vy
      particleProps[i5] = life

      if (checkBounds(x, y) || life > ttl) {
        initParticle(i)
      }
    },
    [xOff, yOff, zOff, noiseSteps, drawParticle, checkBounds, initParticle],
  )

  const drawParticles = useCallback(() => {
    for (let i = 0; i < particlePropsLength; i += particlePropCount) {
      updateParticle(i)
    }
  }, [particlePropsLength, particlePropCount, updateParticle])

  const renderGlow = useCallback(() => {
    // glow works very slowly in firefox
    if (/firefox/i.test(navigator.userAgent) || powerSavingMode) {
      return
    }

    if (!canvasRef.current || !ctxRef.current) return

    const canvas = canvasRef.current
    const ctx = ctxRef.current

    ctx.b.save()
    ctx.b.filter = 'blur(8px) brightness(200%)'
    ctx.b.globalCompositeOperation = 'lighter'
    ctx.b.drawImage(canvas.a, 0, 0)
    ctx.b.restore()

    ctx.b.save()
    ctx.b.filter = 'blur(4px) brightness(200%)'
    ctx.b.globalCompositeOperation = 'lighter'
    ctx.b.drawImage(canvas.a, 0, 0)
    ctx.b.restore()
  }, [powerSavingMode])

  const renderToScreen = useCallback(() => {
    if (!canvasRef.current || !ctxRef.current) return

    const canvas = canvasRef.current
    const ctx = ctxRef.current

    ctx.b.save()
    ctx.b.globalCompositeOperation = 'lighter'
    ctx.b.drawImage(canvas.a, 0, 0)
    ctx.b.restore()
  }, [])

  const draw = useCallback(() => {
    if (!canvasRef.current || !ctxRef.current) return

    const canvas = canvasRef.current
    const ctx = ctxRef.current

    tickRef.current++

    ctx.a.clearRect(0, 0, canvas.a.width, canvas.a.height)

    ctx.b.fillStyle = backgroundColor
    ctx.b.fillRect(0, 0, canvas.a.width, canvas.a.height)

    drawParticles()
    renderGlow()
    renderToScreen()

    animationFrameRef.current = window.requestAnimationFrame(draw)
  }, [drawParticles, renderGlow, renderToScreen, backgroundColor])

  const setup = useCallback(() => {
    noise3DRef.current = createNoise3D(Math.random) // re-init on setup (like original)
    createCanvas()
    resize()
    initParticles()
    draw()
  }, [createCanvas, resize, initParticles, draw])

  useEffect(() => {
    setup()

    const handleResize = () => {
      resize()
    }

    window.addEventListener('resize', handleResize)

    return () => {
      window.removeEventListener('resize', handleResize)
      if (animationFrameRef.current) {
        window.cancelAnimationFrame(animationFrameRef.current)
      }
      // Store refs in variables to avoid the warning
      const container = containerRef.current
      const canvas = canvasRef.current
      if (container && canvas && container.contains(canvas.b)) {
        container.removeChild(canvas.b)
      }
    }
  }, [setup, resize])

  useEffect(() => {
    detectPowerSavingMode().then(setPowerSavingMode)
  }, [])

  return (
    <div
      ref={containerRef}
      className={cn('pointer-events-none fixed inset-0 z-[-1]', className)}
      {...props}
    />
  )
}
