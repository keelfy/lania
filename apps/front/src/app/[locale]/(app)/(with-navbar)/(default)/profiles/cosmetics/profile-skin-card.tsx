'use client'

import { Button } from '@/components/ui/button'
import {
  Card,
  CardAction,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import {
  clearProfileSkin,
  setProfileSkinNickname,
  uploadProfileSkin,
} from '@/lib/api-endpoints'
import { clientApiFetcher } from '@/lib/client'
import {
  capeUrl,
  SKIN_CHANGED_EVENT,
  skinTexture,
  STEVE_SKIN_URL,
} from '@/lib/skin'
import { initializeViewer } from '@/lib/skin-viewer'
import { errorToast } from '@/lib/toasts'
import { cn } from '@/lib/utils'
import { Profile, SkinChange } from '@/models/profile'
import { ImageUpIcon, RotateCcwIcon } from 'lucide-react'
import { useTranslations } from 'next-intl'
import { useRouter } from 'next/navigation'
import React from 'react'
import { toast } from 'sonner'

type Props = {
  // The profile with the skin it wears in the season.
  profile: Profile
  seasonId: string | undefined
  // Named when the player has several running seasons to pick from in the cosmetics card.
  seasonName?: string
  // A skin can be changed in a running season the profile has access to.
  canChange: boolean
}

// What the API checks too: a whole PNG of 64x64, or 64x32 in the legacy format.
const MAX_FILE_BYTES = 64 * 1024
const NICKNAME_PATTERN = /^[A-Za-z0-9_]{3,16}$/
const PREVIEW_WIDTH = 160
const PREVIEW_HEIGHT = 220

// Messages the API answers with.
const ERROR_KEYS: Record<string, 'invalidFile' | 'notLicensed' | 'cooldown'> = {
  skin_file_invalid: 'invalidFile',
  skin_nickname_not_licensed: 'notLicensed',
  skin_cooldown: 'cooldown',
}

type Draft = {
  file: File
  url: string
  slim: boolean
}

// Reads a skin file the way the API does and guesses the model: a slim skin leaves the outer column of its
// arm transparent. A legacy 64x32 skin has no slim model.
async function readSkinFile(
  file: File,
): Promise<Omit<Draft, 'file'> | undefined> {
  if (file.type !== 'image/png' || file.size > MAX_FILE_BYTES) return undefined
  const url = URL.createObjectURL(file)
  try {
    const image = new Image()
    image.src = url
    await image.decode()
    const { naturalWidth: width, naturalHeight: height } = image
    if (width !== 64 || (height !== 64 && height !== 32)) {
      URL.revokeObjectURL(url)
      return undefined
    }
    let slim = false
    if (height === 64) {
      const canvas = document.createElement('canvas')
      canvas.width = width
      canvas.height = height
      const context = canvas.getContext('2d')
      context?.drawImage(image, 0, 0)
      slim = context?.getImageData(54, 20, 1, 1).data[3] === 0
    }
    return { url, slim }
  } catch {
    URL.revokeObjectURL(url)
    return undefined
  }
}

// Changes the skin the player wears in game in a season, the same way /skin does on the server.
export default function ProfileSkinCard({
  profile,
  seasonId,
  seasonName,
  canChange,
}: Props) {
  const t = useTranslations('profiles.skin')
  const router = useRouter()
  const [draft, setDraft] = React.useState<Draft>()
  const [nickname, setNickname] = React.useState('')
  const [dragging, setDragging] = React.useState(false)
  const [isSaving, startSaving] = React.useTransition()

  // The model switch replaces the draft but keeps its file.
  const draftUrl = draft?.url
  React.useEffect(() => {
    if (!draftUrl) return
    return () => URL.revokeObjectURL(draftUrl)
  }, [draftUrl])

  const current = skinTexture(profile)
  const preview = draft
    ? { url: draft.url, slim: draft.slim, cape: '' }
    : {
        url: current?.url ?? STEVE_SKIN_URL,
        slim: current?.slim ?? false,
        cape: capeUrl(profile),
      }

  const pickFile = async (file: File | undefined) => {
    if (!file) return
    const read = await readSkinFile(file)
    if (!read) {
      toast.error(t('errors.invalidFile'), { description: t('fileHint') })
      return
    }
    setDraft({ file, ...read })
  }

  const change = (
    request: () => Promise<SkinChange>,
    onApplied?: () => void,
  ) => {
    if (!seasonId) return
    startSaving(async () => {
      try {
        const result = await request()
        if (result.status === 'pending') {
          toast.info(t('pending'), { description: t('pendingHint') })
        } else {
          toast.success(t('applied'), { description: t('appliedHint') })
        }
        onApplied?.()
        window.dispatchEvent(
          new CustomEvent(SKIN_CHANGED_EVENT, { detail: profile.id }),
        )
        router.refresh()
      } catch (error) {
        const key =
          error instanceof Error ? ERROR_KEYS[error.message.trim()] : undefined
        if (key) {
          toast.error(t(`errors.${key}`))
          return
        }
        errorToast(t('errors.failed'), error)
      }
    })
  }

  const uploadDraft = () => {
    if (!draft || !seasonId) return
    change(
      () =>
        uploadProfileSkin(
          clientApiFetcher,
          profile.id,
          seasonId,
          draft.file,
          draft.slim ? 'slim' : 'classic',
        ),
      () => setDraft(undefined),
    )
  }

  const copyNickname = (event: React.FormEvent) => {
    event.preventDefault()
    if (!seasonId || !NICKNAME_PATTERN.test(nickname)) return
    change(
      () =>
        setProfileSkinNickname(
          clientApiFetcher,
          profile.id,
          seasonId,
          nickname,
        ),
      () => setNickname(''),
    )
  }

  const reset = () => {
    if (!seasonId) return
    change(() => clearProfileSkin(clientApiFetcher, profile.id, seasonId))
  }

  const disabled = !canChange || isSaving

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-lg">{t('title')}</CardTitle>
        <CardDescription>
          {canChange ? t('description') : t('noAccess')}
          {seasonName && ` ${t('season', { season: seasonName })}`}
        </CardDescription>
        {profile.skin && (
          <CardAction>
            <Button
              variant="outline"
              size="sm"
              onClick={reset}
              disabled={disabled}
            >
              <RotateCcwIcon className="size-4" />
              {t('reset')}
            </Button>
          </CardAction>
        )}
      </CardHeader>
      <CardContent className="grid gap-6 sm:grid-cols-[auto_minmax(0,1fr)]">
        <div className="flex flex-col items-center gap-2">
          <SkinPreview
            url={preview.url}
            slim={preview.slim}
            cape={preview.cape}
          />
          <p className="text-muted-foreground text-xs">
            {t(draft ? 'preview.draft' : 'preview.current')}
          </p>
        </div>

        <Tabs defaultValue="file" className="min-w-0">
          <TabsList>
            <TabsTrigger value="file">{t('tabs.file')}</TabsTrigger>
            <TabsTrigger value="nickname">{t('tabs.nickname')}</TabsTrigger>
          </TabsList>

          <TabsContent value="file" className="flex flex-col gap-4">
            <label
              onDragOver={(event) => {
                event.preventDefault()
                setDragging(true)
              }}
              onDragLeave={() => setDragging(false)}
              onDrop={(event) => {
                event.preventDefault()
                setDragging(false)
                if (!disabled) void pickFile(event.dataTransfer.files[0])
              }}
              className={cn(
                'border-input hover:bg-accent/50 flex min-h-28 cursor-pointer flex-col items-center justify-center gap-2 rounded-lg border border-dashed p-4 text-center transition-colors',
                dragging && 'border-primary bg-accent/50',
                disabled && 'pointer-events-none opacity-50',
              )}
            >
              <ImageUpIcon className="text-muted-foreground size-6" />
              <span className="text-sm font-medium">
                {draft ? draft.file.name : t('drop')}
              </span>
              <span className="text-muted-foreground text-xs">
                {t('fileHint')}
              </span>
              <input
                type="file"
                accept="image/png"
                className="sr-only"
                disabled={disabled}
                onChange={(event) => {
                  void pickFile(event.target.files?.[0])
                  event.target.value = ''
                }}
              />
            </label>
            <div className="flex items-center justify-between gap-4">
              <div className="flex items-center gap-2">
                <Switch
                  id="skin-slim"
                  checked={draft?.slim ?? false}
                  onCheckedChange={(slim) =>
                    setDraft((value) => value && { ...value, slim })
                  }
                  disabled={!draft || disabled}
                />
                <Label htmlFor="skin-slim" className="text-sm">
                  {t('slim')}
                </Label>
              </div>
              <Button onClick={uploadDraft} disabled={!draft || disabled}>
                {t(isSaving ? 'saving' : 'upload')}
              </Button>
            </div>
          </TabsContent>

          <TabsContent value="nickname">
            <form onSubmit={copyNickname} className="flex flex-col gap-4">
              <p className="text-muted-foreground text-sm">
                {t('nicknameHint')}
              </p>
              <div className="flex gap-2">
                <Input
                  value={nickname}
                  onChange={(event) => setNickname(event.target.value)}
                  maxLength={16}
                  placeholder="Notch"
                  aria-label={t('nickname')}
                  aria-invalid={
                    nickname.length > 0 && !NICKNAME_PATTERN.test(nickname)
                  }
                  disabled={disabled}
                />
                <Button
                  type="submit"
                  disabled={!NICKNAME_PATTERN.test(nickname) || disabled}
                >
                  {t(isSaving ? 'saving' : 'copy')}
                </Button>
              </div>
            </form>
          </TabsContent>
          <p className="text-muted-foreground text-xs">{t('appliesHint')}</p>
        </Tabs>
      </CardContent>
    </Card>
  )
}

function SkinPreview({
  url,
  slim,
  cape,
}: {
  url: string
  slim: boolean
  cape: string
}) {
  const canvasRef = React.useRef<HTMLCanvasElement>(null)

  React.useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas) return
    const viewer = initializeViewer(url, cape, slim, 'dark', {
      canvas,
      width: PREVIEW_WIDTH,
      height: PREVIEW_HEIGHT,
      transparent: true,
    })
    return () => viewer.dispose()
  }, [url, slim, cape])

  return (
    <canvas
      ref={canvasRef}
      className="bg-muted/40 rounded-lg"
      style={{ width: PREVIEW_WIDTH, height: PREVIEW_HEIGHT }}
    />
  )
}
