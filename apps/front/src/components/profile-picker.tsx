'use client'

import { Input } from '@/components/ui/input'
import { getAdminProfiles } from '@/lib/api-endpoints'
import { clientApiFetcher } from '@/lib/client'
import { useDebouncedState } from '@/lib/use-debounced-state'
import { AdminProfile } from '@/models/admin'
import React from 'react'

// The longest Minecraft username.
const MAX_SEARCH_LENGTH = 16

type Props = {
  onSelect: (profile: AdminProfile) => void
  placeholder: string
  noResultsLabel: string
  // Profiles that must not show up in the results, e.g. already picked or the profile the search is opened from.
  excludeIds?: string[]
}

// A debounced profile search: type a username, click a result. Shared by the profile merge dialog and the
// season screenshot author picker so the fetch/debounce logic exists in one place.
export default function ProfilePicker({
  onSelect,
  placeholder,
  noResultsLabel,
  excludeIds = [],
}: Props) {
  const [query, setQuery] = React.useState('')
  const debouncedQuery = useDebouncedState(query.trim(), 300)
  const [results, setResults] = React.useState<AdminProfile[]>([])
  const excludeKey = excludeIds.join(',')

  // Stale results are hidden through this, not cleared through setState in the effect below,
  // so the effect never calls setState synchronously on its early-return path.
  const visibleResults = debouncedQuery ? results : []

  React.useEffect(() => {
    if (!debouncedQuery) return

    let cancelled = false
    getAdminProfiles(clientApiFetcher, 0, debouncedQuery, 6)
      .then((page) => {
        if (!cancelled) {
          setResults(
            page.content.filter((profile) => !excludeIds.includes(profile.id)),
          )
        }
      })
      .catch(() => {
        if (!cancelled) setResults([])
      })
    return () => {
      cancelled = true
    }
    // excludeKey stands in for excludeIds: only its content, not its identity, should retrigger the search.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [debouncedQuery, excludeKey])

  return (
    <div className="flex flex-col gap-2">
      <Input
        value={query}
        onChange={(e) => setQuery(e.target.value)}
        maxLength={MAX_SEARCH_LENGTH}
        placeholder={placeholder}
        aria-label={placeholder}
      />
      {visibleResults.length > 0 && (
        <ul className="border-border divide-border divide-y rounded-md border">
          {visibleResults.map((profile) => (
            <li key={profile.id}>
              <button
                type="button"
                className="hover:bg-accent w-full px-3 py-2 text-left text-sm"
                onClick={() => {
                  onSelect(profile)
                  setQuery('')
                  setResults([])
                }}
              >
                {profile.username}
              </button>
            </li>
          ))}
        </ul>
      )}
      {debouncedQuery && visibleResults.length === 0 && (
        <p className="text-muted-foreground text-sm">{noResultsLabel}</p>
      )}
    </div>
  )
}
