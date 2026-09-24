'use client'

import { Button } from '@/components/ui/button'
import LoadingSpinner from '@/components/ui/loading-spinner'
import ProfileSelect from '@/components/ui/profile-select'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import {
  adminReleaseChunks,
  claimChunks,
  getChunkClaims,
  getUserProfiles,
  releaseChunks,
} from '@/lib/api-endpoints'
import { clientApiFetcher } from '@/lib/client'
import { errorToast } from '@/lib/toasts'
import {
  ChunkClaim,
  ChunkClaims,
  ChunkPos,
  MAX_CLAIM_BATCH,
  MapLive,
  MapMarker,
  MapPlayer,
  MapWorld,
} from '@/models/claim'
import { Profile } from '@/models/profile'
import { useAuthStore } from '@/providers/auth-store'
import { FlagIcon, FlagOffIcon, XIcon } from 'lucide-react'
import { useLocale, useTranslations } from 'next-intl'
import Link from 'next/link'
import React from 'react'
import { toast } from 'sonner'
import SignInButton from '../../components/sign-in-button'
import ChunkMap, { ChunkHover, SelectionKind, chunkKey } from './chunk-map'

type Props = {
  seasonId: string
  mapUrl: string
  claimLimit: number
  worlds: MapWorld[]
  isAdmin: boolean
}

type Selected = { x: number; z: number; kind: SelectionKind }

// squaremap itself refreshes players every second; a few seconds late is fine for spotting who is where.
const LIVE_POLL_MS = 3000
const NO_MARKERS: MapMarker[] = []
const NO_PLAYERS: MapPlayer[] = []

// A stable hue per profile, so a player's claims keep their colour between visits.
function profileColor(profileId: string) {
  let hash = 0
  for (const char of profileId) hash = (hash * 31 + char.charCodeAt(0)) | 0
  return `hsl(${Math.abs(hash) % 360} 80% 55%)`
}

