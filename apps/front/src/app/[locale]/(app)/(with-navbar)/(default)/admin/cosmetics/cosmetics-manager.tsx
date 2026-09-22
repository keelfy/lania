'use client'

import McUsername from '@/components/ui/mc-username'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Field,
  FieldDescription,
  FieldGroup,
  FieldLabel,
} from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import {
  createNameColor,
  createNamePrefix,
  updateNameColor,
  updateNamePrefix,
} from '@/lib/api-endpoints'
import { clientApiFetcher } from '@/lib/client'
import { errorToast } from '@/lib/toasts'
import {
  AdminCosmeticsCatalog,
  AdminNameColor,
  AdminNamePrefix,
} from '@/models/admin'
import { PaletteIcon, PlusIcon, ShapesIcon } from 'lucide-react'
import { useTranslations } from 'next-intl'
import Image from 'next/image'
import { useRouter } from 'next/navigation'
import React from 'react'
import { toast } from 'sonner'

type Selection =
  | { type: 'color'; item?: AdminNameColor }
  | { type: 'prefix'; item?: AdminNamePrefix }

export default function CosmeticsManager({
  catalog,
}: {
  catalog: AdminCosmeticsCatalog
}) {
  const t = useTranslations('admin.cosmetics')
  const router = useRouter()
  const [selection, setSelection] = React.useState<Selection>({
    type: 'color',
  })
  const [search, setSearch] = React.useState('')
  const [isPending, startTransition] = React.useTransition()
  const [noSpace, setNoSpace] = React.useState(false)

  const query = search.trim().toLocaleLowerCase()
  const colors = catalog.nameColors.filter((item) =>
    item.name.toLocaleLowerCase().includes(query),
  )
  const prefixes = catalog.namePrefixes.filter((item) =>
    item.name.toLocaleLowerCase().includes(query),
  )
  const key = `${selection.type}-${selection.item?.id ?? 'new'}`

  function submit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (isPending) return
    const form = new FormData(event.currentTarget)
    startTransition(async () => {
      try {
        if (selection.type === 'color') {
          const payload = {
            name: String(form.get('name') ?? '').trim(),
            colors: String(form.get('colors') ?? '')
              .split(',')
              .map((value) => value.trim())
              .filter(Boolean),
          }
          if (selection.item)
            await updateNameColor(clientApiFetcher, selection.item.id, payload)
          else await createNameColor(clientApiFetcher, payload)
        } else {
          const payload = {
            name: String(form.get('name') ?? '').trim(),
            prefix: String(form.get('prefix') ?? '').trim(),
            image: String(form.get('image') ?? '').trim(),
            noSpace,
          }
          if (selection.item)
            await updateNamePrefix(clientApiFetcher, selection.item.id, payload)
          else await createNamePrefix(clientApiFetcher, payload)
        }
        toast.success(t(selection.item ? 'updated' : 'created'))
        router.refresh()
      } catch (error) {
        errorToast(t('saveFailed'), error)
      }
    })
  }

  return (
    <div className="grid gap-6 lg:grid-cols-[minmax(18rem,0.85fr)_minmax(26rem,1.4fr)]">
      <section className="border-border overflow-hidden rounded-lg border">
        <div className="bg-muted/40 border-b p-3">
          <Input
            value={search}
            onChange={(event) => setSearch(event.target.value)}
            placeholder={t('search')}
          />
        </div>
        <div className="max-h-[65vh] overflow-y-auto p-2">
          <CatalogGroup
            title={t('colors')}
            icon={<PaletteIcon />}
            items={colors}
            onSelect={(item) => setSelection({ type: 'color', item })}
            onCreate={() => {
              setSelection({ type: 'color' })
              setNoSpace(false)
            }}
            selected={
              selection.type === 'color' ? selection.item?.id : undefined
            }
          />
          <CatalogGroup
            title={t('prefixes')}
            icon={<ShapesIcon />}
            items={prefixes}
            onSelect={(item) => {
              setSelection({ type: 'prefix', item })
              setNoSpace(item.noSpace)
            }}
            onCreate={() => {
              setSelection({ type: 'prefix' })
              setNoSpace(false)
            }}
            selected={
              selection.type === 'prefix' ? selection.item?.id : undefined
            }
          />
        </div>
      </section>

      <section className="border-border rounded-lg border p-5">
        <div className="mb-6 flex items-start justify-between gap-3">
          <div>
            <h2 className="text-xl font-bold">
              {selection.item ? t('edit') : t('create')}
            </h2>
            <p className="text-muted-foreground text-sm">
              {t(selection.type === 'color' ? 'colorHint' : 'prefixHint')}
            </p>
          </div>
          <Badge variant="outline">
            {t(selection.type === 'color' ? 'color' : 'prefix')}
          </Badge>
        </div>
        <form
          key={key}
          onSubmit={submit}
          className="grid gap-6 xl:grid-cols-[1fr_15rem]"
        >
          <FieldGroup>
            <Field>
              <FieldLabel htmlFor={`${key}-name`}>
                {t('fields.name')}
              </FieldLabel>
              <Input
                id={`${key}-name`}
                name="name"
                required
                maxLength={255}
                defaultValue={selection.item?.name}
              />
            </Field>
            {selection.type === 'color' ? (
              <ColorFields item={selection.item} t={t} />
            ) : (
              <PrefixFields
                item={selection.item}
                t={t}
                noSpace={noSpace}
                setNoSpace={setNoSpace}
              />
            )}
            <Button disabled={isPending} className="w-fit">
              {t('save')}
            </Button>
          </FieldGroup>
          <CosmeticPreview selection={selection} t={t} />
        </form>
      </section>
    </div>
  )
}

