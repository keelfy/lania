'use client'

import {
  createContext,
  useContext,
  useState,
  useEffect,
  useMemo,
  type ReactNode,
} from 'react'
import Image from 'next/image'
import { useTranslations } from 'next-intl'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import McUsername from '@/components/ui/mc-username'
import ProfileSelect from '@/components/ui/profile-select'
import { getUserProfiles } from '@/lib/api-endpoints'
import { clientApiFetcher } from '@/lib/client'
import { useAuthStore } from '@/providers/auth-store'
import type { Profile } from '@/models/profile'

type CatalogProfile = {
  username: string
  profileId?: string
  loading: boolean
  failed: boolean
}

const CatalogProfileContext = createContext<CatalogProfile | null>(null)

export function useCatalogProfile() {
  return useContext(CatalogProfileContext)
}

export function CatalogTryOn({
  children,
  showControls = true,
  leadingContent,
}: {
  children: ReactNode
  showControls?: boolean
  leadingContent?: ReactNode
}) {
  const [typedName, setTypedName] = useState('')
  const t = useTranslations('products.catalogTryOn')
  const userId = useAuthStore((state) =>
    state.session?.active ? state.session.identity?.id : undefined,
  )
  const [loaded, setLoaded] = useState<{
    userId: string
    profiles: Profile[]
    failed: boolean
  }>()
  const [selection, setSelection] = useState<{
    userId: string
    profileId: string
  }>()

  useEffect(() => {
    if (!userId) return
    let active = true
    getUserProfiles(clientApiFetcher)
      .then((profiles) => {
        if (active) setLoaded({ userId, profiles, failed: false })
      })
      .catch((error) => {
        console.error(error)
        if (active) setLoaded({ userId, profiles: [], failed: true })
      })
    return () => {
      active = false
    }
  }, [userId])

  const current = userId && loaded?.userId === userId ? loaded : undefined
  const profiles = current?.profiles ?? []
  const selectedProfile =
    profiles.find(
      (profile) =>
        selection?.userId === userId && profile.id === selection?.profileId,
    ) ?? profiles[0]
  const loading = !!userId && !current
  const failed = current?.failed ?? false
  const username = selectedProfile?.username ?? (typedName.trim() || 'Steve')
  const profileId = selectedProfile?.id
  const value = useMemo(
    () => ({ username, profileId, loading, failed }),
    [username, profileId, loading, failed],
  )

  return (
    <CatalogProfileContext value={value}>
      {leadingContent}
      {(showControls || !!userId) && (
        <section className="border-primary/20 bg-primary/5 flex flex-col gap-4 rounded-lg border p-4 sm:flex-row sm:items-center sm:justify-between">
          <div className="space-y-1">
            <h2 className="text-lg font-semibold">
              {selectedProfile || loading ? t('profileTitle') : t('title')}
            </h2>
            <p id="catalog-name-hint" className="text-muted-foreground text-sm">
              {selectedProfile ? t('profileHint') : t('hint')}
            </p>
          </div>
          <div className="flex shrink-0 flex-col gap-2 sm:w-48">
            {selectedProfile || loading ? (
              <>
                <Label htmlFor="catalog-profile">{t('profile')}</Label>
                <ProfileSelect
                  id="catalog-profile"
                  aria-describedby="catalog-name-hint"
                  profiles={profiles}
                  selectedProfileId={profileId}
                  onSelectProfileId={(id) => {
                    if (userId) setSelection({ userId, profileId: id })
                  }}
                  placeholder={loading ? t('loading') : t('profile')}
                  disabled={loading}
                  className="w-full"
                />
              </>
            ) : (
              <>
                <Label htmlFor="catalog-username">{t('username')}</Label>
                <Input
                  id="catalog-username"
                  aria-describedby="catalog-name-hint"
                  value={typedName}
                  onChange={(event) => setTypedName(event.target.value)}
                  maxLength={16}
                  placeholder="Steve"
                  autoComplete="off"
                  spellCheck={false}
                />
              </>
            )}
            {failed && (
              <p role="alert" className="text-destructive text-xs">
                {t('profilesError')}
              </p>
            )}
          </div>
        </section>
      )}
      {children}
    </CatalogProfileContext>
  )
}

export function CatalogNamePreview({
  colors,
  prefix,
  unoptimized,
}: {
  colors?: string[]
  prefix?: string
  unoptimized?: boolean
}) {
  const catalog = useCatalogProfile()
  const username = catalog?.username
  const t = useTranslations('products.catalogTryOn')
  if (!catalog) return null

  return (
    <div className="flex flex-col gap-1.5 rounded-md bg-black/20 px-3 py-3">
      <span className="text-muted-foreground text-xs">{t('preview')}</span>
      <div
        className="flex min-w-0 items-center gap-2 text-lg"
        data-catalog-name-preview
      >
        {prefix && (
          <Image
            src={prefix}
            alt=""
            width={24}
            height={24}
            unoptimized={unoptimized}
            className="size-6 shrink-0 [image-rendering:pixelated]"
          />
        )}
        <McUsername
          username={username}
          colors={colors}
          className="min-w-0 leading-6 break-all"
        />
      </div>
    </div>
  )
}
