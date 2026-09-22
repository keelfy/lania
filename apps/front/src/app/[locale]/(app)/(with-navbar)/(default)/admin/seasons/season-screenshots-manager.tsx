'use client'

import ProfilePicker from '@/components/profile-picker'
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
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import { Field, FieldGroup, FieldLabel } from '@/components/ui/field'
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
  createSeasonScreenshot,
  deleteSeasonScreenshot,
  getSeasonScreenshots,
  updateSeasonScreenshot,
} from '@/lib/api-endpoints'
import { clientApiFetcher } from '@/lib/client'
import { errorToast } from '@/lib/toasts'
import { AdminProfile } from '@/models/admin'
import {
  AdminSeason,
  SaveSeasonScreenshot,
  SeasonScreenshot,
} from '@/models/season'
import {
  ImagesIcon,
  PencilIcon,
  PlusIcon,
  Trash2Icon,
  XIcon,
} from 'lucide-react'
import { useTranslations } from 'next-intl'
import Image from 'next/image'
import React from 'react'
import { toast } from 'sonner'

type ScreenshotAuthorPick = { id: string; username: string }

function optionalString(form: FormData, name: string) {
  const value = String(form.get(name) ?? '').trim()
  return value || undefined
}

function AuthorChips({
  authors,
  onRemove,
}: {
  authors: ScreenshotAuthorPick[]
  onRemove: (id: string) => void
}) {
  if (authors.length === 0) return null
  return (
    <div className="flex flex-wrap gap-1.5">
      {authors.map((author) => (
        <Badge key={author.id} variant="secondary" className="gap-1 pr-1">
          {author.username}
          <button
            type="button"
            onClick={() => onRemove(author.id)}
            aria-label={author.username}
            className="hover:text-destructive"
          >
            <XIcon className="size-3" />
          </button>
        </Badge>
      ))}
    </div>
  )
}

function ScreenshotFormDialog({
  seasonId,
  screenshot,
  trigger,
  onSaved,
}: {
  seasonId: string
  screenshot?: SeasonScreenshot
  trigger: React.ReactNode
  onSaved: () => void
}) {
  const t = useTranslations('admin.seasons.screenshots')
  const [open, setOpen] = React.useState(false)
  const [authors, setAuthors] = React.useState<ScreenshotAuthorPick[]>(
    screenshot?.authors ?? [],
  )
  const [isPending, startTransition] = React.useTransition()

  const handleOpenChange = (next: boolean) => {
    setOpen(next)
    if (next) setAuthors(screenshot?.authors ?? [])
  }

  const handleSubmit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    if (isPending) return

    const form = new FormData(event.currentTarget)
    const payload: SaveSeasonScreenshot = {
      image: String(form.get('image') ?? '').trim(),
      title: optionalString(form, 'title'),
      position: Number(form.get('position') ?? 0),
      authorProfileIds: authors.map((author) => author.id),
    }

    startTransition(async () => {
      try {
        if (screenshot) {
          await updateSeasonScreenshot(
            clientApiFetcher,
            seasonId,
            screenshot.id,
            payload,
          )
          toast.success(t('updated'))
        } else {
          await createSeasonScreenshot(clientApiFetcher, seasonId, payload)
          toast.success(t('created'))
        }
        setOpen(false)
        onSaved()
      } catch (error) {
        errorToast(t('saveFailed'), error)
      }
    })
  }

  const idPrefix = `screenshot-${screenshot?.id ?? 'new'}`

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogTrigger asChild>{trigger}</DialogTrigger>
      <DialogContent className="max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>{screenshot ? t('edit') : t('add')}</DialogTitle>
          <DialogDescription>{t('dialogDescription')}</DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="flex flex-col gap-4">
          <FieldGroup>
            <Field>
              <FieldLabel htmlFor={`${idPrefix}-image`}>
                {t('fields.image')}
              </FieldLabel>
              <Input
                id={`${idPrefix}-image`}
                name="image"
                required
                maxLength={512}
                defaultValue={screenshot?.image}
                placeholder="s3://bucket/image.jpg"
              />
            </Field>
            <Field>
              <FieldLabel htmlFor={`${idPrefix}-title`}>
                {t('fields.title')}
              </FieldLabel>
              <Input
                id={`${idPrefix}-title`}
                name="title"
                maxLength={255}
                defaultValue={screenshot?.title}
              />
            </Field>
            <Field>
              <FieldLabel htmlFor={`${idPrefix}-position`}>
                {t('fields.position')}
              </FieldLabel>
              <Input
                id={`${idPrefix}-position`}
                name="position"
                type="number"
                defaultValue={screenshot?.position ?? 0}
              />
            </Field>
            <Field>
              <FieldLabel>{t('fields.authors')}</FieldLabel>
              {authors.length > 0 ? (
                <AuthorChips
                  authors={authors}
                  onRemove={(id) =>
                    setAuthors((prev) =>
                      prev.filter((author) => author.id !== id),
                    )
                  }
                />
              ) : (
                <p className="text-muted-foreground text-sm">
                  {t('noAuthors')}
                </p>
              )}
              <ProfilePicker
                excludeIds={authors.map((author) => author.id)}
                placeholder={t('authorsSearchPlaceholder')}
                noResultsLabel={t('noResults')}
                onSelect={(profile: AdminProfile) =>
                  setAuthors((prev) => [
                    ...prev,
                    { id: profile.id, username: profile.username },
                  ])
                }
              />
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

