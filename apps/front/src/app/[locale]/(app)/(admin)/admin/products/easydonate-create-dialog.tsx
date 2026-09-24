'use client'

import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import {
  Field,
  FieldDescription,
  FieldGroup,
  FieldLabel,
} from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { createEasyDonateProduct } from '@/lib/api-endpoints'
import { clientApiFetcher } from '@/lib/client'
import { errorToast } from '@/lib/toasts'
import { AdminProduct } from '@/models/admin'
import { useTranslations } from 'next-intl'
import Image from 'next/image'
import React from 'react'
import { toast } from 'sonner'
import { useEdCredentials } from './ed-credentials-context'

type ProductInfo = {
  name: string
  description: string
  priceName: AdminProduct['priceName']
}

type Props = {
  disabled?: boolean
  getProductInfo: () => ProductInfo
  onCreated: (easyDonateProductId: number) => void
}

// The "create in EasyDonate" button next to the ED ID field. It only creates the remote position
// and reports its id back; the admin still has to press "save" on the product form themselves.
export default function EasyDonateCreateDialog({
  disabled,
  getProductInfo,
  onCreated,
}: Props) {
  const t = useTranslations('admin.products.easydonate')
  const { hasCredentials, getCredentials, setCredentials } = useEdCredentials()
  const [open, setOpen] = React.useState(false)
  const [editingCredentials, setEditingCredentials] =
    React.useState(!hasCredentials)
  const [productInfo, setProductInfo] = React.useState<ProductInfo>()
  const [preview, setPreview] = React.useState<string>()
  const [isPending, startTransition] = React.useTransition()
  const objectUrlRef = React.useRef<string>(undefined)
  const imageInputRef = React.useRef<HTMLInputElement>(null)

  React.useEffect(() => {
    return () => {
      if (objectUrlRef.current) URL.revokeObjectURL(objectUrlRef.current)
    }
  }, [])

  function handleOpenChange(next: boolean) {
    setOpen(next)
    if (next) {
      setEditingCredentials(!hasCredentials)
      setProductInfo(getProductInfo())
    }
  }

  function handleImageChange(event: React.ChangeEvent<HTMLInputElement>) {
    if (objectUrlRef.current) URL.revokeObjectURL(objectUrlRef.current)
    const file = event.target.files?.[0]
    if (!file) {
      setPreview(undefined)
      objectUrlRef.current = undefined
      return
    }
    const objectUrl = URL.createObjectURL(file)
    objectUrlRef.current = objectUrl
    setPreview(objectUrl)
  }

  function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (isPending || !productInfo) return

    const form = new FormData(event.currentTarget)
    const userAuth = editingCredentials
      ? String(form.get('userAuth') ?? '').trim()
      : (getCredentials()?.userAuth ?? '')
    if (!userAuth) return
    const image = imageInputRef.current?.files?.[0]

    startTransition(async () => {
      try {
        const result = await createEasyDonateProduct(clientApiFetcher, {
          userAuth,
          name: productInfo.name,
          description: productInfo.description,
          priceName: productInfo.priceName,
          image,
        })
        setCredentials({ userAuth })
        onCreated(result.easyDonateProductId)
        toast.success(t('created'))
        setOpen(false)
      } catch (error) {
        errorToast(t('createFailed'), error)
      }
    })
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogTrigger asChild>
        <Button type="button" variant="outline" disabled={disabled}>
          {t('button')}
        </Button>
      </DialogTrigger>
      <DialogContent>
        <form onSubmit={handleSubmit} className="flex flex-col gap-4">
          <DialogHeader>
            <DialogTitle>{t('title')}</DialogTitle>
            <DialogDescription>{t('description')}</DialogDescription>
          </DialogHeader>
          <FieldGroup>
            {editingCredentials ? (
              <fieldset className="grid gap-3 rounded-md border p-3">
                <Field>
                  <FieldLabel htmlFor="ed-user-auth">
                    {t('userAuth')}
                  </FieldLabel>
                  <Input
                    id="ed-user-auth"
                    name="userAuth"
                    type="password"
                    autoComplete="off"
                    required
                  />
                  <FieldDescription>{t('userAuthHint')}</FieldDescription>
                </Field>
              </fieldset>
            ) : (
              <div className="flex items-center justify-between rounded-md border p-3">
                <FieldDescription>{t('credentialsKept')}</FieldDescription>
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  onClick={() => setEditingCredentials(true)}
                >
                  {t('changeCredentials')}
                </Button>
              </div>
            )}
            <Field>
              <FieldLabel htmlFor="ed-image">{t('image')}</FieldLabel>
              <input
                id="ed-image"
                ref={imageInputRef}
                type="file"
                accept="image/png,image/jpeg"
                onChange={handleImageChange}
                className="text-sm"
              />
              <FieldDescription>{t('imageHint')}</FieldDescription>
              {preview && (
                <Image
                  src={preview}
                  alt=""
                  width={64}
                  height={64}
                  unoptimized
                  className="border-border mt-2 rounded-md border object-cover"
                />
              )}
            </Field>
            {productInfo && (
              <div className="text-muted-foreground rounded-md border border-dashed p-3 text-xs">
                <p className="mb-1 font-medium">{t('summary')}</p>
                <p>{productInfo.name}</p>
                <p>{productInfo.description}</p>
                <p>{productInfo.priceName}</p>
              </div>
            )}
          </FieldGroup>
          <DialogFooter>
            <Button type="submit" disabled={isPending}>
              {t('submit')}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
