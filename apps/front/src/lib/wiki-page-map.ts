import type { PageMapItem } from 'nextra'

export function getWikiSeasonFolders(pageMap: PageMapItem[]): string[] {
  return pageMap.flatMap((item) =>
    'children' in item &&
    item.children.some((child) => 'name' in child && child.name === 'gameplay')
      ? [item.name]
      : [],
  )
}

// Identify season folders from content, even when the API returns no selectable seasons.
export function getSeasonPageMap(
  pageMap: PageMapItem[],
  seasonFolders: string[],
  currentSlug?: string,
): PageMapItem[] {
  const first = pageMap[0]
  const meta = first && 'data' in first ? first.data : {}
  const items = first && 'data' in first ? pageMap.slice(1) : pageMap
  const current = items.find(
    (item) => 'name' in item && item.name === currentSlug,
  )

  return [
    {
      data: {
        ...meta,
        ...(currentSlug && { [currentSlug]: { display: 'children' } }),
      },
    },
    ...(current ? [current] : []),
    ...items.filter(
      (item) => !('name' in item && seasonFolders.includes(item.name)),
    ),
  ]
}
