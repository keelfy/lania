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
import { Checkbox } from '@/components/ui/checkbox'
import {
  Dialog,
  DialogClose,
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
import { Switch } from '@/components/ui/switch'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { createSeason, deleteSeason, updateSeason } from '@/lib/api-endpoints'
import { clientApiFetcher } from '@/lib/client'
import { errorToast } from '@/lib/toasts'
import { AdminSeason, SaveSeason } from '@/models/season'
import { PencilIcon, PlusIcon, Trash2Icon } from 'lucide-react'
import { useTranslations } from 'next-intl'
import { useRouter } from 'next/navigation'
import React from 'react'
import { toast } from 'sonner'

type Props = {
  seasons: AdminSeason[]
  locale: string
}

function dateInputValue(value?: number) {
  return value === undefined ? '' : new Date(value).toISOString().slice(0, 10)
}

function optionalString(form: FormData, name: string) {
  const value = String(form.get(name) ?? '').trim()
  return value || undefined
}

function SeasonDialog({ season }: { season?: AdminSeason }) {
  const t = useTranslations('admin.seasons')
  const router = useRouter()
  const [open, setOpen] = React.useState(false)
  const [isActive, setIsActive] = React.useState(season?.isActive ?? false)
  const [isPending, startTransition] = React.useTransition()

  const handleSubmit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    if (isPending) return

    const form = new FormData(event.currentTarget)
    const password = optionalString(form, 'rconPassword')
    const clearPassword = form.get('clearRconPassword') === 'on'
    const serverPort = optionalString(form, 'serverPort')
    const payload: SaveSeason = {
      seasonNumber: Number(form.get('seasonNumber')),
      name: String(form.get('name') ?? '').trim(),
      previewImage: optionalString(form, 'previewImage'),
      startDate: String(form.get('startDate') ?? ''),
      endDate: optionalString(form, 'endDate'),
      serverIp: optionalString(form, 'serverIp'),
      serverPort: serverPort ? Number(serverPort) : undefined,
      isActive,
      ...(password !== undefined
        ? { rconPassword: password }
        : clearPassword
          ? { rconPassword: '' }
          : {}),
    }

    startTransition(async () => {
      try {
        if (season) {
          await updateSeason(clientApiFetcher, season.id, payload)
          toast.success(t('updated'))
        } else {
          await createSeason(clientApiFetcher, payload)
          toast.success(t('created'))
        }
        setOpen(false)
        router.refresh()
      } catch (error) {
        errorToast(t('saveFailed'), error)
      }
    })
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button variant={season ? 'outline' : 'default'} size="sm">
          {season ? (
            <PencilIcon data-icon="inline-start" />
          ) : (
            <PlusIcon data-icon="inline-start" />
          )}
          {season ? t('edit') : t('create')}
        </Button>
      </DialogTrigger>
      <DialogContent className="max-h-[90vh] overflow-y-auto sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>
            {season ? t('editTitle') : t('createTitle')}
          </DialogTitle>
          <DialogDescription>{t('formDescription')}</DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="flex flex-col gap-6">
          <FieldGroup>
            <div className="grid gap-4 sm:grid-cols-2">
              <Field>
                <FieldLabel htmlFor={`season-number-${season?.id ?? 'new'}`}>
                  {t('fields.number')}
                </FieldLabel>
                <Input
                  id={`season-number-${season?.id ?? 'new'}`}
                  name="seasonNumber"
                  type="number"
                  min={1}
                  required
                  defaultValue={season?.seasonNumber}
                />
              </Field>
              <Field>
                <FieldLabel htmlFor={`season-name-${season?.id ?? 'new'}`}>
                  {t('fields.name')}
                </FieldLabel>
                <Input
                  id={`season-name-${season?.id ?? 'new'}`}
                  name="name"
                  maxLength={255}
                  required
                  defaultValue={season?.name}
                />
              </Field>
              <Field>
                <FieldLabel htmlFor={`season-start-${season?.id ?? 'new'}`}>
                  {t('fields.startDate')}
                </FieldLabel>
                <Input
                  id={`season-start-${season?.id ?? 'new'}`}
                  name="startDate"
                  type="date"
                  required
                  defaultValue={dateInputValue(season?.startDate)}
                />
              </Field>
              <Field>
                <FieldLabel htmlFor={`season-end-${season?.id ?? 'new'}`}>
                  {t('fields.endDate')}
                </FieldLabel>
                <Input
                  id={`season-end-${season?.id ?? 'new'}`}
                  name="endDate"
                  type="date"
                  defaultValue={dateInputValue(season?.endDate)}
                />
              </Field>
            </div>
            <Field>
              <FieldLabel htmlFor={`season-preview-${season?.id ?? 'new'}`}>
                {t('fields.previewImage')}
              </FieldLabel>
              <Input
                id={`season-preview-${season?.id ?? 'new'}`}
                name="previewImage"
                defaultValue={season?.previewImage}
                placeholder="s3://bucket/image.jpg"
              />
            </Field>
            <div className="grid gap-4 sm:grid-cols-2">
              <Field>
                <FieldLabel htmlFor={`season-ip-${season?.id ?? 'new'}`}>
                  {t('fields.serverIp')}
                </FieldLabel>
                <Input
                  id={`season-ip-${season?.id ?? 'new'}`}
                  name="serverIp"
                  defaultValue={season?.serverIp}
                  placeholder="203.0.113.10"
                />
              </Field>
              <Field>
                <FieldLabel htmlFor={`season-port-${season?.id ?? 'new'}`}>
                  {t('fields.serverPort')}
                </FieldLabel>
                <Input
                  id={`season-port-${season?.id ?? 'new'}`}
                  name="serverPort"
                  type="number"
                  min={1}
                  max={65535}
                  defaultValue={season?.serverPort}
                />
              </Field>
              <Field>
                <FieldLabel
                  htmlFor={`season-rcon-password-${season?.id ?? 'new'}`}
                >
                  {t('fields.rconPassword')}
                </FieldLabel>
                <Input
                  id={`season-rcon-password-${season?.id ?? 'new'}`}
                  name="rconPassword"
                  type="password"
                  autoComplete="new-password"
                  maxLength={4096}
                  placeholder={
                    season?.rconPasswordSet
                      ? t('passwordConfigured')
                      : undefined
                  }
                />
                {season?.rconPasswordSet && (
                  <FieldDescription>{t('passwordHint')}</FieldDescription>
                )}
              </Field>
            </div>
            {season?.rconPasswordSet && (
              <Field orientation="horizontal">
                <Checkbox
                  id={`season-clear-password-${season.id}`}
                  name="clearRconPassword"
                />
                <FieldLabel htmlFor={`season-clear-password-${season.id}`}>
                  {t('clearPassword')}
                </FieldLabel>
              </Field>
            )}
            <Field orientation="horizontal">
              <Switch
                id={`season-active-${season?.id ?? 'new'}`}
                checked={isActive}
                onCheckedChange={setIsActive}
                disabled={season?.isActive}
              />
              <div className="flex flex-col gap-1">
                <FieldLabel htmlFor={`season-active-${season?.id ?? 'new'}`}>
                  {t('fields.active')}
                </FieldLabel>
                <FieldDescription>
                  {season?.isActive ? t('activeHint') : t('inactiveHint')}
                </FieldDescription>
              </div>
            </Field>
          </FieldGroup>
          <DialogFooter>
            <DialogClose asChild>
              <Button type="button" variant="outline" disabled={isPending}>
                {t('cancel')}
              </Button>
            </DialogClose>
            <Button type="submit" disabled={isPending}>
              {t('save')}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}

function DeleteSeasonDialog({ season }: { season: AdminSeason }) {
  const t = useTranslations('admin.seasons')
  const router = useRouter()
  const [isPending, startTransition] = React.useTransition()

  const handleDelete = () => {
    startTransition(async () => {
      try {
        await deleteSeason(clientApiFetcher, season.id)
        toast.success(t('deleted'))
        router.refresh()
      } catch (error) {
        errorToast(t('deleteFailed'), error)
      }
    })
  }

  return (
    <AlertDialog>
      <AlertDialogTrigger asChild>
        <Button
          variant="destructive"
          size="sm"
          disabled={season.isActive || isPending}
          aria-label={t('delete')}
        >
          <Trash2Icon />
        </Button>
      </AlertDialogTrigger>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>{t('deleteTitle')}</AlertDialogTitle>
          <AlertDialogDescription>
            {t('deleteDescription', { name: season.name })}
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>{t('cancel')}</AlertDialogCancel>
          <AlertDialogAction onClick={handleDelete} disabled={isPending}>
            {t('delete')}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}

export default function SeasonManager({ seasons, locale }: Props) {
  const t = useTranslations('admin.seasons')
  const date = new Intl.DateTimeFormat(locale, { dateStyle: 'medium' })

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div>
          <h2 className="text-xl font-bold">{t('title')}</h2>
          <p className="text-muted-foreground text-sm">{t('description')}</p>
        </div>
        <SeasonDialog />
      </div>
      {seasons.length === 0 ? (
        <p className="text-muted-foreground py-10 text-center">{t('empty')}</p>
      ) : (
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{t('columns.season')}</TableHead>
              <TableHead>{t('columns.dates')}</TableHead>
              <TableHead>{t('columns.server')}</TableHead>
              <TableHead>{t('columns.rcon')}</TableHead>
              <TableHead className="text-right">
                {t('columns.actions')}
              </TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {seasons.map((season) => (
              <TableRow key={season.id}>
                <TableCell>
                  <div className="flex flex-col gap-1">
                    <span className="font-medium">{season.name}</span>
                    <span className="text-muted-foreground text-sm">
                      #{season.seasonNumber}
                    </span>
                    {season.isActive && <Badge>{t('active')}</Badge>}
                  </div>
                </TableCell>
                <TableCell>
                  {date.format(season.startDate)} –{' '}
                  {season.endDate ? date.format(season.endDate) : '—'}
                </TableCell>
                <TableCell>
                  {season.serverIp && season.serverPort
                    ? `${season.serverIp}:${season.serverPort}`
                    : '—'}
                </TableCell>
                <TableCell>
                  {season.rconPasswordSet
                    ? t('passwordOnly')
                    : '—'}
                </TableCell>
                <TableCell>
                  <div className="flex justify-end gap-2">
                    <SeasonDialog season={season} />
                    <DeleteSeasonDialog season={season} />
                  </div>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      )}
    </div>
  )
}
