'use client'

import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import { ArrowUpIcon } from 'lucide-react'
import { useLandingNavigation } from './landing-navigation-ctx'

export default function LandingSidebar() {
  const { activeSection } = useLandingNavigation()

  return (
    <nav
      aria-label="Переход по разделам"
      className={cn(
        'fixed right-20 bottom-20 z-20 hidden flex-col gap-2 opacity-0 transition-opacity duration-300 xl:flex',
        activeSection !== 'section-1' && 'opacity-100',
      )}
    >
      <Button
        type="button"
        variant="outline"
        size="icon"
        className={cn(
          'size-12 rounded-full',
          activeSection === 'section-1' && 'pointer-events-none',
        )}
        onClick={() => {
          window.scrollTo({
            top: 0,
            behavior: 'smooth',
          })
        }}
      >
        <ArrowUpIcon className="size-8" />
      </Button>
    </nav>
  )
}
