'use client'

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Field,
  FieldDescription,
  FieldGroup,
  FieldLabel,
} from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import {
  createPrivilege,
  deletePrivilege,
  updatePrivilege,
} from '@/lib/api-endpoints'
import { clientApiFetcher } from '@/lib/client'
import { CURRENCY_SYMBOLS, SUPPORTED_CURRENCIES } from '@/lib/currency'
import { errorToast } from '@/lib/toasts'
import { AdminPrivilege, AdminProduct, CosmeticNames } from '@/models/admin'
import { KeyRoundIcon, PlusIcon, ShoppingBagIcon } from 'lucide-react'
import { useTranslations } from 'next-intl'
import Link from 'next/link'
import { useRouter } from 'next/navigation'
import React from 'react'
import { toast } from 'sonner'
import AdminPageHeader from '../admin-page-header'

type Translate = ReturnType<typeof useTranslations>

function linkedProducts(products: AdminProduct[], id?: string): AdminProduct[] {
  if (!id) return []
  return products.filter(
    (product) =>
      product.category === 'privilege' && product.metadata.privilegeId === id,
  )
}

// A blank translation is left out, so the main name stands in for it.
function formNames(form: FormData): CosmeticNames {
  const ru = String(form.get('names-ru') ?? '').trim()
  return ru ? { ru } : {}
}

export default function PrivilegesManager({
  title,
  privileges,
  products,
}: {
  title: string
  privileges: AdminPrivilege[]
  products: AdminProduct[]
}) {
  const t = useTranslations('admin.privileges')
  const router = useRouter()
  const [selected, setSelected] = React.useState<AdminPrivilege>()
  const [search, setSearch] = React.useState('')
  const [name, setName] = React.useState('')
  const [isPending, startTransition] = React.useTransition()

  const query = search.trim().toLocaleLowerCase()
  const items = privileges.filter((item) =>
    [item.name, item.names.ru, item.permission].some((value) =>
      value?.toLocaleLowerCase().includes(query),
    ),
  )
  const key = selected?.id ?? 'new'
  const selectedProducts = linkedProducts(products, selected?.id)

  function select(item?: AdminPrivilege) {
    setSelected(item)
    setName(item?.name ?? '')
  }

  function submit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (isPending) return
    const form = new FormData(event.currentTarget)
    const payload = {
      name: String(form.get('name') ?? '').trim(),
      names: formNames(form),
      permission: String(form.get('permission') ?? '').trim(),
      prices: SUPPORTED_CURRENCIES.map((currency) => ({
        currency,
        amount: Number(form.get(`price-${currency}`)),
      })),
    }
    startTransition(async () => {
      try {
        const saved = selected
          ? await updatePrivilege(clientApiFetcher, selected.id, payload)
          : await createPrivilege(clientApiFetcher, payload)
        toast.success(t(selected ? 'updated' : 'created'))
        select(saved)
        router.refresh()
      } catch (error) {
        errorToast(t('saveFailed'), error)
      }
    })
  }

  function remove() {
    if (!selected || isPending) return
    const id = selected.id
    startTransition(async () => {
      try {
        await deletePrivilege(clientApiFetcher, id)
        toast.success(t('deleted'))
        select()
        router.refresh()
      } catch (error) {
        errorToast(t('deleteFailed'), error)
      }
    })
  }

  return (
    <div className="flex flex-col gap-6">
      <AdminPageHeader
        title={title}
        actions={
          <Button onClick={() => select()}>
            <PlusIcon />
            {t('create')}
          </Button>
        }
      />

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
            {items.length === 0 ? (
              <p className="text-muted-foreground py-8 text-center text-sm">
                {t('empty')}
              </p>
            ) : (
              <div className="grid gap-1">
                {items.map((item) => (
                  <button
                    key={item.id}
                    type="button"
                    onClick={() => select(item)}
                    className={`flex min-w-0 items-center gap-3 rounded-md border px-3 py-2 text-left text-sm transition-colors ${selected?.id === item.id ? 'border-primary bg-accent' : 'hover:bg-muted border-transparent'}`}
                  >
                    <KeyRoundIcon className="text-muted-foreground size-5 shrink-0" />
                    <span className="flex min-w-0 flex-1 flex-col">
                      <span className="truncate font-medium">{item.name}</span>
                      <span className="text-muted-foreground truncate font-mono text-xs">
                        {item.permission}
                      </span>
                    </span>
                    {linkedProducts(products, item.id).length > 0 && (
                      <ShoppingBagIcon className="text-muted-foreground size-3.5 shrink-0" />
                    )}
                  </button>
                ))}
              </div>
            )}
          </div>
        </section>

        <section className="border-border rounded-lg border p-5">
          <div className="mb-6">
            <h2 className="text-xl font-bold">
              {selected ? t('edit') : t('createTitle')}
            </h2>
            <p className="text-muted-foreground text-sm">{t('hint')}</p>
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
                  defaultValue={selected?.name}
                  onChange={(event) => setName(event.target.value)}
                />
                <FieldDescription>{t('nameHint')}</FieldDescription>
              </Field>
              <Field>
                <FieldLabel htmlFor={`${key}-names-ru`}>
                  {t('fields.nameRu')}
                </FieldLabel>
                <Input
                  id={`${key}-names-ru`}
                  name="names-ru"
                  maxLength={255}
                  defaultValue={selected?.names.ru}
                  placeholder={name}
                />
                <FieldDescription>{t('nameRuHint')}</FieldDescription>
              </Field>
              <Field>
                <FieldLabel htmlFor={`${key}-permission`}>
                  {t('fields.permission')}
                </FieldLabel>
                <Input
                  id={`${key}-permission`}
                  name="permission"
                  required
                  maxLength={200}
                  defaultValue={selected?.permission}
                  placeholder="homes.commands.*"
                  className="font-mono"
                  autoComplete="off"
                  spellCheck={false}
                />
                <FieldDescription>{t('permissionHint')}</FieldDescription>
              </Field>
              <PriceFields id={key} item={selected} t={t} />
              <div className="flex flex-wrap items-center gap-2">
                <Button disabled={isPending}>{t('save')}</Button>
                {selected && (
                  <DeleteButton
                    disabled={isPending || selectedProducts.length > 0}
                    blockedByProduct={selectedProducts.length > 0}
                    onConfirm={remove}
                    t={t}
                  />
                )}
              </div>
            </FieldGroup>
            {selected && <LinkedProducts products={selectedProducts} t={t} />}
          </form>
        </section>
      </div>
    </div>
  )
}

