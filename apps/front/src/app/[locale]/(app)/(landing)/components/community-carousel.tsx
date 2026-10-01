'use client'

import { AspectRatio } from '@/components/ui/aspect-ratio'
import {
  Carousel,
  CarouselContent,
  CarouselItem,
  CarouselNext,
  CarouselPrevious,
  type CarouselApi,
} from '@/components/ui/carousel'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { useEffect, useState } from 'react'
import ScreenshotPreview from '@/components/screenshot-preview'
import { Label } from '@/components/ui/label'
import { useMediaQuery } from '@/lib/use-media-query'
import Autoplay from 'embla-carousel-autoplay'
import Image from 'next/image'
import { Noto_Sans } from 'next/font/google'

const notoSans = Noto_Sans({
  subsets: ['latin'],
})

type LandingCommunityCarouselProps = {
  gallery: {
    src: string
    alt: string
    name: string
    authors?: string[]
  }[]
}

export default function LandingCommunityCarousel({
  gallery,
}: LandingCommunityCarouselProps) {
  const reducedMotion = useMediaQuery('(prefers-reduced-motion: reduce)')
  const isDesktop = useMediaQuery('(min-width: 1024px)')
  const [api, setApi] = useState<CarouselApi>()
  const [autoplay] = useState(() =>
    Autoplay({
      delay: 5000,
      playOnInit: false,
      stopOnMouseEnter: true,
      stopOnFocusIn: true,
      stopOnInteraction: false,
      breakpoints: {
        '(prefers-reduced-motion: reduce)': { active: false },
      },
    }),
  )
  useEffect(() => {
    if (!api) return
    const motion = window.matchMedia('(prefers-reduced-motion: reduce)')
    const sync = () => {
      const root = api.rootNode()
      if (
        motion.matches ||
        document.hidden ||
        root.matches(':hover') ||
        root.contains(document.activeElement)
      )
        autoplay.stop()
      else autoplay.play()
    }
    sync()
    api.on('reInit', sync)
    motion.addEventListener('change', sync)
    document.addEventListener('visibilitychange', sync)
    return () => {
      motion.removeEventListener('change', sync)
      document.removeEventListener('visibilitychange', sync)
      autoplay.stop()
      api.off('reInit', sync)
    }
  }, [api, autoplay])
  return (
    <Carousel
      setApi={setApi}
      opts={{ loop: true, duration: reducedMotion ? 0 : 25 }}
      plugins={[autoplay]}
    >
      <CarouselContent>
        {gallery.map((item) => (
          <Dialog key={item.alt}>
            <CarouselItem className="lg:basis-1/2">
              <ScreenshotPreview
                src={item.src}
                alt={item.alt}
                title={item.alt}
                authors={item.authors}
                sizes="(min-width: 1024px) 480px, (min-width: 768px) 90vw, 100vw"
              >
                <div className="bg-accent/50 absolute top-0 right-0 m-2 rounded-sm px-2 py-1">
                  <Label
                    className={`text-sm font-bold ${notoSans.className} antialiased`}
                  >
                    {item.name}
                  </Label>
                </div>
              </ScreenshotPreview>
            </CarouselItem>
            <DialogContent className="border-0 p-0 md:max-w-5xl">
              <DialogHeader className="hidden">
                <DialogTitle>Detailed view</DialogTitle>
              </DialogHeader>
              <AspectRatio ratio={16 / 9} className="relative">
                <Image
                  src={item.src}
                  alt={item.alt}
                  fill
                  sizes="(min-width: 1024px) 1024px, 100vw"
                  className="rounded-2xl object-cover"
                />
                <div className="bg-accent/50 absolute right-0 bottom-0 m-4 rounded-sm px-3 py-1">
                  <Label
                    className={`text-md font-bold ${notoSans.className} antialiased`}
                  >
                    {item.name}
                  </Label>
                </div>
              </AspectRatio>
            </DialogContent>
          </Dialog>
        ))}
      </CarouselContent>
      {isDesktop ? <CarouselPrevious /> : null}
      {isDesktop ? <CarouselNext /> : null}
    </Carousel>
  )
}
