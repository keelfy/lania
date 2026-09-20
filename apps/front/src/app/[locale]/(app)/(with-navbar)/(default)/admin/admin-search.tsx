'use client'

import { Input } from '@/components/ui/input'
import { useDebouncedState } from '@/lib/use-debounced-state'
import { useRouter } from 'next/navigation'
import React from 'react'

type Props = {
  // Page the search reloads, for example /en/admin/users.
  path: string
  defaultValue?: string
  placeholder: string
  maxLength: number
}

export default function AdminSearch({
  path,
  defaultValue,
  placeholder,
  maxLength,
}: Props) {
  const router = useRouter()
  const [value, setValue] = React.useState(defaultValue ?? '')
  const search = useDebouncedState(value.trim(), 300)

  React.useEffect(() => {
    if (search === (defaultValue ?? '')) return
    const params = new URLSearchParams()
    if (search) params.set('q', search)
    router.replace(`${path}?${params.toString()}`)
  }, [search, defaultValue, path, router])

  return (
    <Input
      type="search"
      value={value}
      onChange={(e) => setValue(e.target.value)}
      maxLength={maxLength}
      placeholder={placeholder}
      aria-label={placeholder}
      className="w-full sm:w-72"
    />
  )
}
