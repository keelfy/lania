'use client'

import ImageUploadField from '@/components/admin/image-upload-field'
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
  Card,
  CardAction,
  CardContent,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
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
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import {
  createSeason,
  deleteSeason,
  updateSeason,
  uploadSeasonPreview,
} from '@/lib/api-endpoints'
import { clientApiFetcher } from '@/lib/client'
import { errorToast } from '@/lib/toasts'
import { AdminSeason, SaveSeason } from '@/models/season'
import {
  ExternalLinkIcon,
  PencilIcon,
  PlusIcon,
  Trash2Icon,
} from 'lucide-react'
import { useTranslations } from 'next-intl'
import { useRouter } from 'next/navigation'
import React from 'react'
import { toast } from 'sonner'
import AdminPageHeader from '../admin-page-header'
import SeasonScreenshotsManager from './season-screenshots-manager'

type Props = {
  seasons: AdminSeason[]
  locale: string
  title: string
}

function dateInputValue(value?: number) {
  return value === undefined ? '' : new Date(value).toISOString().slice(0, 10)
}

function optionalString(form: FormData, name: string) {
  const value = String(form.get(name) ?? '').trim()
  return value || undefined
}

function SeasonDialog({
  season,
  compact = false,
}: {
  season?: AdminSeason
  compact?: boolean
}) {
  const t = useTranslations('admin.seasons')
  const router = useRouter()
  const [open, setOpen] = React.useState(false)
  const [isActive, setIsActive] = React.useState(season?.isActive ?? false)
  const [isPrimary, setIsPrimary] = React.useState(season?.isPrimary ?? false)
  const [preregistration, setPreregistration] = React.useState(
    season?.preregistration ?? false,
  )
  const [freeRegistration, setFreeRegistration] = React.useState(
    season?.freeRegistration ?? false,
  )
  const [previewImage, setPreviewImage] = React.useState(
    season?.previewImage ?? '',
  )
  const [isPending, startTransition] = React.useTransition()

  const handleOpenChange = (nextOpen: boolean) => {
    setOpen(nextOpen)
    if (nextOpen) {
      setIsActive(season?.isActive ?? false)
      setIsPrimary(season?.isPrimary ?? false)
      setPreregistration(season?.preregistration ?? false)
      setFreeRegistration(season?.freeRegistration ?? false)
      setPreviewImage(season?.previewImage ?? '')
    }
  }

  const handleSubmit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    if (isPending) return

    const form = new FormData(event.currentTarget)
    const payload: SaveSeason = {
      name: String(form.get('name') ?? '').trim(),
      previewImage: previewImage || undefined,
      startDate: String(form.get('startDate') ?? ''),
      endDate: optionalString(form, 'endDate'),
      publicAddress: optionalString(form, 'publicAddress'),
      shellAddress: optionalString(form, 'shellAddress'),
      planUrl: optionalString(form, 'planUrl'),
      isActive,
      isPrimary,
      preregistration,
      freeRegistration,
      gameVersion: optionalString(form, 'gameVersion'),
      worldUrl: optionalString(form, 'worldUrl'),
      mapUrl: optionalString(form, 'mapUrl'),
      claimLimit: Number(form.get('claimLimit') ?? 0),
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
    <Dialog open={open} onOpenChange={handleOpenChange}>
      {compact ? (
        <Tooltip>
          <TooltipTrigger asChild>
            <DialogTrigger asChild>
              <Button
                variant="outline"
                size="icon"
                className="size-8"
                aria-label={t('edit')}
              >
                <PencilIcon />
              </Button>
            </DialogTrigger>
          </TooltipTrigger>
          <TooltipContent>{t('edit')}</TooltipContent>
        </Tooltip>
      ) : (
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
      )}
      <DialogContent className="max-h-[90vh] overflow-y-auto sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>
            {season ? t('editTitle') : t('createTitle')}
          </DialogTitle>
          <DialogDescription>{t('formDescription')}</DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="flex flex-col gap-6">
          <FieldGroup>
            <FieldGroup className="grid gap-4 sm:grid-cols-2">
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
            </FieldGroup>
            <Field>
              <FieldLabel htmlFor={`season-preview-${season?.id ?? 'new'}`}>
                {t('fields.previewImage')}
              </FieldLabel>
              <ImageUploadField
                id={`season-preview-${season?.id ?? 'new'}`}
                value={previewImage}
                onChange={setPreviewImage}
                upload={(file) => uploadSeasonPreview(clientApiFetcher, file)}
              />
            </Field>
            <FieldGroup className="grid gap-4 sm:grid-cols-2">
              <Field>
                <FieldLabel
                  htmlFor={`season-public-address-${season?.id ?? 'new'}`}
                >
                  {t('fields.publicAddress')}
                </FieldLabel>
                <Input
                  id={`season-public-address-${season?.id ?? 'new'}`}
                  name="publicAddress"
                  defaultValue={season?.publicAddress}
                  placeholder="play.example.com"
                />
              </Field>
              <Field>
                <FieldLabel
                  htmlFor={`season-shell-address-${season?.id ?? 'new'}`}
                >
                  {t('fields.shellAddress')}
                </FieldLabel>
                <Input
                  id={`season-shell-address-${season?.id ?? 'new'}`}
                  name="shellAddress"
                  defaultValue={season?.shellAddress}
                  placeholder="shell.internal:9090"
                />
                <FieldDescription>{t('shellAddressHint')}</FieldDescription>
              </Field>
            </FieldGroup>
            <Field>
              <FieldLabel htmlFor={`season-plan-url-${season?.id ?? 'new'}`}>
                {t('fields.planUrl')}
              </FieldLabel>
              <Input
                id={`season-plan-url-${season?.id ?? 'new'}`}
                name="planUrl"
                type="url"
                maxLength={2048}
                defaultValue={season?.planUrl}
                placeholder="https://plan.example.com"
              />
              <FieldDescription>{t('planUrlHint')}</FieldDescription>
            </Field>
            <FieldGroup className="grid gap-4 sm:grid-cols-2">
              <Field>
                <FieldLabel
                  htmlFor={`season-game-version-${season?.id ?? 'new'}`}
                >
                  {t('fields.gameVersion')}
                </FieldLabel>
                <Input
                  id={`season-game-version-${season?.id ?? 'new'}`}
                  name="gameVersion"
                  maxLength={64}
                  defaultValue={season?.gameVersion}
                  placeholder="1.21.1"
                />
              </Field>
              <Field>
                <FieldLabel htmlFor={`season-world-url-${season?.id ?? 'new'}`}>
                  {t('fields.worldUrl')}
                </FieldLabel>
                <Input
                  id={`season-world-url-${season?.id ?? 'new'}`}
                  name="worldUrl"
                  type="url"
                  maxLength={2048}
                  defaultValue={season?.worldUrl}
                  placeholder="https://cdn.example.com/world.zip"
                />
                <FieldDescription>{t('worldUrlHint')}</FieldDescription>
              </Field>
            </FieldGroup>
            <FieldGroup className="grid gap-4 sm:grid-cols-2">
              <Field>
                <FieldLabel htmlFor={`season-map-url-${season?.id ?? 'new'}`}>
                  {t('fields.mapUrl')}
                </FieldLabel>
                <Input
                  id={`season-map-url-${season?.id ?? 'new'}`}
                  name="mapUrl"
                  type="url"
                  maxLength={512}
                  defaultValue={season?.mapUrl}
                  placeholder="https://survival-map.lania.network"
                />
                <FieldDescription>{t('mapUrlHint')}</FieldDescription>
              </Field>
              <Field>
                <FieldLabel
                  htmlFor={`season-claim-limit-${season?.id ?? 'new'}`}
                >
                  {t('fields.claimLimit')}
                </FieldLabel>
                <Input
                  id={`season-claim-limit-${season?.id ?? 'new'}`}
                  name="claimLimit"
                  type="number"
                  min={0}
                  max={100000}
                  required
                  defaultValue={season?.claimLimit ?? 100}
                />
                <FieldDescription>{t('claimLimitHint')}</FieldDescription>
              </Field>
            </FieldGroup>
            <Field orientation="horizontal">
              <Switch
                id={`season-active-${season?.id ?? 'new'}`}
                checked={isActive}
                onCheckedChange={setIsActive}
              />
              <div className="flex flex-col gap-1">
                <FieldLabel htmlFor={`season-active-${season?.id ?? 'new'}`}>
                  {t('fields.active')}
                </FieldLabel>
                <FieldDescription>{t('activeHint')}</FieldDescription>
              </div>
            </Field>
            <Field orientation="horizontal">
              <Switch
                id={`season-primary-${season?.id ?? 'new'}`}
                checked={isPrimary}
                onCheckedChange={setIsPrimary}
                disabled={season?.isPrimary}
              />
              <div className="flex flex-col gap-1">
                <FieldLabel htmlFor={`season-primary-${season?.id ?? 'new'}`}>
                  {t('fields.primary')}
                </FieldLabel>
                <FieldDescription>{t('primaryHint')}</FieldDescription>
              </div>
            </Field>
            <Field orientation="horizontal">
              <Switch
                id={`season-preregistration-${season?.id ?? 'new'}`}
                checked={preregistration}
                onCheckedChange={setPreregistration}
              />
              <div className="flex flex-col gap-1">
                <FieldLabel
                  htmlFor={`season-preregistration-${season?.id ?? 'new'}`}
                >
                  {t('fields.preregistration')}
                </FieldLabel>
                <FieldDescription>{t('preregistrationHint')}</FieldDescription>
              </div>
            </Field>
            <Field orientation="horizontal">
              <Switch
                id={`season-free-registration-${season?.id ?? 'new'}`}
                checked={freeRegistration}
                onCheckedChange={setFreeRegistration}
              />
              <div className="flex flex-col gap-1">
                <FieldLabel
                  htmlFor={`season-free-registration-${season?.id ?? 'new'}`}
                >
                  {t('fields.freeRegistration')}
                </FieldLabel>
                <FieldDescription>{t('freeRegistrationHint')}</FieldDescription>
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

  if (season.isActive || season.isPrimary) return null

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
      <Tooltip>
        <TooltipTrigger asChild>
          <AlertDialogTrigger asChild>
            <Button
              variant="destructive"
              size="icon"
              className="size-8"
              disabled={isPending}
              aria-label={t('delete')}
            >
              <Trash2Icon />
            </Button>
          </AlertDialogTrigger>
        </TooltipTrigger>
        <TooltipContent>{t('delete')}</TooltipContent>
      </Tooltip>
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

function SeasonBadges({ season }: { season: AdminSeason }) {
  const t = useTranslations('admin.seasons')

  return (
    <div className="flex flex-wrap items-center gap-1.5">
      {season.isActive && <Badge>{t('active')}</Badge>}
      {season.isPrimary && <Badge variant="secondary">{t('primary')}</Badge>}
      {season.preregistration && (
        <Badge variant="outline">{t('preregistration')}</Badge>
      )}
      {season.freeRegistration && (
        <Badge variant="outline">{t('freeRegistration')}</Badge>
      )}
    </div>
  )
}

function PlanLink({ url, label }: { url: string; label: string }) {
  return (
    <Button asChild variant="outline" size="sm">
      <a href={url} target="_blank" rel="noopener noreferrer">
        {label}
        <ExternalLinkIcon data-icon="inline-end" />
      </a>
    </Button>
  )
}

function SeasonInfo({
  label,
  children,
}: {
  label: string
  children: React.ReactNode
}) {
  return (
    <div className="flex flex-col gap-1">
      <dt className="text-muted-foreground text-xs">{label}</dt>
      <dd className="text-sm">{children}</dd>
    </div>
  )
}

export default function SeasonManager({ seasons, locale, title }: Props) {
  const t = useTranslations('admin.seasons')
  const dateFormat = new Intl.DateTimeFormat(locale, { dateStyle: 'medium' })
  const dates = (season: AdminSeason) =>
    season.endDate
      ? dateFormat.formatRange(season.startDate, season.endDate)
      : t('since', { date: dateFormat.format(season.startDate) })

  const active = seasons.filter((season) => season.isActive)
  const archive = seasons.filter((season) => !season.isActive)

  return (
    <div className="flex flex-col gap-6">
      <AdminPageHeader
        title={title}
        description={t('description')}
        actions={<SeasonDialog />}
      />
      {seasons.length === 0 && (
        <p className="text-muted-foreground py-10 text-center">{t('empty')}</p>
      )}
      {active.length > 0 && (
        <section className="flex flex-col gap-3">
          <h3 className="text-muted-foreground text-sm font-semibold">
            {t('sections.active')}
          </h3>
          {active.map((season) => (
            <Card key={season.id} className="gap-4 py-5">
              <CardHeader>
                <CardTitle className="text-lg">{season.name}</CardTitle>
                <SeasonBadges season={season} />
                <CardAction className="flex gap-2">
                  <SeasonScreenshotsManager season={season} />
                  <SeasonDialog season={season} />
                </CardAction>
              </CardHeader>
              <CardContent>
                <dl className="grid gap-x-6 gap-y-4 sm:grid-cols-2 lg:grid-cols-4">
                  <SeasonInfo label={t('columns.dates')}>
                    {dates(season)}
                  </SeasonInfo>
                  <SeasonInfo label={t('columns.server')}>
                    {season.publicAddress ?? '—'}
                  </SeasonInfo>
                  <SeasonInfo label={t('columns.shell')}>
                    {season.shellAddress ?? '—'}
                  </SeasonInfo>
                  <SeasonInfo label={t('columns.plan')}>
                    {season.planUrl ? (
                      <PlanLink url={season.planUrl} label={t('openPlan')} />
                    ) : (
                      '—'
                    )}
                  </SeasonInfo>
                </dl>
              </CardContent>
            </Card>
          ))}
        </section>
      )}
      {archive.length > 0 && (
        <section className="flex flex-col gap-3">
          <h3 className="text-muted-foreground text-sm font-semibold">
            {t('sections.archive')}
          </h3>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{t('columns.season')}</TableHead>
                <TableHead className="hidden sm:table-cell">
                  {t('columns.dates')}
                </TableHead>
                <TableHead className="text-right">
                  {t('columns.actions')}
                </TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {archive.map((season) => {
                const addresses = [season.publicAddress, season.shellAddress]
                  .filter(Boolean)
                  .join(' · ')

                return (
                  <TableRow key={season.id}>
                    <TableCell>
                      <div className="flex flex-col gap-1">
                        <div className="flex flex-wrap items-center gap-2">
                          <span className="font-medium">{season.name}</span>
                          <SeasonBadges season={season} />
                        </div>
                        <span className="text-muted-foreground text-xs sm:hidden">
                          {dates(season)}
                        </span>
                        {addresses && (
                          <span className="text-muted-foreground text-xs">
                            {addresses}
                          </span>
                        )}
                      </div>
                    </TableCell>
                    <TableCell className="hidden sm:table-cell">
                      {dates(season)}
                    </TableCell>
                    <TableCell>
                      <div className="flex justify-end gap-2">
                        {season.planUrl && (
                          <Tooltip>
                            <TooltipTrigger asChild>
                              <Button
                                asChild
                                variant="outline"
                                size="icon"
                                className="size-8"
                                aria-label={t('openPlan')}
                              >
                                <a
                                  href={season.planUrl}
                                  target="_blank"
                                  rel="noopener noreferrer"
                                >
                                  <ExternalLinkIcon />
                                </a>
                              </Button>
                            </TooltipTrigger>
                            <TooltipContent>{t('openPlan')}</TooltipContent>
                          </Tooltip>
                        )}
                        <SeasonScreenshotsManager season={season} compact />
                        <SeasonDialog season={season} compact />
                        <DeleteSeasonDialog season={season} />
                      </div>
                    </TableCell>
                  </TableRow>
                )
              })}
            </TableBody>
          </Table>
        </section>
      )}
    </div>
  )
}
