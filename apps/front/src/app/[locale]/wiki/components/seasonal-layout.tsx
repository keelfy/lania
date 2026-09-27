'use client'

import { WikiSeason } from '@/lib/wiki-seasons'
import { usePathname } from 'next/navigation'
import type { PageMapItem } from 'nextra'
import { Layout } from 'nextra-theme-docs'
import { ComponentProps, useMemo } from 'react'

type Props = ComponentProps<typeof Layout> & {
  seasons: WikiSeason[]
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

// Nextra layout whose sidebar holds the pages of one season, flattened next to the shared pages. The other
// seasons' folders are dropped, they are reached through the season select in the navbar.
export function SeasonalLayout({ seasons, pageMap, ...props }: Props) {
  const current = useWikiSeason(seasons)

  const seasonPageMap = useMemo(() => {
    const first = pageMap[0]
    const meta = first && 'data' in first ? first.data : {}
    const items = first && 'data' in first ? pageMap.slice(1) : pageMap
    const isSeason = (item: PageMapItem) =>
      'name' in item && seasons.some((season) => season.slug === item.name)

    const seasonFolder = items.find(
      (item) => 'name' in item && item.name === current?.slug,
    )
    return [
      {
        data: {
          ...meta,
          ...(current && { [current.slug]: { display: 'children' } }),
        },
      },
      ...(seasonFolder ? [seasonFolder] : []),
      ...items.filter((item) => !isSeason(item)),
    ]
  }, [pageMap, seasons, current])

  return <Layout {...props} pageMap={seasonPageMap} />
}
