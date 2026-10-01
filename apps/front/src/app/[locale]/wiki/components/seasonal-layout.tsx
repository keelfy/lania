'use client'

import { WikiSeason } from '@/lib/wiki-seasons'
import { getSeasonPageMap } from '@/lib/wiki-page-map'
import { usePathname } from 'next/navigation'
import { Layout } from 'nextra-theme-docs'
import { ComponentProps, useMemo } from 'react'

type Props = ComponentProps<typeof Layout> & {
  seasons: WikiSeason[]
  seasonFolders: string[]
}

// The season whose wiki the path belongs to: /<locale>/wiki/<slug>/... Shared pages (rules, legal) have no
// season in the path and show the primary one.
export function useWikiSeason(seasons: WikiSeason[]) {
  const segment = usePathname().split('/')[3]
  return (
    seasons.find((season) => season.slug === segment) ??
    seasons.find((season) => season.isPrimary)
  )
}

// The sidebar holds only the selected season, flattened next to shared pages.
export function SeasonalLayout({
  seasons,
  seasonFolders,
  pageMap,
  ...props
}: Props) {
  const current = useWikiSeason(seasons)

  const seasonPageMap = useMemo(
    () => getSeasonPageMap(pageMap, seasonFolders, current?.slug),
    [pageMap, seasonFolders, current?.slug],
  )

  return <Layout {...props} pageMap={seasonPageMap} />
}
