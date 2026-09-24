'use client'

import McUsername from '@/components/ui/mc-username'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Field, FieldDescription, FieldLabel } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import { createAdminProduct, updateAdminProduct } from '@/lib/api-endpoints'
import { clientApiFetcher } from '@/lib/client'
import { errorToast } from '@/lib/toasts'
import {
  AdminCosmeticsCatalog,
  AdminProduct,
  SaveProduct,
} from '@/models/admin'
import { PlusIcon } from 'lucide-react'
import { useTranslations } from 'next-intl'
import Image from 'next/image'
import { useRouter } from 'next/navigation'
import React from 'react'
import { toast } from 'sonner'
import EasyDonateCreateDialog from './easydonate-create-dialog'

const priceByCategory = {
  upgrade: 'season_access',
  'name-color': 'name_color',
  'name-prefix': 'name_prefix',
} as const

export default function ProductsManager({
  products,
  catalog,
}: {
  products: AdminProduct[]
  catalog: AdminCosmeticsCatalog
}) {
  const t = useTranslations('admin.products')
  const router = useRouter()
  const [selected, setSelected] = React.useState<AdminProduct>()
  const [search, setSearch] = React.useState('')
  const [category, setCategory] =
    React.useState<AdminProduct['category']>('name-color')
  const [cosmeticId, setCosmeticId] = React.useState('')
  const [active, setActive] = React.useState(false)
  const [easyDonateProductId, setEasyDonateProductId] = React.useState('')
  const [ruFilled, setRuFilled] = React.useState(false)
  const [isPending, startTransition] = React.useTransition()
  const formRef = React.useRef<HTMLFormElement>(null)

  function chooseProduct(product?: AdminProduct) {
    setSelected(product)
    setCategory(product?.category ?? 'name-color')
    setCosmeticId(
      product?.metadata.nameColorId ?? product?.metadata.namePrefixId ?? '',
    )
    setActive(product?.isActive ?? false)
    setEasyDonateProductId(
      product?.easyDonateProductId ? String(product.easyDonateProductId) : '',
    )
    const ru = product?.localizations.find((item) => item.locale === 'ru')
    setRuFilled(Boolean(ru?.name.trim() && ru?.description.trim()))
  }

  function checkRuFilled(form: HTMLFormElement) {
    const data = new FormData(form)
    const name = String(data.get('name-ru') ?? '').trim()
    const description = String(data.get('description-ru') ?? '').trim()
    setRuFilled(Boolean(name && description))
  }

  function getProductInfo() {
    const data = formRef.current ? new FormData(formRef.current) : undefined
    return {
      name: String(data?.get('name-ru') ?? '').trim(),
      description: String(data?.get('description-ru') ?? '').trim(),
      priceName: priceByCategory[category],
    }
  }

  const filtered = products.filter((product) =>
    product.localizations.some((item) =>
      item.name.toLocaleLowerCase().includes(search.trim().toLocaleLowerCase()),
    ),
  )
  const ru = selected?.localizations.find((item) => item.locale === 'ru')
  const en = selected?.localizations.find((item) => item.locale === 'en')
  const cosmetics =
    category === 'name-color'
      ? catalog.nameColors
      : category === 'name-prefix'
        ? catalog.namePrefixes
        : []

  function submit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (isPending) return
    const form = new FormData(event.currentTarget)
    const edValue = easyDonateProductId.trim()
    const payload: SaveProduct = {
      category,
      cosmeticId: category === 'upgrade' ? undefined : cosmeticId,
      priceName: priceByCategory[category],
      isActive: active,
      easyDonateProductId: edValue ? Number(edValue) : undefined,
      localizations: [
        {
          locale: 'ru',
          name: String(form.get('name-ru') ?? '').trim(),
          description: String(form.get('description-ru') ?? '').trim(),
        },
        {
          locale: 'en',
          name: String(form.get('name-en') ?? '').trim(),
          description: String(form.get('description-en') ?? '').trim(),
        },
      ],
    }
    startTransition(async () => {
      try {
        if (selected)
          await updateAdminProduct(clientApiFetcher, selected.id, payload)
        else await createAdminProduct(clientApiFetcher, payload)
        toast.success(t(selected ? 'updated' : 'created'))
        router.refresh()
      } catch (error) {
        errorToast(t('saveFailed'), error)
      }
    })
  }

  return (
    <div className="flex flex-col gap-4">
      <div className="flex justify-end">
        <Button onClick={() => chooseProduct()}>
          <PlusIcon />
          {t('create')}
        </Button>
      </div>

      <div className="grid items-start gap-6 lg:grid-cols-[minmax(18rem,24rem)_minmax(0,1fr)]">
        <section className="border-border overflow-hidden rounded-lg border lg:sticky lg:top-6">
          <div className="bg-muted/40 border-b p-3">
            <Input
              value={search}
              onChange={(event) => setSearch(event.target.value)}
              placeholder={t('search')}
            />
          </div>
          <div className="max-h-[calc(100svh-14rem)] overflow-y-auto p-2">
            {filtered.length === 0 && (
              <p className="text-muted-foreground py-8 text-center text-sm">
                {t('empty')}
              </p>
            )}
            {filtered.map((product) => {
              const name =
                product.localizations.find((item) => item.locale === 'ru')
                  ?.name ?? product.id
              return (
                <button
                  key={product.id}
                  type="button"
                  onClick={() => chooseProduct(product)}
                  className={`mb-1 flex w-full items-center justify-between gap-3 rounded-md border px-3 py-2 text-left transition-colors ${selected?.id === product.id ? 'border-primary bg-accent' : 'hover:bg-muted border-transparent'}`}
                >
                  <span className="min-w-0">
                    <span className="block truncate text-sm font-medium">
                      {name}
                    </span>
                    <span className="text-muted-foreground text-xs">
                      {t(`categories.${product.category}`)}
                    </span>
                  </span>
                  <Badge variant={product.isActive ? 'default' : 'secondary'}>
                    {t(product.isActive ? 'active' : 'draft')}
                  </Badge>
                </button>
              )
            })}
          </div>
        </section>

        <section className="border-border rounded-lg border">
          <div className="flex flex-col gap-4 border-b p-5 sm:flex-row sm:items-center sm:justify-between">
            <div>
              <h2 className="text-xl font-bold">
                {selected ? t('edit') : t('create')}
              </h2>
              <p className="text-muted-foreground text-sm">
                {t('description')}
              </p>
            </div>
            <ProductPreview
              category={category}
              cosmeticId={cosmeticId}
              catalog={catalog}
              prices={selected?.prices ?? []}
              t={t}
            />
          </div>
          <form
            key={selected?.id ?? 'new'}
            ref={formRef}
            onSubmit={submit}
            onInput={(event) => checkRuFilled(event.currentTarget)}
          >
            <FormSection title={t('sections.main')}>
              <div className="grid gap-4 sm:grid-cols-2">
                <Field>
                  <FieldLabel htmlFor="product-category">
                    {t('fields.category')}
                  </FieldLabel>
                  <select
                    id="product-category"
                    value={category}
                    disabled={Boolean(selected)}
                    onChange={(event) => {
                      setCategory(
                        event.target.value as AdminProduct['category'],
                      )
                      setCosmeticId('')
                    }}
                    className="border-input bg-background h-9 rounded-md border px-3 text-sm"
                  >
                    <option value="upgrade">{t('categories.upgrade')}</option>
                    <option value="name-color">
                      {t('categories.name-color')}
                    </option>
                    <option value="name-prefix">
                      {t('categories.name-prefix')}
                    </option>
                  </select>
                  <FieldDescription>
                    {t.rich('tariffHint', {
                      tariff: priceByCategory[category],
                      code: (chunks) => (
                        <code className="font-mono">{chunks}</code>
                      ),
                    })}
                  </FieldDescription>
                </Field>
                {category !== 'upgrade' && (
                  <Field>
                    <FieldLabel htmlFor="product-cosmetic">
                      {t('fields.cosmetic')}
                    </FieldLabel>
                    <select
                      id="product-cosmetic"
                      value={cosmeticId}
                      disabled={Boolean(selected)}
                      required
                      onChange={(event) => setCosmeticId(event.target.value)}
                      className="border-input bg-background h-9 rounded-md border px-3 text-sm"
                    >
                      <option value="">{t('selectCosmetic')}</option>
                      {cosmetics.map((item) => (
                        <option key={item.id} value={item.id}>
                          {item.name}
                        </option>
                      ))}
                    </select>
                  </Field>
                )}
              </div>
            </FormSection>

            <FormSection title={t('sections.texts')}>
              <div className="grid gap-6 md:grid-cols-2">
                <LocalizationFields locale="ru" values={ru} t={t} />
                <LocalizationFields locale="en" values={en} t={t} />
              </div>
            </FormSection>

            <FormSection title={t('sections.sale')}>
              <Field>
                <FieldLabel htmlFor="product-ed">
                  {t('fields.easyDonate')}
                </FieldLabel>
                <div className="flex max-w-md gap-2">
                  <Input
                    id="product-ed"
                    name="easyDonateProductId"
                    type="number"
                    min={1}
                    value={easyDonateProductId}
                    onChange={(event) =>
                      setEasyDonateProductId(event.target.value)
                    }
                    required={active}
                  />
                  <EasyDonateCreateDialog
                    disabled={!ruFilled}
                    getProductInfo={getProductInfo}
                    onCreated={(id) => setEasyDonateProductId(String(id))}
                  />
                </div>
                <FieldDescription>{t('easyDonateHint')}</FieldDescription>
              </Field>
              <Field orientation="horizontal">
                <Switch
                  id="product-active"
                  checked={active}
                  onCheckedChange={setActive}
                />
                <div>
                  <FieldLabel htmlFor="product-active">
                    {t('fields.active')}
                  </FieldLabel>
                  <FieldDescription>{t('activeHint')}</FieldDescription>
                </div>
              </Field>
            </FormSection>

            <div className="bg-muted/40 flex justify-end rounded-b-lg px-5 py-3">
              <Button disabled={isPending}>{t('save')}</Button>
            </div>
          </form>
        </section>
      </div>
    </div>
  )
}

