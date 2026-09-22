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
          <DialogContent className="border-0 p-0 md:max-w-4xl">
            <DialogHeader className="hidden">
              <DialogTitle>{screenshot.title ?? 'Screenshot'}</DialogTitle>
            </DialogHeader>
            <AspectRatio ratio={16 / 9} className="relative">
              <Image
                src={screenshot.image}
                alt={screenshot.title ?? ''}
                fill
                className="rounded-2xl object-cover"
              />
            </AspectRatio>
            {(screenshot.title || screenshot.authors.length > 0) && (
              <div className="flex flex-col gap-1 px-4 pb-4">
                {screenshot.title && (
                  <p className="font-medium">{screenshot.title}</p>
                )}
                {screenshot.authors.length > 0 && (
                  <p className="text-muted-foreground text-sm">
                    {screenshot.authors.map((author, index) => (
                      <span key={author.id}>
                        {index > 0 && ', '}
                        <Link
                          href={communityProfileHref(
                            locale,
                            author.username,
                            seasonId,
                          )}
                          className="hover:underline"
                        >
                          {author.username}
                        </Link>
                      </span>
                    ))}
                  </p>
                )}
              </div>
            )}
          </DialogContent>
        </Dialog>
      ))}
    </div>
  )
}
