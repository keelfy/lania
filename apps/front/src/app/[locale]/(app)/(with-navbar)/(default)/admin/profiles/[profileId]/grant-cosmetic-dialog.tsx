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
import { grantCosmetic } from '@/lib/api-endpoints'
import { clientApiFetcher } from '@/lib/client'
import { errorToast } from '@/lib/toasts'
import {
  AdminCosmeticsCatalog,
  AdminNameColor,
  GrantCosmeticReq,
} from '@/models/admin'
import { Season } from '@/models/season'
import { SparklesIcon } from 'lucide-react'
import { useTranslations } from 'next-intl'
import Image from 'next/image'
import { useRouter } from 'next/navigation'
import React from 'react'
import { toast } from 'sonner'

type Props = {
  profileId: string
  seasons: Season[]
  // Missing when the catalog could not be loaded.
  catalog: AdminCosmeticsCatalog | undefined
}

// SelectItem cannot hold an empty value, so "for good" gets a value of its own.
const PERMANENT = 'permanent'

function ColorSwatch({ nameColor }: { nameColor: AdminNameColor }) {
  const background =
    nameColor.colors.length > 1
      ? `linear-gradient(90deg, ${nameColor.colors.join(', ')})`
      : nameColor.colors[0]

  return (
    <span
      className="border-border inline-block h-3 w-8 rounded-sm border"
      style={{ background }}
    />
  )
}

export default function GrantCosmeticDialog({
  profileId,
  seasons,
  catalog,
}: Props) {
  const t = useTranslations('admin.profiles.grants.cosmeticDialog')
  const router = useRouter()
  const [open, setOpen] = React.useState(false)
  const [type, setType] = React.useState<GrantCosmeticReq['type']>('name-color')
  const [itemId, setItemId] = React.useState('')
  const [prefixType, setPrefixType] =
    React.useState<NonNullable<GrantCosmeticReq['prefixType']>>('glyth')
  const [seasonId, setSeasonId] = React.useState(PERMANENT)
  const [isPending, startTransition] = React.useTransition()

  const handleTypeChange = (value: string) => {
    setType(value as GrantCosmeticReq['type'])
    setItemId('')
  }

  const handleSubmit = (event: React.FormEvent) => {
    event.preventDefault()
    if (isPending || !itemId) return

    startTransition(async () => {
      try {
        await grantCosmetic(clientApiFetcher, profileId, {
          type,
          itemId,
          prefixType: type === 'name-prefix' ? prefixType : undefined,
          seasonId: seasonId === PERMANENT ? undefined : seasonId,
        })
        toast.success(t('granted'))
        setItemId('')
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
        <Button size="sm" variant="outline" disabled={!catalog}>
          <SparklesIcon className="size-4" />
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
            <Label>{t('type')}</Label>
            <Select value={type} onValueChange={handleTypeChange}>
              <SelectTrigger className="w-full">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="name-color">
                  {t('types.name-color')}
                </SelectItem>
                <SelectItem value="name-prefix">
                  {t('types.name-prefix')}
                </SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div className="flex flex-col gap-2">
            <Label>{t('item')}</Label>
            <Select value={itemId} onValueChange={setItemId}>
              <SelectTrigger className="w-full">
                <SelectValue placeholder={t('selectItem')} />
              </SelectTrigger>
              <SelectContent>
                {type === 'name-color'
                  ? catalog?.nameColors.map((nameColor) => (
                      <SelectItem key={nameColor.id} value={nameColor.id}>
                        <ColorSwatch nameColor={nameColor} />
                        {nameColor.name}
                      </SelectItem>
                    ))
                  : catalog?.namePrefixes.map((namePrefix) => (
                      <SelectItem key={namePrefix.id} value={namePrefix.id}>
                        {namePrefix.image && (
                          <Image
                            src={namePrefix.image}
                            alt=""
                            width={16}
                            height={16}
                            unoptimized
                          />
                        )}
                        {namePrefix.name}
                      </SelectItem>
                    ))}
              </SelectContent>
            </Select>
          </div>
          {type === 'name-prefix' && (
            <div className="flex flex-col gap-2">
              <Label>{t('prefixType')}</Label>
              <Select
                value={prefixType}
                onValueChange={(value) =>
                  setPrefixType(value as typeof prefixType)
                }
              >
                <SelectTrigger className="w-full">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="glyth">
                    {t('prefixTypes.glyth')}
                  </SelectItem>
                  <SelectItem value="special">
                    {t('prefixTypes.special')}
                  </SelectItem>
                </SelectContent>
              </Select>
              <p className="text-muted-foreground text-xs">
                {t('prefixTypeHint')}
              </p>
            </div>
          )}
          <div className="flex flex-col gap-2">
            <Label>{t('season')}</Label>
            <Select value={seasonId} onValueChange={setSeasonId}>
              <SelectTrigger className="w-full">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value={PERMANENT}>{t('permanent')}</SelectItem>
                {seasons.map((season) => (
                  <SelectItem key={season.id} value={season.id}>
                    {season.name}
                    {season.isPrimary ? ` (${t('primarySeason')})` : ''}
                    {season.isActive ? ` (${t('activeSeason')})` : ''}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <DialogFooter>
            <Button type="submit" disabled={isPending || !itemId}>
              {t('submit')}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
