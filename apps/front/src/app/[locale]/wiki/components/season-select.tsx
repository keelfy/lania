'use client'

import { WikiSeason } from '@/lib/wiki-seasons'
import { usePathname, useRouter } from 'next/navigation'
import { Select } from 'nextra/components'
import { useWikiSeason } from './seasonal-layout'

type Props = {
  seasons: WikiSeason[]
  // Page routes of the wiki without the locale: /wiki/lania-v/commands.
  routes: string[]
  label: string
  // "{name}" is replaced with the season name; marks the primary season.
  currentTemplate: string
}

// Switches the wiki to another season, staying on the same page when that season has it.
export function SeasonSelect({
  seasons,
  routes,
  label,
  currentTemplate,
}: Props) {
  const pathname = usePathname()
  const router = useRouter()
  const selected = useWikiSeason(seasons)
  if (!selected) return null

  const [, locale, , segment, ...rest] = pathname.split('/')
  const onChange = (slug: string) => {
    const samePage = `/wiki/${slug}/${rest.join('/')}`
    const target =
      segment === selected.slug && rest.length > 0 && routes.includes(samePage)
        ? samePage
        : `/wiki/${slug}`
    router.push(`/${locale}${target}`)
  }

  const nameOf = (season: WikiSeason) =>
    season.isPrimary
      ? currentTemplate.replace('{name}', season.name)
      : season.name

  return (
    <Select
      title={label}
      value={selected.slug}
      selectedOption={selected.name}
      onChange={onChange}
      options={seasons.map((season) => ({
        id: season.slug,
        name: nameOf(season),
      }))}
    />
  )
}
