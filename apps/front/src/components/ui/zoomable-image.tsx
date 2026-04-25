'use client'

import Image from 'next/image'
import { DetailedHTMLProps, ImgHTMLAttributes } from 'react'
import { Dialog, DialogContent, DialogTrigger } from './dialog'

export default function ZoomableImage({
  src,
  alt,
  className,
}: DetailedHTMLProps<ImgHTMLAttributes<HTMLImageElement>, HTMLImageElement>) {
  if (!src) return null
  return (
    <Dialog>
      <DialogTrigger asChild>
        <Image
          src={src as string}
          alt={alt || ''}
          sizes="100vw"
          className={className}
          style={{
            width: '100%',
            height: 'auto',
          }}
          width={500}
          height={100}
          unoptimized
        />
      </DialogTrigger>
      <DialogContent className="max-w-7xl border-0 bg-transparent p-0">
        <div className="relative h-[calc(100vh-220px)] w-full overflow-clip rounded-md bg-transparent shadow-md">
          <Image
            src={src as string}
            fill
            alt={alt || ''}
            className="h-full w-full object-contain"
            unoptimized
          />
        </div>
      </DialogContent>
    </Dialog>
  )
}
