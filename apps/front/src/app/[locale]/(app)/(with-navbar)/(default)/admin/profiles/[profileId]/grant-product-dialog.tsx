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
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { grantProduct } from '@/lib/api-endpoints'
import { clientApiFetcher } from '@/lib/client'
import { errorToast } from '@/lib/toasts'
import { Product, ProductMetadata } from '@/models/product'
import { Season } from '@/models/season'
import { PlusIcon } from 'lucide-react'
import { useTranslations } from 'next-intl'
import { useRouter } from 'next/navigation'
import React from 'react'
import { toast } from 'sonner'

type Props = {
  profileId: string
  seasons: Season[]
  products: Product<ProductMetadata>[]
}

export default function GrantProductDialog({
  profileId,
  seasons,
  products,
}: Props) {
  const t = useTranslations('admin.profiles.grants.grantDialog')
  const router = useRouter()
  const [open, setOpen] = React.useState(false)
  const [productId, setProductId] = React.useState('')
  const [seasonId, setSeasonId] = React.useState(
    seasons.find((season) => season.isActive)?.id ?? '',
  )
  const [isPending, startTransition] = React.useTransition()

  const handleSubmit = (event: React.FormEvent) => {
    event.preventDefault()
    if (isPending || !productId || !seasonId) return

    startTransition(async () => {
      try {
        await grantProduct(clientApiFetcher, profileId, { productId, seasonId })
        toast.success(t('granted'))
        setProductId('')
        setOpen(false)
        router.refresh()
      } catch (error) {
        errorToast(t('failed'), error)
      }
    })
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button size="sm">
          <PlusIcon className="size-4" />
          {t('open')}
        </Button>
      </DialogTrigger>
      <DialogContent>
        <form onSubmit={handleSubmit} className="flex flex-col gap-4">
          <DialogHeader>
            <DialogTitle>{t('title')}</DialogTitle>
            <DialogDescription>{t('description')}</DialogDescription>
          </DialogHeader>
          <div className="flex flex-col gap-2">
            <Label>{t('product')}</Label>
            <Select value={productId} onValueChange={setProductId}>
              <SelectTrigger className="w-full">
                <SelectValue placeholder={t('selectProduct')} />
              </SelectTrigger>
              <SelectContent>
                {products.map((product) => (
                  <SelectItem key={product.id} value={product.id}>
                    {product.name} ({product.category})
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div className="flex flex-col gap-2">
            <Label>{t('season')}</Label>
            <Select value={seasonId} onValueChange={setSeasonId}>
              <SelectTrigger className="w-full">
                <SelectValue placeholder={t('selectSeason')} />
              </SelectTrigger>
              <SelectContent>
                {seasons.map((season) => (
                  <SelectItem key={season.id} value={season.id}>
                    {season.name}
                    {season.isActive ? ` (${t('activeSeason')})` : ''}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <DialogFooter>
            <Button
              type="submit"
              disabled={isPending || !productId || !seasonId}
            >
              {t('submit')}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
