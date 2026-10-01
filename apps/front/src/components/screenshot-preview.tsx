import { AspectRatio } from '@/components/ui/aspect-ratio'
import { DialogTrigger } from '@/components/ui/dialog'
import { cn } from '@/lib/utils'
import Image from 'next/image'
import type { ReactNode } from 'react'
import styles from './screenshot-preview.module.css'

type Props = {
  src: string
  alt: string
  title?: string
  authors?: string[]
  sizes: string
  children?: ReactNode
}

export default function ScreenshotPreview({
  src,
  alt,
  title,
  authors,
  sizes,
  children,
}: Props) {
  return (
    <DialogTrigger
      aria-label={alt}
      className={cn(
        'w-full overflow-hidden rounded-2xl focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-teal-300',
        styles.galleryTrigger,
      )}
    >
      <AspectRatio ratio={16 / 9} className="relative">
        <Image
          src={src}
          alt={alt}
          fill
          sizes={sizes}
          className={cn('rounded-2xl object-cover', styles.galleryImage)}
        />
        {title || authors?.length ? (
          <div className={styles.galleryCaption}>
            {title && <p className="font-medium">{title}</p>}
            {authors?.length ? (
              <p className="mt-1 text-sm text-white/75">{authors.join(', ')}</p>
            ) : null}
          </div>
        ) : null}
        {children}
      </AspectRatio>
    </DialogTrigger>
  )
}
