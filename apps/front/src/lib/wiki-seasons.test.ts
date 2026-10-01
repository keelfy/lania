import { expect, mock, test } from 'bun:test'
import type { PageMapItem } from 'nextra'
import { getSeasonPageMap, getWikiSeasonFolders } from './wiki-page-map'

const fetchSeasons = mock(async () => [] as object[])
mock.module('./fetcher', () => ({ apiFetcher: fetchSeasons }))
const { getWikiSeasons } = await import('./wiki-seasons')

const pageMap: PageMapItem[] = [
  { data: { rules: { title: 'Rules' } } },
  { name: 'rules', route: '/wiki/rules' },
  {
    name: 'legal',
    route: '/wiki/legal',
    children: [{ name: 'privacy', route: '/wiki/legal/privacy' }],
  },
  ...['lania-iv', 'lania-v'].map((name) => ({
    name,
    route: `/wiki/${name}`,
    children: [
      { name: 'gameplay', route: `/wiki/${name}/gameplay`, children: [] },
    ],
  })),
]

test('season folders exclude shared legal pages', () => {
  expect(getWikiSeasonFolders(pageMap)).toEqual(['lania-iv', 'lania-v'])
})

test('sidebar displays only the selected season alongside shared pages', () => {
  const result = getSeasonPageMap(
    pageMap,
    getWikiSeasonFolders(pageMap),
    'lania-v',
  )
  expect(result.slice(1).map((item) => 'name' in item && item.name)).toEqual([
    'lania-v',
    'rules',
    'legal',
  ])
  expect(result[0]).toEqual({
    data: { rules: { title: 'Rules' }, 'lania-v': { display: 'children' } },
  })
})

test('an unavailable API never exposes every season in the sidebar', () => {
  const result = getSeasonPageMap(pageMap, getWikiSeasonFolders(pageMap))
  expect(result.slice(1).map((item) => 'name' in item && item.name)).toEqual([
    'rules',
    'legal',
  ])
})

test('active and archived seasons with content are selectable; API primary is preserved', async () => {
  fetchSeasons.mockResolvedValueOnce([
    { name: 'Lania IV', isActive: false, isPrimary: false },
    { name: 'Lania V', isActive: true, isPrimary: true },
    { name: 'Lania VI', isActive: true, isPrimary: false },
  ])
  expect(await getWikiSeasons(['lania-iv', 'lania-v'])).toEqual([
    { slug: 'lania-iv', name: 'Lania IV', isPrimary: false },
    { slug: 'lania-v', name: 'Lania V', isPrimary: true },
  ])
})

test('selecting an archive replaces the primary season in the sidebar', () => {
  const result = getSeasonPageMap(
    pageMap,
    getWikiSeasonFolders(pageMap),
    'lania-iv',
  )
  expect(result.slice(1).map((item) => 'name' in item && item.name)).toEqual([
    'lania-iv',
    'rules',
    'legal',
  ])
})

test('an API error produces no selectable seasons', async () => {
  fetchSeasons.mockRejectedValueOnce(new Error('API unavailable'))
  expect(await getWikiSeasons(['lania-iv', 'lania-v'])).toEqual([])
})