function FormSection({
  title,
  children,
}: {
  title: string
  children: React.ReactNode
}) {
  return (
    <div className="grid gap-4 border-b p-5">
      <h3 className="text-muted-foreground text-xs font-semibold tracking-wider uppercase">
        {title}
      </h3>
      {children}
    </div>
  )
}

function LocalizationFields({
  locale,
  values,
  t,
}: {
  locale: 'ru' | 'en'
  values?: AdminProduct['localizations'][number]
  t: ReturnType<typeof useTranslations>
}) {
  return (
    <div className="grid content-start gap-3">
      <Badge variant="outline" className="w-fit">
        {locale.toUpperCase()}
      </Badge>
      <Field>
        <FieldLabel htmlFor={`product-name-${locale}`}>
          {t('fields.name')}
        </FieldLabel>
        <Input
          id={`product-name-${locale}`}
          name={`name-${locale}`}
          required
          defaultValue={values?.name}
        />
      </Field>
      <Field>
        <FieldLabel htmlFor={`product-description-${locale}`}>
          {t('fields.description')}
        </FieldLabel>
        <textarea
          id={`product-description-${locale}`}
          name={`description-${locale}`}
          required
          defaultValue={values?.description}
          className="border-input bg-background min-h-24 resize-y rounded-md border px-3 py-2 text-sm"
        />
      </Field>
    </div>
  )
}

