'use client'

import McUsername from '@/components/ui/mc-username'
import { cn } from '@/lib/utils'
import dynamic from 'next/dynamic'
import Image from 'next/image'
import { useTranslations } from 'next-intl'
import { useEffect, useRef, useState } from 'react'

const SkinPreview = dynamic(() => import('./customization-skin-preview'), {
  ssr: false,
})

import styles from './landing-motion.module.css'

const asset = '/images/landing-customization/'
type ColorOption = { id: string; name: string; colors: string[] }
type GlyphOption = { id: string; name: string; image: string | null }

const skins = [
  { name: 'Steve', src: '/images/steve_skin.png', slim: false },
  { name: 'Alex', src: `${asset}alex.png`, slim: true },
  { name: 'Sunny', src: `${asset}sunny.png`, slim: false },
  { name: 'Efe', src: `${asset}efe.png`, slim: true },
]
const choiceClass =
  styles.choice +
  ' ' +
  'flex min-h-11 items-center justify-center gap-2 rounded-lg border px-3 py-2 text-sm text-white/80 hover:bg-white/10 focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-teal-300'
const selectedClass = 'border-teal-300 bg-teal-300/10 text-white'

export default function CustomizationShowcase({
  colors: catalogColors,
  glyphs: catalogGlyphs,
}: {
  colors: ColorOption[]
  glyphs: GlyphOption[]
}) {
  const colors = [
    { id: 'classic', name: 'classic', colors: ['#FFFFFF'] },
    ...catalogColors,
  ]
  const glyphs = [{ id: 'none', name: 'none', image: null }, ...catalogGlyphs]
  const t = useTranslations('landing.customization')
  const [colorIndex, setColorIndex] = useState(catalogColors.length > 0 ? 1 : 0)
  const [glyphIndex, setGlyphIndex] = useState(catalogGlyphs.length > 0 ? 1 : 0)
  const [skinIndex, setSkinIndex] = useState(0)
  const [visible, setVisible] = useState(false)
  const stageRef = useRef<HTMLDivElement>(null)
  const color = colors[colorIndex]
  const glyph = glyphs[glyphIndex]
  const skin = skins[skinIndex]

  useEffect(() => {
    const observer = new IntersectionObserver(
      ([entry]) => setVisible(entry.isIntersecting),
      { rootMargin: '200px' },
    )
    if (stageRef.current) observer.observe(stageRef.current)
    return () => observer.disconnect()
  }, [])

  return (
    <div className="grid gap-8 lg:grid-cols-[1fr_1.1fr] lg:gap-14">
      <div
        ref={stageRef}
        className="relative flex flex-col items-center justify-center rounded-2xl border border-white/10 bg-black/20 px-4 pt-6 pb-4"
      >
        <div
          aria-hidden
          className={`pointer-events-none absolute inset-0 rounded-2xl ${styles.glow}`}
          style={{
            backgroundColor: `color-mix(in srgb, ${color.colors[color.colors.length - 1]} 18%, transparent)`,
          }}
        />
        <div
          className="relative flex min-h-12 items-center gap-2 rounded-md bg-black/35 px-4 py-2"
          aria-live="polite"
          aria-atomic="true"
        >
          {glyph.image ? (
            <Image
              key={glyph.id}
              src={glyph.image}
              alt={glyph.name}
              width={26}
              height={26}
              className={`[image-rendering:pixelated] ${styles.glyph}`}
            />
          ) : null}
          <McUsername
            username="YourName"
            colors={color.colors}
            className="text-2xl"
          />
        </div>
        <div className="relative flex h-[360px] w-full items-center justify-center">
          {visible ? (
            <SkinPreview
              key={skin.name}
              src={skin.src}
              slim={skin.slim}
              label={t('preview', { skin: skin.name })}
              fallback={t('fallback')}
            />
          ) : null}
        </div>
        <p className="relative text-center text-xs text-white/55">
          {t('rotate')}
        </p>
      </div>

      <div className="flex flex-col justify-center gap-6">
        <fieldset className="space-y-3">
          <legend className="mb-1 font-medium text-white">{t('color')}</legend>
          <div className="flex flex-wrap gap-2">
            {colors.map((option, index) => (
              <button
                key={option.id}
                type="button"
                aria-pressed={colorIndex === index}
                onClick={() => setColorIndex(index)}
                className={cn(
                  choiceClass,
                  colorIndex === index ? selectedClass : 'border-white/15',
                )}
              >
                <span
                  aria-hidden
                  className="size-4 shrink-0 rounded-full"
                  style={{
                    background: `linear-gradient(to right, ${option.colors.length === 1 ? `${option.colors[0]}, ${option.colors[0]}` : option.colors.join(', ')})`,
                  }}
                />
                {option.id === 'classic' ? t('classic') : option.name}
              </button>
            ))}
          </div>
        </fieldset>
        <fieldset className="space-y-3">
          <legend className="mb-1 font-medium text-white">{t('glyph')}</legend>
          <div className="flex flex-wrap gap-2">
            {glyphs.map((option, index) => (
              <button
                key={option.id}
                type="button"
                aria-label={option.image ? option.name : t('none')}
                aria-pressed={glyphIndex === index}
                onClick={() => setGlyphIndex(index)}
                className={cn(
                  choiceClass,
                  'min-w-12',
                  glyphIndex === index ? selectedClass : 'border-white/15',
                )}
              >
                {option.image ? (
                  <Image
                    src={option.image}
                    alt=""
                    width={24}
                    height={24}
                    className={`[image-rendering:pixelated] ${styles.glyph}`}
                  />
                ) : (
                  t('none')
                )}
              </button>
            ))}
          </div>
        </fieldset>
        <fieldset className="space-y-3">
          <legend className="mb-1 font-medium text-white">{t('skin')}</legend>
          <div className="flex flex-wrap gap-2">
            {skins.map((option, index) => (
              <button
                key={option.name}
                type="button"
                aria-pressed={skinIndex === index}
                onClick={() => setSkinIndex(index)}
                className={cn(
                  choiceClass,
                  skinIndex === index ? selectedClass : 'border-white/15',
                )}
              >
                {option.name}
              </button>
            ))}
          </div>
        </fieldset>
        <p className="max-w-sm text-sm leading-relaxed text-white/65">
          {t('ownSkin')}
        </p>
        <p className="text-xs leading-relaxed text-white/45">{t('demo')}</p>
      </div>
    </div>
  )
}
