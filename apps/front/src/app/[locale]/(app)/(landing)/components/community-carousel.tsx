'use client'

import { AspectRatio } from '@/components/ui/aspect-ratio'
import {
  Carousel,
  CarouselContent,
  CarouselItem,
  CarouselNext,
  CarouselPrevious,
} from '@/components/ui/carousel'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
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
  }[]
}

export default function LandingCommunityCarousel({
  gallery,
}: LandingCommunityCarouselProps) {
  const isDesktop = useMediaQuery('(min-width: 1024px)')
  return (
    <Carousel opts={{ loop: true }} plugins={[Autoplay({ delay: 3000 })]}>
      <CarouselContent>
        {gallery.map((item) => (
          <Dialog key={item.alt}>
            <CarouselItem className="lg:basis-1/2">
              <DialogTrigger className="w-full">
                <AspectRatio ratio={16 / 9} className="relative">
                  <Image
                    src={item.src}
                    alt={item.alt}
                    fill
                    className="rounded-2xl object-cover"
                  />
                  <div className="bg-accent/50 absolute right-0 bottom-0 m-2 rounded-sm px-2 py-1">
                    <Label className={`text-sm font-bold ${notoSans.className} antialiased`}>
                      {item.name}
                    </Label>
                  </div>
                </AspectRatio>
              </DialogTrigger>
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
                  className="rounded-2xl object-cover"
                />
                <div className="bg-accent/50 absolute right-0 bottom-0 m-4 rounded-sm px-3 py-1">
                  <Label className={`text-md font-bold ${notoSans.className} antialiased`}>
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
