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
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  createSeasonWorld,
  deleteSeasonWorld,
  getSeasonWorlds,
  updateSeasonWorld,
  uploadSeasonScreenshotImage,
} from '@/lib/api-endpoints'
import { clientApiFetcher } from '@/lib/client'
import { errorToast } from '@/lib/toasts'
import { AdminSeason, SaveSeasonWorld, SeasonWorld } from '@/models/season'
import { GlobeIcon, PencilIcon, PlusIcon, Trash2Icon } from 'lucide-react'
import { useTranslations } from 'next-intl'
import React from 'react'
import { toast } from 'sonner'

// squaremap names the vanilla dimensions this way; anything else is typed in by hand.
const VANILLA_DIMENSIONS = [
  'minecraft_overworld',
  'minecraft_the_nether',
  'minecraft_the_end',
] as const

function optionalString(form: FormData, name: string) {
  const value = String(form.get(name) ?? '').trim()
  return value || undefined
}

function WorldFormDialog({
  seasonId,
  world,
  nextPosition,
  trigger,
  onSaved,
}: {
  seasonId: string
  world?: SeasonWorld
  nextPosition: number
  trigger: React.ReactNode
  onSaved: () => void
}) {
  const t = useTranslations('admin.seasons.worlds')
  const [open, setOpen] = React.useState(false)
  const [image, setImage] = React.useState(world?.previewImage ?? '')
  // New worlds claim in the overworld, as most do.
  const initialDimensions = () =>
    world?.claimDimensions ?? ['minecraft_overworld']
  const [dimensions, setDimensions] = React.useState(initialDimensions)
  const [isPending, startTransition] = React.useTransition()

  const handleOpenChange = (next: boolean) => {
    setOpen(next)
    if (next) {
      setImage(world?.previewImage ?? '')
      setDimensions(initialDimensions())
    }
  }

  const toggleDimension = (name: string, checked: boolean) =>
    setDimensions((current) =>
      checked ? [...current, name] : current.filter((item) => item !== name),
    )
  const otherDimensions = dimensions.filter(
    (name) => !(VANILLA_DIMENSIONS as readonly string[]).includes(name),
  )

  const handleSubmit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    if (isPending) return

    const form = new FormData(event.currentTarget)
    const others = String(form.get('otherDimensions') ?? '')
      .split(',')
      .map((name) => name.trim())
      .filter(Boolean)
    const payload: SaveSeasonWorld = {
      slug: String(form.get('slug') ?? '').trim(),
      name: String(form.get('name') ?? '').trim(),
      previewImage: image || undefined,
      mapUrl: optionalString(form, 'mapUrl'),
      claimLimit: Number(form.get('claimLimit') ?? 0),
      claimDimensions: [
        ...new Set([
          ...VANILLA_DIMENSIONS.filter((name) => dimensions.includes(name)),
          ...others,
        ]),
      ],
      planServer: optionalString(form, 'planServer'),
      claimMinPlaytimeHours: Number(form.get('claimMinPlaytimeHours') ?? 0),
      position: Number(form.get('position') ?? 0),
    }

    startTransition(async () => {
      try {
        if (world) {
          await updateSeasonWorld(clientApiFetcher, world.id, payload)
          toast.success(t('updated'))
        } else {
          await createSeasonWorld(clientApiFetcher, seasonId, payload)
          toast.success(t('created'))
        }
        setOpen(false)
        onSaved()
      } catch (error) {
        errorToast(t('saveFailed'), error)
      }
    })
  }

  const idPrefix = `world-${world?.id ?? 'new'}`

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogTrigger asChild>{trigger}</DialogTrigger>
      <DialogContent className="max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>{world ? t('edit') : t('add')}</DialogTitle>
          <DialogDescription>{t('formDescription')}</DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="flex flex-col gap-4">
          <FieldGroup>
            <FieldGroup className="grid gap-4 sm:grid-cols-2">
              <Field>
                <FieldLabel htmlFor={`${idPrefix}-name`}>
                  {t('fields.name')}
                </FieldLabel>
                <Input
                  id={`${idPrefix}-name`}
                  name="name"
                  maxLength={64}
                  required
                  defaultValue={world?.name}
                />
              </Field>
              <Field>
                <FieldLabel htmlFor={`${idPrefix}-slug`}>
                  {t('fields.slug')}
                </FieldLabel>
                <Input
                  id={`${idPrefix}-slug`}
                  name="slug"
                  maxLength={32}
                  pattern="[a-z0-9]+(-[a-z0-9]+)*"
                  required
                  defaultValue={world?.slug}
                  placeholder="survival"
                />
                <FieldDescription>{t('slugHint')}</FieldDescription>
              </Field>
            </FieldGroup>
            <Field>
              <FieldLabel htmlFor={`${idPrefix}-image`}>
                {t('fields.previewImage')}
              </FieldLabel>
              <ImageUploadField
                id={`${idPrefix}-image`}
                value={image}
                onChange={setImage}
                upload={(file) =>
                  uploadSeasonScreenshotImage(clientApiFetcher, seasonId, file)
                }
                previewSize={80}
              />
            </Field>
            <Field>
              <FieldLabel htmlFor={`${idPrefix}-map-url`}>
                {t('fields.mapUrl')}
              </FieldLabel>
              <Input
                id={`${idPrefix}-map-url`}
                name="mapUrl"
                type="url"
                maxLength={512}
                defaultValue={world?.mapUrl}
                placeholder="https://survival-map.lania.network"
              />
              <FieldDescription>{t('mapUrlHint')}</FieldDescription>
            </Field>
            <Field>
              <FieldLabel>{t('fields.claimDimensions')}</FieldLabel>
              <div className="flex flex-wrap gap-x-5 gap-y-2">
                {VANILLA_DIMENSIONS.map((name) => (
                  <label
                    key={name}
                    className="flex items-center gap-2 text-sm font-normal"
                  >
                    <Checkbox
                      checked={dimensions.includes(name)}
                      onCheckedChange={(checked) =>
                        toggleDimension(name, checked === true)
                      }
                    />
                    {t(`dimensions.${name}`)}
                  </label>
                ))}
              </div>
              <Input
                name="otherDimensions"
                defaultValue={otherDimensions.join(', ')}
                placeholder={t('otherDimensionsPlaceholder')}
                aria-label={t('otherDimensionsPlaceholder')}
              />
              <FieldDescription>{t('claimDimensionsHint')}</FieldDescription>
            </Field>
            <FieldGroup className="grid gap-4 sm:grid-cols-2">
              <Field>
                <FieldLabel htmlFor={`${idPrefix}-claim-limit`}>
                  {t('fields.claimLimit')}
                </FieldLabel>
                <Input
                  id={`${idPrefix}-claim-limit`}
                  name="claimLimit"
                  type="number"
                  min={0}
                  max={100000}
                  required
                  defaultValue={world?.claimLimit ?? 100}
                />
                <FieldDescription>{t('claimLimitHint')}</FieldDescription>
              </Field>
              <Field>
                <FieldLabel htmlFor={`${idPrefix}-claim-min-playtime`}>
                  {t('fields.claimMinPlaytimeHours')}
                </FieldLabel>
                <Input
                  id={`${idPrefix}-claim-min-playtime`}
                  name="claimMinPlaytimeHours"
                  type="number"
                  min={0}
                  max={10000}
                  required
                  defaultValue={world?.claimMinPlaytimeHours ?? 5}
                />
                <FieldDescription>
                  {t('claimMinPlaytimeHoursHint')}
                </FieldDescription>
              </Field>
              <Field>
                <FieldLabel htmlFor={`${idPrefix}-plan-server`}>
                  {t('fields.planServer')}
                </FieldLabel>
                <Input
                  id={`${idPrefix}-plan-server`}
                  name="planServer"
                  maxLength={100}
                  defaultValue={world?.planServer}
                  placeholder="survival"
                />
                <FieldDescription>{t('planServerHint')}</FieldDescription>
              </Field>
              <Field>
                <FieldLabel htmlFor={`${idPrefix}-position`}>
                  {t('fields.position')}
                </FieldLabel>
                <Input
                  id={`${idPrefix}-position`}
                  name="position"
                  type="number"
                  defaultValue={world?.position ?? nextPosition}
                />
              </Field>
            </FieldGroup>
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

