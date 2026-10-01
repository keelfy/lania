'use client'

import { cn } from '@/lib/utils'
import { useTranslations } from 'next-intl'
import { useLandingNavigation } from './landing-navigation-ctx'
import styles from './landing-motion.module.css'

export default function LandingSidebar() {
  const { activeSection } = useLandingNavigation()
  const t = useTranslations('landing')
  const sections = [
    { id: 'section-1', label: 'Lania' },
    { id: 'section-2', label: t('section2.title') },
    { id: 'section-customization', label: t('customization.title') },
    { id: 'section-3', label: t('section2.seeMore') },
    { id: 'section-4', label: t('section4.title') },
  ]
  const activeIndex = Math.max(
    0,
    sections.findIndex((section) => section.id === activeSection),
  )
  return (
    <nav
      aria-label={t('navigation')}
      className="fixed top-1/2 right-8 z-20 hidden -translate-y-1/2 rounded-full border border-white/10 bg-black/20 p-2 backdrop-blur-sm xl:block"
    >
      <div className="relative flex flex-col">
        <span
          aria-hidden
          className={cn(
            'pointer-events-none absolute top-3 left-3 size-2 rounded-full bg-teal-300 shadow-[0_0_12px_#5eead4]',
            styles.indicator,
          )}
          style={{ transform: `translateY(${activeIndex * 32}px)` }}
        />
        {sections.map((section) => (
          <button
            key={section.id}
            type="button"
            aria-label={section.label}
            title={section.label}
            aria-current={activeSection === section.id ? 'location' : undefined}
            className="flex size-8 items-center justify-center rounded-full focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-teal-300"
            onClick={() =>
              document.getElementById(section.id)?.scrollIntoView({
                block: 'start',
                behavior: window.matchMedia('(prefers-reduced-motion: reduce)')
                  .matches
                  ? 'instant'
                  : 'smooth',
              })
            }
          >
            <span aria-hidden className="size-2 rounded-full bg-white/30" />
          </button>
        ))}
      </div>
    </nav>
  )
}