export default function ClaimsMap({
  seasonId,
  mapUrl,
  claimLimit,
  worlds,
  isAdmin,
}: Props) {
  const t = useTranslations('claims')
  const locale = useLocale()
  const session = useAuthStore((state) => state.session)
  const signedIn = session?.active === true

  const [worldName, setWorldName] = React.useState(worlds[0].name)
  const world = worlds.find((world) => world.name === worldName) ?? worlds[0]

  const [data, setData] = React.useState<ChunkClaims>()
  const loadClaims = React.useCallback(
    () =>
      getChunkClaims(clientApiFetcher, seasonId, world.name)
        .then(setData)
        .catch((error) => errorToast(t('loadFailed'), error)),
    [seasonId, world.name, t],
  )
  React.useEffect(() => {
    void loadClaims()
  }, [loadClaims])

  // Only the profiles that can claim: the user's own ones with access to the season.
  const [profiles, setProfiles] = React.useState<Profile[]>([])
  const [profileId, setProfileId] = React.useState<string>()
  React.useEffect(() => {
    if (!signedIn) return
    let cancelled = false
    getUserProfiles(clientApiFetcher, undefined, seasonId)
      .then((profiles) => {
        if (cancelled) return
        const withAccess = profiles.filter((profile) =>
          profile.accesses.some(
            (access) =>
              access.seasonId === seasonId && access.status === 'active',
          ),
        )
        setProfiles(withAccess)
        setProfileId((current) => current ?? withAccess[0]?.id)
      })
      .catch(console.error)
    return () => {
      cancelled = true
    }
  }, [signedIn, seasonId])

  const claimsByKey = React.useMemo(() => {
    const byKey = new Map<string, ChunkClaim>()
    for (const claim of data?.claims ?? []) {
      byKey.set(chunkKey(claim.x, claim.z), claim)
    }
    return byKey
  }, [data])
  const usernames = React.useMemo(
    () =>
      new Map(data?.profiles.map((profile) => [profile.id, profile.username])),
    [data],
  )
  const claimColors = React.useMemo(() => {
    const colors = new Map<string, string>()
    for (const [key, claim] of claimsByKey) {
      colors.set(key, profileColor(claim.profileId))
    }
    return colors
  }, [claimsByKey])

  // What clicking the chunk would do, or undefined when the user cannot touch it.
  const selectionKind = React.useCallback(
    (cx: number, cz: number): SelectionKind | undefined => {
      const claim = claimsByKey.get(chunkKey(cx, cz))
      if (!claim) return profileId ? 'claim' : undefined
      if (claim.profileId === profileId) return 'release'
      return isAdmin ? 'adminRelease' : undefined
    },
    [claimsByKey, profileId, isAdmin],
  )

  const [selection, setSelection] = React.useState<Map<string, Selected>>(
    () => new Map(),
  )
  const clearSelection = React.useCallback(() => setSelection(new Map()), [])

  // The chunk last clicked, whose owner and claim date the card over the map shows.
  const [focused, setFocused] = React.useState<ChunkPos>()

  // A click focuses the chunk and toggles it. A drag adds every chunk it covers,
  // or removes them when it started on a chunk that was already selected.
  const onChunkArea = React.useCallback(
    ([ax, az]: ChunkPos, [bx, bz]: ChunkPos) => {
      if (ax === bx && az === bz) setFocused([ax, az])
      setSelection((current) => {
        const next = new Map(current)
        const removing = next.has(chunkKey(ax, az))
        for (let x = Math.min(ax, bx); x <= Math.max(ax, bx); x++) {
          for (let z = Math.min(az, bz); z <= Math.max(az, bz); z++) {
            const key = chunkKey(x, z)
            if (removing) {
              next.delete(key)
              continue
            }
            if (next.size >= MAX_CLAIM_BATCH) return next
            const kind = selectionKind(x, z)
            if (kind) next.set(key, { x, z, kind })
          }
        }
        return next
      })
    },
    [selectionKind],
  )

  // Markers and online players, polled through the site: squaremap cannot be read across origins.
  const [live, setLive] = React.useState<{ world: string } & MapLive>()
  React.useEffect(() => {
    let cancelled = false
    const load = () => {
      if (document.hidden) return
      fetch(`/api/claims/live?world=${encodeURIComponent(world.name)}`)
        .then((response) => (response.ok ? response.json() : undefined))
        .then((data?: MapLive) => {
          if (!cancelled && data) setLive({ world: world.name, ...data })
        })
        .catch(console.error)
    }
    load()
    const timer = setInterval(load, LIVE_POLL_MS)
    return () => {
      cancelled = true
      clearInterval(timer)
    }
  }, [world.name])
  const liveHere = live?.world === world.name ? live : undefined

  const [hover, setHover] = React.useState<ChunkHover>()
  const hoverClaim = hover
    ? claimsByKey.get(chunkKey(hover.cx, hover.cz))
    : undefined
  const focusedClaim = focused
    ? claimsByKey.get(chunkKey(...focused))
    : undefined
  const formatDate = (millis: number) =>
    new Date(millis).toLocaleString(locale, {
      dateStyle: 'medium',
      timeStyle: 'short',
    })

  const byKind = (kind: SelectionKind): ChunkPos[] =>
    [...selection.values()]
      .filter((chunk) => chunk.kind === kind)
      .map(({ x, z }) => [x, z])
  const toClaim = byKind('claim')
  const toRelease = byKind('release')
  const toAdminRelease = byKind('adminRelease')
  const held = profileId
    ? (data?.claims.filter((claim) => claim.profileId === profileId).length ??
      0)
    : 0

  const [isPending, startTransition] = React.useTransition()
  const run = (action: () => Promise<void>, done: string, failed: string) =>
    startTransition(async () => {
      try {
        await action()
        toast.success(done)
      } catch (error) {
        errorToast(failed, error)
      }
      // The selection was made against the claims before the change.
      clearSelection()
      await loadClaims()
    })

  return (
    <div className="flex flex-col gap-3">
      {worlds.length > 1 && (
        <Tabs
          value={world.name}
          onValueChange={(name) => {
            clearSelection()
            setFocused(undefined)
            setWorldName(name)
          }}
        >
          <TabsList>
            {worlds.map((world) => (
              <TabsTrigger key={world.name} value={world.name}>
                {t(`worlds.${world.type}`)}
              </TabsTrigger>
            ))}
          </TabsList>
        </Tabs>
      )}

      <div className="relative">
        <ChunkMap
          mapUrl={mapUrl}
          world={world}
          claimColors={claimColors}
          selection={selection}
          focused={focused}
          markers={liveHere?.markers ?? NO_MARKERS}
          players={liveHere?.players ?? NO_PLAYERS}
          onChunkArea={onChunkArea}
          onChunkHover={setHover}
        />
        {focused && (
          <div className="bg-background/90 absolute top-3 right-3 z-[1000] flex w-64 flex-col gap-1 rounded-lg border p-3 text-sm shadow-lg backdrop-blur">
            <div className="flex items-center justify-between gap-2">
              <span className="font-mono font-semibold">
                {t('hover.chunk', { x: focused[0], z: focused[1] })}
              </span>
              <Button
                variant="ghost"
                size="icon"
                className="size-6"
                aria-label={t('focus.close')}
                onClick={() => setFocused(undefined)}
              >
                <XIcon className="size-4" />
              </Button>
            </div>
            <span className="text-muted-foreground font-mono text-xs">
              {t('focus.blocks', {
                x1: focused[0] * 16,
                x2: focused[0] * 16 + 15,
                z1: focused[1] * 16,
                z2: focused[1] * 16 + 15,
              })}
            </span>
            {focusedClaim ? (
              <>
                <span>
                  {t.rich('focus.owner', {
                    username: usernames.get(focusedClaim.profileId) ?? '?',
                    link: (chunks) => (
                      <Link
                        href={`/${locale}/community/${usernames.get(focusedClaim.profileId)}`}
                        className="font-semibold underline"
                        style={{ color: profileColor(focusedClaim.profileId) }}
                      >
                        {chunks}
                      </Link>
                    ),
                  })}
                </span>
                <span className="text-muted-foreground">
                  {t('focus.since', {
                    date: formatDate(focusedClaim.claimedAt),
                  })}
                </span>
              </>
            ) : (
              <span className="text-muted-foreground">{t('hover.free')}</span>
            )}
          </div>
        )}
      </div>

      <div className="text-muted-foreground flex min-h-5 flex-wrap gap-x-4 font-mono text-sm">
        {hover ? (
          <>
            <span>{t('hover.chunk', { x: hover.cx, z: hover.cz })}</span>
            <span>{t('hover.block', { x: hover.x, z: hover.z })}</span>
            {hoverClaim ? (
              <span className="text-foreground">
                {t('hover.claimed', {
                  username: usernames.get(hoverClaim.profileId) ?? '?',
                  date: formatDate(hoverClaim.claimedAt),
                })}
              </span>
            ) : (
              <span>{t('hover.free')}</span>
            )}
          </>
        ) : (
          <span>{t('hint')}</span>
        )}
      </div>

      <div className="flex flex-wrap items-center gap-3 rounded-lg border p-3">
        {!signedIn ? (
          <>
            <span className="text-muted-foreground text-sm">{t('signIn')}</span>
            <SignInButton />
          </>
        ) : profiles.length === 0 && !isAdmin ? (
          <span className="text-muted-foreground text-sm">
            {t.rich('noAccess', {
              link: (chunks) => (
                <Link
                  href={`/${locale}/obtain-access`}
                  className="text-foreground underline"
                >
                  {chunks}
                </Link>
              ),
            })}
          </span>
        ) : (
          <>
            {profiles.length > 0 && (
              <ProfileSelect
                profiles={profiles}
                selectedProfileId={profileId}
                onSelectProfileId={(id) => {
                  // What the selected chunks would do depends on whose they are.
                  clearSelection()
                  setProfileId(id)
                }}
                className="w-48"
              />
            )}
            {profileId && (
              <span className="text-muted-foreground text-sm">
                {t('held', { count: held, limit: claimLimit })}
              </span>
            )}
            <div className="ms-auto flex flex-wrap gap-2">
              {selection.size > 0 && (
                <Button
                  variant="ghost"
                  onClick={clearSelection}
                  disabled={isPending}
                >
                  <XIcon className="size-4" />
                  {t('clear')}
                </Button>
              )}
              {toRelease.length > 0 && profileId && (
                <Button
                  variant="outline"
                  disabled={isPending}
                  onClick={() =>
                    run(
                      () =>
                        releaseChunks(
                          clientApiFetcher,
                          seasonId,
                          profileId,
                          world.name,
                          toRelease,
                        ),
                      t('released', { count: toRelease.length }),
                      t('releaseFailed'),
                    )
                  }
                >
                  <FlagOffIcon className="size-4" />
                  {t('release', { count: toRelease.length })}
                </Button>
              )}
              {toAdminRelease.length > 0 && (
                <Button
                  variant="destructive"
                  disabled={isPending}
                  onClick={() =>
                    run(
                      () =>
                        adminReleaseChunks(
                          clientApiFetcher,
                          seasonId,
                          world.name,
                          toAdminRelease,
                        ),
                      t('released', { count: toAdminRelease.length }),
                      t('releaseFailed'),
                    )
                  }
                >
                  <FlagOffIcon className="size-4" />
                  {t('adminRelease', { count: toAdminRelease.length })}
                </Button>
              )}
              {profileId && (
                <Button
                  disabled={isPending || toClaim.length === 0}
                  onClick={() =>
                    run(
                      () =>
                        claimChunks(
                          clientApiFetcher,
                          seasonId,
                          profileId,
                          world.name,
                          toClaim,
                        ),
                      t('claimed', { count: toClaim.length }),
                      t('claimFailed'),
                    )
                  }
                >
                  {isPending ? (
                    <LoadingSpinner className="size-4" />
                  ) : (
                    <FlagIcon className="size-4" />
                  )}
                  {t('claim', { count: toClaim.length })}
                </Button>
              )}
            </div>
          </>
        )}
      </div>
      <p className="text-muted-foreground text-xs">{t('rules')}</p>
    </div>
  )
}