function CatalogGroup<T extends { id: string; name: string }>({
  title,
  icon,
  items,
  onSelect,
  onCreate,
  selected,
}: {
  title: string
  icon: React.ReactNode
  items: T[]
  onSelect: (item: T) => void
  onCreate: () => void
  selected?: string
}) {
  return (
    <div className="mb-4">
      <h3 className="text-muted-foreground flex items-center gap-2 px-2 py-2 text-sm font-semibold">
        {icon}
        <span className="flex-1">{title}</span>
        <Badge variant="secondary">{items.length}</Badge>
        <Button
          type="button"
          size="icon"
          variant="ghost"
          className="size-7"
          onClick={onCreate}
          aria-label={title}
        >
          <PlusIcon />
        </Button>
      </h3>
      <div className="grid gap-1">
        {items.map((item) => (
          <button
            key={item.id}
            onClick={() => onSelect(item)}
            className={`rounded-md px-3 py-2 text-left text-sm ${selected === item.id ? 'bg-primary text-primary-foreground' : 'hover:bg-muted'}`}
          >
            {item.name}
          </button>
        ))}
      </div>
    </div>
  )
}

function ColorFields({
  item,
  t,
}: {
  item?: AdminNameColor
  t: ReturnType<typeof useTranslations>
}) {
  return (
    <Field>
      <FieldLabel htmlFor="cosmetic-colors">{t('fields.colors')}</FieldLabel>
      <Input
        id="cosmetic-colors"
        name="colors"
        defaultValue={item?.colors.join(', ')}
        placeholder="#22c55e, #16a34a"
      />
      <FieldDescription>{t('colorsHint')}</FieldDescription>
    </Field>
  )
}

function PrefixFields({
  item,
  t,
  noSpace,
  setNoSpace,
}: {
  item?: AdminNamePrefix
  t: ReturnType<typeof useTranslations>
  noSpace: boolean
  setNoSpace: (value: boolean) => void
}) {
  return (
    <>
      <Field>
        <FieldLabel htmlFor="cosmetic-prefix">{t('fields.prefix')}</FieldLabel>
        <Input
          id="cosmetic-prefix"
          name="prefix"
          required
          defaultValue={item?.prefix}
        />
      </Field>
      <Field>
        <FieldLabel htmlFor="cosmetic-image">{t('fields.image')}</FieldLabel>
        <Input
          id="cosmetic-image"
          name="image"
          type="url"
          required
          defaultValue={item?.image}
        />
      </Field>
      <Field orientation="horizontal">
        <Switch
          checked={noSpace}
          onCheckedChange={setNoSpace}
          id="cosmetic-no-space"
        />
        <div>
          <FieldLabel htmlFor="cosmetic-no-space">
            {t('fields.noSpace')}
          </FieldLabel>
          <FieldDescription>{t('noSpaceHint')}</FieldDescription>
        </div>
      </Field>
    </>
  )
}

function CosmeticPreview({
  selection,
  t,
}: {
  selection: Selection
  t: ReturnType<typeof useTranslations>
}) {
  return (
    <aside className="bg-muted/40 flex min-h-40 flex-col items-center justify-center gap-3 rounded-lg border p-5">
      <span className="text-muted-foreground text-xs">{t('preview')}</span>
      {selection.type === 'color' ? (
        <McUsername
          username="Keelfy"
          colors={selection.item?.colors}
          className="text-2xl"
        />
      ) : (
        <div className="flex items-center gap-2">
          {selection.item?.image && (
            <Image
              src={selection.item.image}
              alt=""
              width={32}
              height={32}
              unoptimized
            />
          )}
          <McUsername username="Keelfy" className="text-2xl" />
        </div>
      )}
    </aside>
  )
}