function DeleteWorldDialog({
  world,
  onDeleted,
}: {
  world: SeasonWorld
  onDeleted: () => void
}) {
  const t = useTranslations('admin.seasons.worlds')
  const [isPending, startTransition] = React.useTransition()

  const handleDelete = () => {
    startTransition(async () => {
      try {
        await deleteSeasonWorld(clientApiFetcher, world.id)
        toast.success(t('deleted'))
        onDeleted()
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
          size="icon"
          className="size-8"
          disabled={isPending}
          aria-label={t('delete')}
        >
          <Trash2Icon />
        </Button>
      </AlertDialogTrigger>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>{t('deleteTitle')}</AlertDialogTitle>
          <AlertDialogDescription>
            {t('deleteDescription', { name: world.name })}
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

export default function SeasonWorldsManager({
  season,
  compact = false,
}: {
  season: AdminSeason
  compact?: boolean
}) {
  const t = useTranslations('admin.seasons.worlds')
  const [open, setOpen] = React.useState(false)
  const [worlds, setWorlds] = React.useState<SeasonWorld[]>([])
  const [loading, setLoading] = React.useState(false)

  const load = React.useCallback(() => {
    setLoading(true)
    getSeasonWorlds(clientApiFetcher, season.id)
      .then(setWorlds)
      .catch((error) => errorToast(t('loadFailed'), error))
      .finally(() => setLoading(false))
  }, [season.id, t])

  const handleOpenChange = (next: boolean) => {
    setOpen(next)
    if (next) load()
  }

  const nextPosition =
    worlds.length === 0
      ? 0
      : Math.max(...worlds.map((world) => world.position)) + 1

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogTrigger asChild>
        <Button
          variant="outline"
          size={compact ? 'icon' : 'sm'}
          className={compact ? 'size-8' : undefined}
          aria-label={t('manage')}
        >
          <GlobeIcon data-icon={compact ? undefined : 'inline-start'} />
          {!compact && t('manage')}
        </Button>
      </DialogTrigger>
      <DialogContent className="max-h-[90vh] overflow-y-auto sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>{t('dialogTitle', { name: season.name })}</DialogTitle>
          <DialogDescription>{t('dialogDescription')}</DialogDescription>
        </DialogHeader>

        <div className="flex flex-col gap-4">
          <div className="flex justify-end">
            <WorldFormDialog
              seasonId={season.id}
              nextPosition={nextPosition}
              onSaved={load}
              trigger={
                <Button size="sm">
                  <PlusIcon data-icon="inline-start" />
                  {t('add')}
                </Button>
              }
            />
          </div>

          {worlds.length === 0 ? (
            <p className="text-muted-foreground py-6 text-center text-sm">
              {loading ? '…' : t('empty')}
            </p>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>{t('columns.world')}</TableHead>
                  <TableHead>{t('columns.claims')}</TableHead>
                  <TableHead className="text-right">
                    {t('columns.actions')}
                  </TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {worlds.map((world) => (
                  <TableRow key={world.id}>
                    <TableCell>
                      <div className="flex flex-col gap-0.5">
                        <span className="font-medium">{world.name}</span>
                        <span className="text-muted-foreground font-mono text-xs">
                          /worlds/{world.slug}
                        </span>
                        <span className="text-muted-foreground max-w-64 truncate text-xs">
                          {world.mapUrl ?? t('noMap')}
                        </span>
                      </div>
                    </TableCell>
                    <TableCell className="text-sm">
                      {world.mapUrl && world.claimDimensions.length > 0
                        ? t('claimsSummary', {
                            limit: world.claimLimit,
                            dimensions: world.claimDimensions.length,
                          })
                        : t('claimsOff')}
                    </TableCell>
                    <TableCell>
                      <div className="flex justify-end gap-2">
                        <WorldFormDialog
                          seasonId={season.id}
                          world={world}
                          nextPosition={nextPosition}
                          onSaved={load}
                          trigger={
                            <Button
                              variant="outline"
                              size="icon"
                              className="size-8"
                              aria-label={t('edit')}
                            >
                              <PencilIcon />
                            </Button>
                          }
                        />
                        <DeleteWorldDialog world={world} onDeleted={load} />
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </div>
      </DialogContent>
    </Dialog>
  )
}
