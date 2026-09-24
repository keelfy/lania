'use client'

import { showResync } from '@/components/profile-resync-card'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import ProfilePicker from '@/components/profile-picker'
import { mergeProfiles, previewMergeProfiles } from '@/lib/api-endpoints'
import { clientApiFetcher } from '@/lib/client'
import { errorToast } from '@/lib/toasts'
import {
  AdminProfile,
  ProfileMergeCounts,
  ProfileMergeSummary,
} from '@/models/admin'
import { AlertTriangleIcon, GitMergeIcon } from 'lucide-react'
import { useTranslations } from 'next-intl'
import { useRouter } from 'next/navigation'
import React from 'react'
import { toast } from 'sonner'

type Props = {
  profileId: string
  profileUsername: string
  locale: string
}

// Only the fields with something to show, in reading order.
const COUNT_KEYS = [
  'playtimeMoved',
  'playtimeSummed',
  'accessesMoved',
  'accessesDropped',
  'violationsMoved',
  'nameColorOptionsMoved',
  'nameColorOptionsDropped',
  'namePrefixOptionsMoved',
  'namePrefixOptionsDropped',
  'seasonCosmeticsMoved',
  'seasonCosmeticsDropped',
  'prefixesMoved',
  'prefixesDropped',
  'orderItemsMoved',
  'basketItemsMoved',
  'basketItemsDropped',
  'notificationsRepointed',
  'screenshotAuthorsMoved',
  'screenshotAuthorsDropped',
] as const satisfies readonly (keyof ProfileMergeCounts)[]

function MergeCounts({ counts }: { counts: ProfileMergeCounts }) {
  const t = useTranslations('admin.profiles.merge.counts')
  const rows = COUNT_KEYS.filter((key) => counts[key] > 0)

  if (rows.length === 0) return null

  return (
    <ul className="text-muted-foreground flex flex-col gap-0.5 text-sm">
      {rows.map((key) => (
        <li key={key}>{t(key, { count: counts[key] })}</li>
      ))}
    </ul>
  )
}

function MergeBlockers({ summary }: { summary: ProfileMergeSummary }) {
  const t = useTranslations('admin.profiles.merge.blockers')

  return (
    <ul className="text-destructive flex flex-col gap-1 text-sm">
      {summary.blockers?.map((blocker, i) => (
        <li key={i} className="flex items-start gap-1.5">
          <AlertTriangleIcon className="mt-0.5 size-4 shrink-0" />
          {blocker.kind === 'live-season-playtime'
            ? t('live-season-playtime', {
                seasons: (blocker.seasonNames ?? []).join(', '),
              })
            : t(blocker.kind)}
        </li>
      ))}
    </ul>
  )
}

export default function MergeCard({
  profileId,
  profileUsername,
  locale,
}: Props) {
  const t = useTranslations('admin.profiles.merge')
  const tResync = useTranslations('profileResync')
  const tRole = useTranslations('admin.profiles.role.names')
  const router = useRouter()
  const [open, setOpen] = React.useState(false)

  const [target, setTarget] = React.useState<AdminProfile | undefined>()
  const [summary, setSummary] = React.useState<
    ProfileMergeSummary | undefined
  >()
  const [previewFailed, setPreviewFailed] = React.useState(false)
  const [confirmText, setConfirmText] = React.useState('')
  const [isPreviewPending, startPreviewTransition] = React.useTransition()
  const [isMergePending, startMergeTransition] = React.useTransition()

  const reset = () => {
    setTarget(undefined)
    setSummary(undefined)
    setPreviewFailed(false)
    setConfirmText('')
  }

  const selectTarget = (profile: AdminProfile) => {
    setTarget(profile)
    setSummary(undefined)
    setPreviewFailed(false)
    setConfirmText('')

    startPreviewTransition(async () => {
      try {
        const preview = await previewMergeProfiles(
          clientApiFetcher,
          profileId,
          profile.id,
        )
        setSummary(preview)
      } catch (error) {
        setPreviewFailed(true)
        errorToast(t('previewFailed'), error)
      }
    })
  }

  const handleMerge = () => {
    if (isMergePending || !summary?.canMerge || confirmText !== profileUsername)
      return

    startMergeTransition(async () => {
      try {
        const result = await mergeProfiles(
          clientApiFetcher,
          profileId,
          summary.targetProfileId,
        )
        toast.success(t('merged', { username: summary.targetUsername }))
        if (result.resync) showResync(result.resync, tResync)
        setOpen(false)
        router.push(`/${locale}/admin/profiles/${summary.targetProfileId}`)
      } catch (error) {
        errorToast(t('mergeFailed'), error)
      }
    })
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-lg">{t('title')}</CardTitle>
        <CardDescription>{t('description')}</CardDescription>
      </CardHeader>
      <CardContent>
        <Dialog
          open={open}
          onOpenChange={(next) => {
            setOpen(next)
            if (!next) reset()
          }}
        >
          <DialogTrigger asChild>
            <Button variant="outline" className="w-fit">
              <GitMergeIcon className="size-4" />
              {t('open')}
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>
                {t('dialogTitle', { username: profileUsername })}
              </DialogTitle>
              <DialogDescription>{t('dialogDescription')}</DialogDescription>
            </DialogHeader>

            <div className="flex flex-col gap-4">
              {!target ? (
                <div className="flex flex-col gap-2">
                  <Label>{t('targetLabel')}</Label>
                  <ProfilePicker
                    onSelect={selectTarget}
                    excludeIds={[profileId]}
                    placeholder={t('searchPlaceholder')}
                    noResultsLabel={t('noResults')}
                  />
                </div>
              ) : (
                <div className="flex flex-col gap-4">
                  <div className="flex items-center justify-between gap-2">
                    <p className="text-sm">
                      {t('targetSelected', {
                        source: profileUsername,
                        target: target.username,
                      })}
                    </p>
                    <Button
                      type="button"
                      variant="ghost"
                      size="sm"
                      onClick={reset}
                      disabled={isMergePending}
                    >
                      {t('changeTarget')}
                    </Button>
                  </div>

                  {isPreviewPending && (
                    <p className="text-muted-foreground text-sm">
                      {t('previewLoading')}
                    </p>
                  )}

                  {summary && !summary.canMerge && (
                    <MergeBlockers summary={summary} />
                  )}

                  {summary && summary.canMerge && (
                    <div className="flex flex-col gap-3">
                      <MergeCounts counts={summary.counts} />
                      {summary.roleBefore !== summary.roleAfter && (
                        <p className="text-sm">
                          {t('roleChange', {
                            before: tRole(summary.roleBefore),
                            after: tRole(summary.roleAfter),
                          })}
                        </p>
                      )}
                      <div className="flex flex-col gap-2">
                        <Label>
                          {t('confirmLabel', { username: profileUsername })}
                        </Label>
                        <Input
                          value={confirmText}
                          onChange={(e) => setConfirmText(e.target.value)}
                          placeholder={profileUsername}
                          aria-label={t('confirmLabel', {
                            username: profileUsername,
                          })}
                        />
                      </div>
                    </div>
                  )}

                  {previewFailed && (
                    <p className="text-destructive text-sm">
                      {t('previewFailed')}
                    </p>
                  )}
                </div>
              )}
            </div>

            <DialogFooter>
              <Button
                type="button"
                variant="destructive"
                disabled={
                  !summary?.canMerge ||
                  confirmText !== profileUsername ||
                  isMergePending
                }
                onClick={handleMerge}
              >
                {t('submit')}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </CardContent>
    </Card>
  )
}
