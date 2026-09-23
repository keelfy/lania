'use client'

import { AspectRatio } from '@/components/ui/aspect-ratio'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import { communityProfileHref } from '../../community/community-href'
import { SeasonScreenshot } from '@/models/season'
import Image from 'next/image'
import Link from 'next/link'

type Props = {
  locale: string
  seasonId: string
  screenshots: SeasonScreenshot[]
}

export default function SeasonScreenshots({
  locale,
  seasonId,
  screenshots,
}: Props) {
  return (
    <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
      {screenshots.map((screenshot) => (
        <Dialog key={screenshot.id}>
          <DialogTrigger className="w-full">
            <AspectRatio
              ratio={16 / 9}
              className="relative overflow-hidden rounded-md"
            >
              <Image
                src={screenshot.image}
                alt={screenshot.title ?? ''}
                fill
                sizes="(max-width: 640px) 100vw, (max-width: 1024px) 50vw, 33vw"
                className="object-cover transition-transform duration-300 hover:scale-105"
              />
            </AspectRatio>
          </DialogTrigger>
          <DialogContent className="w-[min(calc(100vw-2rem),160dvh)] gap-0 overflow-hidden border-0 p-0 sm:max-w-4xl">
            <DialogHeader className="sr-only">
              <DialogTitle>{screenshot.title ?? 'Screenshot'}</DialogTitle>
            </DialogHeader>
            <AspectRatio ratio={16 / 9} className="relative">
              <Image
                src={screenshot.image}
                alt={screenshot.title ?? ''}
                fill
                sizes="(max-width: 896px) 100vw, 896px"
                className="object-cover"
              />
              {(screenshot.title || screenshot.authors.length > 0) && (
                <div className="absolute inset-x-0 bottom-0 bg-gradient-to-t from-black/90 via-black/65 to-transparent px-4 pt-10 pb-4 text-white sm:px-5 sm:pt-14 sm:pb-5">
                  {screenshot.title && (
                    <p className="font-medium">{screenshot.title}</p>
                  )}
                  {screenshot.authors.length > 0 && (
                    <p className="mt-1 text-sm text-white/80">
                      {screenshot.authors.map((author, index) => (
                        <span key={author.id}>
                          {index > 0 && ', '}
                          <Link
                            href={communityProfileHref(
                              locale,
                              author.username,
                              seasonId,
                            )}
                            className="rounded-sm hover:text-white hover:underline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-white"
                          >
                            {author.username}
                          </Link>
                        </span>
                      ))}
                    </p>
                  )}
                </div>
              )}
            </AspectRatio>
          </DialogContent>
        </Dialog>
      ))}
    </div>
  )
}
