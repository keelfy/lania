'use client'

import NotificationItem from '@/components/notification-item'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Field, FieldDescription, FieldLabel } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { sendAnnouncement } from '@/lib/api-endpoints'
import { clientApiFetcher } from '@/lib/client'
import { Locale, LOCALE_NAMES, LOCALES } from '@/lib/locale'
import { errorToast } from '@/lib/toasts'
import { AnnouncementNotification, LocalizedText } from '@/models/notification'
import { SendIcon } from 'lucide-react'
import { useTranslations } from 'next-intl'
import React from 'react'
import { toast } from 'sonner'

// The limits the API checks, kept here so the form stops the admin first.
const MAX_TITLE_LENGTH = 120
const MAX_BODY_LENGTH = 1000
const MAX_LINK_LENGTH = 255

const emptyText = (): LocalizedText =>
  Object.fromEntries(LOCALES.map((locale) => [locale, ''])) as LocalizedText

const trimText = (text: LocalizedText): LocalizedText =>
  Object.fromEntries(
    LOCALES.map((locale) => [locale, text[locale].trim()]),
  ) as LocalizedText

const noop = () => {}

// AnnouncementForm writes news in every language of the site and sends it to every user's bell.
export default function AnnouncementForm() {
  const t = useTranslations('admin.announcements')
  const [title, setTitle] = React.useState(emptyText)
  const [body, setBody] = React.useState(emptyText)
  const [link, setLink] = React.useState('')
  const [confirming, setConfirming] = React.useState(false)
  const [isPending, startTransition] = React.useTransition()

  const trimmedBody = trimText(body)
  const hasBody = LOCALES.some((locale) => trimmedBody[locale])
  const trimmedLink = link.trim()
  const announcement = {
    title: trimText(title),
    body: hasBody ? trimmedBody : undefined,
    link: trimmedLink || undefined,
  }

  const preview: AnnouncementNotification = {
    id: 'preview',
    type: 'announcement',
    createdAt: new Date().toISOString(),
    payload: {
      ...announcement,
      title: {
        ...announcement.title,
        // An empty preview still shows where the headline goes.
        ...Object.fromEntries(
          LOCALES.filter((locale) => !announcement.title[locale]).map(
            (locale) => [locale, t('previewPlaceholder')],
          ),
        ),
      },
    },
  }

  function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setConfirming(true)
  }

  function send() {
    startTransition(async () => {
      try {
        const { recipients } = await sendAnnouncement(
          clientApiFetcher,
          announcement,
        )
        toast.success(t('sent', { recipients }))
        setTitle(emptyText())
        setBody(emptyText())
        setLink('')
      } catch (error) {
        errorToast(t('sendFailed'), error)
      }
    })
  }

  return (
    <div className="grid items-start gap-6 lg:grid-cols-[minmax(0,1fr)_24rem]">
      <form
        onSubmit={handleSubmit}
        className="bg-card flex flex-col gap-6 rounded-xl border p-5"
      >
        <div className="grid gap-6 md:grid-cols-2">
          {LOCALES.map((locale) => (
            <LocaleFields
              key={locale}
              locale={locale}
              title={title[locale]}
              body={body[locale]}
              bodyRequired={hasBody}
              onTitleChange={(value) =>
                setTitle((current) => ({ ...current, [locale]: value }))
              }
              onBodyChange={(value) =>
                setBody((current) => ({ ...current, [locale]: value }))
              }
            />
          ))}
        </div>
        <Field>
          <FieldLabel htmlFor="announcement-link">
            {t('fields.link')}
          </FieldLabel>
          <Input
            id="announcement-link"
            value={link}
            onChange={(event) => setLink(event.target.value)}
            placeholder="/shop"
            maxLength={MAX_LINK_LENGTH}
            // A path on the site: //host would open another site.
            pattern="/([^\/\\\s]\S*)?"
          />
          <FieldDescription>{t('linkHint')}</FieldDescription>
        </Field>
        <div className="flex justify-end">
          <Button disabled={isPending}>
            <SendIcon data-icon="inline-start" />
            {t('send')}
          </Button>
        </div>
      </form>

      <section className="flex flex-col gap-2 lg:sticky lg:top-6">
        <h2 className="text-muted-foreground text-xs font-semibold tracking-wider uppercase">
          {t('preview')}
        </h2>
        {/* The row is the real bell row, inert, so what the admin sees is what users get. */}
        <ul
          inert
          className="bg-popover overflow-hidden rounded-xl border shadow-sm"
        >
          <NotificationItem notification={preview} onOpen={noop} />
        </ul>
        <p className="text-muted-foreground text-xs">{t('previewHint')}</p>
      </section>

      <AlertDialog open={confirming} onOpenChange={setConfirming}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{t('confirmTitle')}</AlertDialogTitle>
            <AlertDialogDescription>
              {t('confirmDescription')}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>{t('cancel')}</AlertDialogCancel>
            <AlertDialogAction onClick={send}>{t('confirm')}</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}

function LocaleFields({
  locale,
  title,
  body,
  bodyRequired,
  onTitleChange,
  onBodyChange,
}: {
  locale: Locale
  title: string
  body: string
  // bodyRequired asks for a body in this language once another language has one.
  bodyRequired: boolean
  onTitleChange: (value: string) => void
  onBodyChange: (value: string) => void
}) {
  const t = useTranslations('admin.announcements')

  return (
    <div className="grid content-start gap-3">
      <Badge variant="outline" className="w-fit">
        {LOCALE_NAMES[locale]}
      </Badge>
      <Field>
        <FieldLabel htmlFor={`announcement-title-${locale}`}>
          {t('fields.title')}
        </FieldLabel>
        <Input
          id={`announcement-title-${locale}`}
          value={title}
          onChange={(event) => onTitleChange(event.target.value)}
          maxLength={MAX_TITLE_LENGTH}
          required
        />
      </Field>
      <Field>
        <FieldLabel htmlFor={`announcement-body-${locale}`}>
          {t('fields.body')}
        </FieldLabel>
        <textarea
          id={`announcement-body-${locale}`}
          value={body}
          onChange={(event) => onBodyChange(event.target.value)}
          maxLength={MAX_BODY_LENGTH}
          required={bodyRequired}
          className="border-input bg-background min-h-32 resize-y rounded-md border px-3 py-2 text-sm"
        />
        <FieldDescription>
          {t('bodyLength', { length: body.length, max: MAX_BODY_LENGTH })}
        </FieldDescription>
      </Field>
    </div>
  )
}
