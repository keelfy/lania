'use client'

import { Button } from '@/components/ui/button'
import { FieldDescription } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { UploadedImage } from '@/models/admin'
import { ImageUpIcon, Loader2Icon } from 'lucide-react'
import Image from 'next/image'
import { useTranslations } from 'next-intl'
import React from 'react'

type Props = {
  id: string
  value: string
  onChange: (location: string) => void
  upload: (file: File) => Promise<UploadedImage>
  accept?: string
  previewSize?: number
}

// A file picker + clipboard-paste drop zone for an admin image field. Uploads go straight to the
// project's S3 bucket through the API; the text input underneath stays as a manual override for
// rows that still point at an external host and as an escape hatch if an upload fails.
export default function ImageUploadField({
  id,
  value,
  onChange,
  upload,
  accept = 'image/png,image/jpeg',
  previewSize = 64,
}: Props) {
  const t = useTranslations('admin.uploads')
  const [isPending, startTransition] = React.useTransition()
  const [preview, setPreview] = React.useState<string>()
  const [error, setError] = React.useState<string>()
  const objectUrlRef = React.useRef<string>(undefined)
  const inputRef = React.useRef<HTMLInputElement>(null)

  React.useEffect(() => {
    return () => {
      if (objectUrlRef.current) URL.revokeObjectURL(objectUrlRef.current)
    }
  }, [])

  const handleFile = (file: File) => {
    if (objectUrlRef.current) URL.revokeObjectURL(objectUrlRef.current)
    const objectUrl = URL.createObjectURL(file)
    objectUrlRef.current = objectUrl
    setPreview(objectUrl)
    setError(undefined)

    startTransition(async () => {
      try {
        const result = await upload(file)
        onChange(result.location)
      } catch (err) {
        setError(err instanceof Error ? err.message : t('uploadFailed'))
      } finally {
        URL.revokeObjectURL(objectUrl)
        objectUrlRef.current = undefined
        setPreview(undefined)
      }
    })
  }

  const handlePaste = (event: React.ClipboardEvent) => {
    const file = Array.from(event.clipboardData.items)
      .find((item) => item.type.startsWith('image/'))
      ?.getAsFile()
    if (file) handleFile(file)
  }

  const displaySrc = preview ?? value

  return (
    <div className="flex flex-col gap-2" onPaste={handlePaste}>
      <div className="flex items-center gap-3">
        {displaySrc ? (
          <Image
            src={displaySrc}
            alt=""
            width={previewSize}
            height={previewSize}
            unoptimized={preview !== undefined}
            className="border-border shrink-0 rounded-md border object-cover"
            style={{ width: previewSize, height: previewSize }}
          />
        ) : (
          <div
            className="border-border bg-muted/40 text-muted-foreground flex shrink-0 items-center justify-center rounded-md border border-dashed"
            style={{ width: previewSize, height: previewSize }}
          >
            <ImageUpIcon className="size-5" />
          </div>
        )}
        <div className="flex flex-1 flex-col gap-1.5">
          <div className="flex items-center gap-2">
            <Button
              type="button"
              variant="outline"
              size="sm"
              disabled={isPending}
              onClick={() => inputRef.current?.click()}
            >
              {isPending ? (
                <Loader2Icon className="animate-spin" />
              ) : (
                <ImageUpIcon />
              )}
              {t('choose')}
            </Button>
            <input
              ref={inputRef}
              type="file"
              accept={accept}
              className="hidden"
              onChange={(event) => {
                const file = event.target.files?.[0]
                event.target.value = ''
                if (file) handleFile(file)
              }}
            />
          </div>
          <FieldDescription>{t('orPaste')}</FieldDescription>
        </div>
      </div>
      <Input
        id={id}
        value={value}
        onChange={(event) => onChange(event.target.value.trim())}
        placeholder="s3://bucket/key.png"
      />
      {error ? (
        <p className="text-destructive text-xs">{error}</p>
      ) : (
        <FieldDescription>{t('manualHint')}</FieldDescription>
      )}
    </div>
  )
}