function PriceFields({
  id,
  item,
  t,
}: {
  id: string
  item?: AdminPrivilege
  t: Translate
}) {
  return (
    <fieldset className="flex flex-col gap-3">
      <legend className="mb-1 text-sm font-medium">{t('fields.prices')}</legend>
      <div className="grid grid-cols-2 gap-3 sm:grid-cols-3">
        {SUPPORTED_CURRENCIES.map((currency) => {
          const price = item?.prices.find((p) => p.currency === currency)
          return (
            <Field key={currency}>
              <FieldLabel htmlFor={`${id}-price-${currency}`}>
                {currency} ({CURRENCY_SYMBOLS[currency]})
              </FieldLabel>
              <Input
                id={`${id}-price-${currency}`}
                name={`price-${currency}`}
                type="number"
                inputMode="decimal"
                min="0.01"
                step="0.01"
                required
                defaultValue={price?.amount}
              />
            </Field>
          )
        })}
      </div>
      <p className="text-muted-foreground text-sm">{t('pricesHint')}</p>
    </fieldset>
  )
}

function DeleteButton({
  disabled,
  blockedByProduct,
  onConfirm,
  t,
}: {
  disabled: boolean
  blockedByProduct: boolean
  onConfirm: () => void
  t: Translate
}) {
  return (
    <>
      <AlertDialog>
        <AlertDialogTrigger asChild>
          <Button type="button" variant="destructive" disabled={disabled}>
            {t('delete')}
          </Button>
        </AlertDialogTrigger>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{t('deleteConfirmTitle')}</AlertDialogTitle>
            <AlertDialogDescription>
              {t('deleteConfirmDescription')}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>{t('cancel')}</AlertDialogCancel>
            <AlertDialogAction onClick={onConfirm}>
              {t('delete')}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
      {blockedByProduct && (
        <span className="text-muted-foreground text-sm">
          {t('deleteBlockedByProduct')}
        </span>
      )}
    </>
  )
}

function LinkedProducts({
  products,
  t,
}: {
  products: AdminProduct[]
  t: Translate
}) {
  return (
    <div className="border-border h-fit rounded-lg border p-4">
      <h3 className="text-muted-foreground mb-3 flex items-center gap-2 text-sm font-semibold">
        <ShoppingBagIcon className="size-4" />
        {t('linkedProduct')}
      </h3>
      {products.length === 0 ? (
        <p className="text-muted-foreground text-sm">
          {t('linkedProductNone')}
        </p>
      ) : (
        <ul className="flex flex-col gap-2">
          {products.map((product) => {
            const name =
              product.localizations.find((item) => item.locale === 'ru')
                ?.name ?? product.id
            return (
              <li key={product.id}>
                <Link
                  href="/admin/products"
                  className="hover:bg-muted flex items-center justify-between gap-3 rounded-md border px-3 py-2 text-sm"
                >
                  <span className="font-medium">{name}</span>
                  <Badge variant={product.isActive ? 'default' : 'secondary'}>
                    {t(product.isActive ? 'productActive' : 'productDraft')}
                  </Badge>
                </Link>
              </li>
            )
          })}
        </ul>
      )}
    </div>
  )
}