function ProductPreview({
  category,
  cosmeticId,
  catalog,
  prices,
  t,
}: {
  category: AdminProduct['category']
  cosmeticId: string
  catalog: AdminCosmeticsCatalog
  prices: AdminProduct['prices']
  t: ReturnType<typeof useTranslations>
}) {
  const color = catalog.nameColors.find((item) => item.id === cosmeticId)
  const prefix = catalog.namePrefixes.find((item) => item.id === cosmeticId)
  return (
    <aside className="bg-muted/40 flex min-w-56 shrink-0 flex-col items-center gap-1 rounded-lg border px-4 py-3">
      {category === 'name-color' && (
        <McUsername
          username="Keelfy"
          colors={color?.colors}
          className="text-xl"
        />
      )}
      {category === 'name-prefix' && (
        <div className="flex items-center gap-2">
          {prefix?.image && (
            <Image src={prefix.image} alt="" width={24} height={24} />
          )}
          <McUsername username="Keelfy" className="text-xl" />
        </div>
      )}
      {category === 'upgrade' && <strong>{t('seasonAccess')}</strong>}
      <div className="text-muted-foreground text-center text-xs">
        {prices.length
          ? prices
              .map((price) => `${price.amount} ${price.currency}`)
              .join(' · ')
          : t('pricesAfterSave')}
      </div>
    </aside>
  )
}
