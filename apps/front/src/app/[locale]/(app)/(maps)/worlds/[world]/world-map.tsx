'use client'

import SignInButton from '@/app/[locale]/(app)/(with-navbar)/components/sign-in-button'
import { Button } from '@/components/ui/button'
import LoadingSpinner from '@/components/ui/loading-spinner'
import McUsername from '@/components/ui/mc-username'
import NamePrefixes from '@/components/ui/name-prefixes'
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
  MapDimension,
  MapLive,
  MapMarker,
  MapPlayer,
} from '@/models/claim'
import { Profile } from '@/models/profile'
import { SeasonWorld } from '@/models/season'
import { useAuthStore } from '@/providers/auth-store'
import { FlagIcon, FlagOffIcon, XIcon } from 'lucide-react'
import { useLocale, useTranslations } from 'next-intl'
import Link from 'next/link'
import React from 'react'
import { toast } from 'sonner'
import ChunkMap, { ChunkHover, SelectionKind, chunkKey } from './chunk-map'

type Props = {
  world: SeasonWorld
  mapUrl: string
  dimensions: MapDimension[]
  isAdmin: boolean
  // Shown at the start of the bar over the map, the breadcrumb of the page.
  header: React.ReactNode
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

export default function WorldMap({
  world,
  mapUrl,
  dimensions,
  isAdmin,
  header,
}: Props) {
  const t = useTranslations('claims')
  const locale = useLocale()
  const session = useAuthStore((state) => state.session)
  const signedIn = session?.active === true
  const seasonId = world.seasonId

  // The map opens on the first dimension with claims, usually the overworld.
  const [dimensionName, setDimensionName] = React.useState(
    () =>
      (
        dimensions.find((dimension) =>
          world.claimDimensions.includes(dimension.name),
        ) ?? dimensions[0]
      ).name,
  )
  const dimension =
    dimensions.find((dimension) => dimension.name === dimensionName) ??
    dimensions[0]
  const claimsOn = world.claimDimensions.includes(dimension.name)

  const [data, setData] = React.useState<ChunkClaims>()
  const loadClaims = React.useCallback(
    () =>
      getChunkClaims(clientApiFetcher, world.id, dimension.name)
        .then(setData)
        .catch((error) => errorToast(t('loadFailed'), error)),
    [world.id, dimension.name, t],
  )
  React.useEffect(() => {
    void loadClaims()
  }, [loadClaims])

  // Only the profiles that can claim: the user's own ones with access to the season.
  const [profiles, setProfiles] = React.useState<Profile[]>([])
  const [profileId, setProfileId] = React.useState<string>()
  React.useEffect(() => {
    if (!signedIn || world.claimDimensions.length === 0) return
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
  }, [signedIn, seasonId, world.claimDimensions.length])

  const claimsByKey = React.useMemo(() => {
    const byKey = new Map<string, ChunkClaim>()
    for (const claim of data?.claims ?? []) {
      byKey.set(chunkKey(claim.x, claim.z), claim)
    }
    return byKey
  }, [data])
  const owners = React.useMemo(
    () => new Map(data?.profiles.map((profile) => [profile.id, profile])),
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
  // Where claims are off, claims left from before stay and only an admin can release them.
  const selectionKind = React.useCallback(
    (cx: number, cz: number): SelectionKind | undefined => {
      const claim = claimsByKey.get(chunkKey(cx, cz))
      if (!claim) return claimsOn && profileId ? 'claim' : undefined
      if (claimsOn && claim.profileId === profileId) return 'release'
      return isAdmin ? 'adminRelease' : undefined
    },
    [claimsByKey, claimsOn, profileId, isAdmin],
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
  const [live, setLive] = React.useState<{ dimension: string } & MapLive>()
  React.useEffect(() => {
    let cancelled = false
    const load = () => {
      if (document.hidden) return
      fetch(
        `/api/worlds/${world.id}/live?dimension=${encodeURIComponent(dimension.name)}`,
      )
        .then((response) => (response.ok ? response.json() : undefined))
        .then((data?: MapLive) => {
          if (!cancelled && data) {
            setLive({ dimension: dimension.name, ...data })
          }
        })
        .catch(console.error)
    }
    load()
    const timer = setInterval(load, LIVE_POLL_MS)
    return () => {
      cancelled = true
      clearInterval(timer)
    }
  }, [world.id, dimension.name])
  const liveHere = live?.dimension === dimension.name ? live : undefined

  const [hover, setHover] = React.useState<ChunkHover>()
  const hoverClaim = hover
    ? claimsByKey.get(chunkKey(hover.cx, hover.cz))
    : undefined
  const focusedClaim = focused
    ? claimsByKey.get(chunkKey(...focused))
    : undefined
  const focusedOwner = focusedClaim && owners.get(focusedClaim.profileId)
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
    <div className="flex h-full flex-col">
      <div className="flex min-h-11 flex-wrap items-center justify-between gap-x-4 gap-y-2 px-4 py-1.5">
        {header}
        {dimensions.length > 1 && (
          <Tabs
            value={dimension.name}
            onValueChange={(name) => {
              clearSelection()
              setFocused(undefined)
              setDimensionName(name)
            }}
          >
            <TabsList>
              {dimensions.map((dimension) => (
                <TabsTrigger key={dimension.name} value={dimension.name}>
                  {world.claimDimensions.includes(dimension.name) && (
                    <FlagIcon className="size-3.5" />
                  )}
                  {t(`dimensions.${dimension.type}`)}
                </TabsTrigger>
              ))}
            </TabsList>
          </Tabs>
        )}
      </div>

      <div className="relative min-h-0 flex-1">
        <ChunkMap
          mapUrl={mapUrl}
          dimension={dimension}
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
                {focusedOwner && (
                  <span className="flex flex-wrap items-center gap-1.5">
                    {t('focus.owner')}
                    <Link
                      href={`/${locale}/community/${focusedOwner.username}`}
                      className="hover:bg-accent flex items-center gap-1 rounded-xs px-1"
                    >
                      <NamePrefixes
                        cosmetics={focusedOwner.cosmetics.name}
                        size={16}
                      />
                      <McUsername
                        username={focusedOwner.username}
                        colors={focusedOwner.cosmetics.name.colors?.colors}
                      />
                    </Link>
                  </span>
                )}
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
        <div className="bg-background/80 pointer-events-none absolute bottom-3 left-3 z-[1000] flex max-w-[calc(100%-1.5rem)] flex-wrap gap-x-4 rounded-md px-2 py-1 font-mono text-xs backdrop-blur">
          {hover ? (
            <>
              <span>{t('hover.chunk', { x: hover.cx, z: hover.cz })}</span>
              <span className="text-muted-foreground">
                {t('hover.block', { x: hover.x, z: hover.z })}
              </span>
              {hoverClaim ? (
                <span>
                  {t('hover.claimed', {
                    username: owners.get(hoverClaim.profileId)?.username ?? '?',
                    date: formatDate(hoverClaim.claimedAt),
                  })}
                </span>
              ) : (
                <span className="text-muted-foreground">{t('hover.free')}</span>
              )}
            </>
          ) : (
            <span className="text-muted-foreground">
              {claimsOn || isAdmin ? t('hint') : t('viewHint')}
            </span>
          )}
        </div>
      </div>

      <div className="flex min-h-14 flex-wrap items-center gap-3 border-t px-4 py-2">
        {!claimsOn ? (
          <span className="text-muted-foreground text-sm">
            {world.claimDimensions.length > 0 ? t('claimsOff') : t('viewOnly')}
          </span>
        ) : !signedIn ? (
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
                {t('held', { count: held, limit: world.claimLimit })}
              </span>
            )}
          </>
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
                      world.id,
                      profileId,
                      dimension.name,
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
                      world.id,
                      dimension.name,
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
          {claimsOn && signedIn && profileId && (
            <Button
              disabled={isPending || toClaim.length === 0}
              onClick={() =>
                run(
                  () =>
                    claimChunks(
                      clientApiFetcher,
                      world.id,
                      profileId,
                      dimension.name,
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
        {claimsOn && (
          <p className="text-muted-foreground basis-full text-xs">
            {t('rules')}
          </p>
        )}
      </div>
    </div>
  )
}