function DeleteScreenshotDialog({
  seasonId,
  screenshot,
  onDeleted,
}: {
  seasonId: string
  screenshot: SeasonScreenshot
  onDeleted: () => void
}) {
  const t = useTranslations('admin.seasons.screenshots')
  const [isPending, startTransition] = React.useTransition()

  const handleDelete = () => {
    startTransition(async () => {
      try {
        await deleteSeasonScreenshot(clientApiFetcher, seasonId, screenshot.id)
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
            {t('deleteDescription')}
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

export default function SeasonScreenshotsManager({
  season,
  compact = false,
}: {
  season: AdminSeason
  compact?: boolean
}) {
  const t = useTranslations('admin.seasons.screenshots')
  const [open, setOpen] = React.useState(false)
  const [screenshots, setScreenshots] = React.useState<SeasonScreenshot[]>([])
  const [loading, setLoading] = React.useState(false)

  const load = React.useCallback(() => {
    setLoading(true)
    getSeasonScreenshots(clientApiFetcher, season.id)
      .then((result) =>
        setScreenshots([...result].sort((a, b) => a.position - b.position)),
      )
      .catch((error) => errorToast(t('loadFailed'), error))
      .finally(() => setLoading(false))
  }, [season.id, t])

  const handleOpenChange = (next: boolean) => {
    setOpen(next)
    if (next) load()
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogTrigger asChild>
        <Button
          variant="outline"
          size={compact ? 'icon' : 'sm'}
          className={compact ? 'size-8' : undefined}
          aria-label={t('manage')}
        >
          <ImagesIcon data-icon={compact ? undefined : 'inline-start'} />
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
            <ScreenshotFormDialog
              seasonId={season.id}
              onSaved={load}
              trigger={
                <Button variant="default" size="sm">
                  <PlusIcon data-icon="inline-start" />
                  {t('add')}
                </Button>
              }
            />
          </div>

          {screenshots.length === 0 ? (
            <p className="text-muted-foreground py-6 text-center text-sm">
              {loading ? '…' : t('empty')}
            </p>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>{t('columns.image')}</TableHead>
                  <TableHead>{t('columns.title')}</TableHead>
                  <TableHead>{t('columns.authors')}</TableHead>
                  <TableHead className="text-right">
                    {t('columns.actions')}
                  </TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {screenshots.map((screenshot) => (
                  <TableRow key={screenshot.id}>
                    <TableCell>
                      <div className="relative size-12 overflow-hidden rounded-md">
                        <Image
                          src={screenshot.image}
                          alt={screenshot.title ?? ''}
                          fill
                          className="object-cover"
                        />
                      </div>
                    </TableCell>
                    <TableCell className="max-w-48 truncate">
                      {screenshot.title ?? '—'}
                    </TableCell>
                    <TableCell>
                      {screenshot.authors.length > 0
                        ? screenshot.authors
                            .map((author) => author.username)
                            .join(', ')
                        : '—'}
                    </TableCell>
                    <TableCell>
                      <div className="flex justify-end gap-2">
                        <ScreenshotFormDialog
                          seasonId={season.id}
                          screenshot={screenshot}
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
                        <DeleteScreenshotDialog
                          seasonId={season.id}
                          screenshot={screenshot}
                          onDeleted={load}
                        />
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
