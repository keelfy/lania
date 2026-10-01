'use client'

import PlayerFace from '@/components/ui/player-face'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Profile } from '@/models/profile'
import { LoaderCircleIcon } from 'lucide-react'
import { useTranslations } from 'next-intl'
import { usePathname, useRouter, useSearchParams } from 'next/navigation'
import { useTransition } from 'react'
import { useSelectedProfile } from './use-selected-profile'

type Props = { profiles: Profile[] }

export default function ProfileSelectWrapper({ profiles }: Props) {
  const router = useRouter()
  const pathname = usePathname()
  const searchParams = useSearchParams()
  const selectedProfile = useSelectedProfile(profiles)
  const [pending, startTransition] = useTransition()
  const t = useTranslations('profiles')
  if (!profiles.length) return null

  return (
    <Select
      value={selectedProfile?.id}
      disabled={pending}
      onValueChange={(id) => {
        if (id === selectedProfile?.id) return
        const query = new URLSearchParams(searchParams)
        query.set('id', id)
        startTransition(() => router.push(`${pathname}?${query}`))
      }}
    >
      <SelectTrigger
        className="w-full sm:w-56"
        aria-label={t('chooseProfile')}
        aria-busy={pending}
      >
        <SelectValue
          placeholder={t('chooseProfile')}
          className="min-w-0 flex-1 text-left"
        />
        {pending && (
          <LoaderCircleIcon className="size-4 motion-safe:animate-spin" />
        )}
      </SelectTrigger>
      <SelectContent>
        {profiles.map((profile) => (
          <SelectItem
            key={profile.id}
            value={profile.id}
            textValue={profile.username}
          >
            <PlayerFace player={profile} className="size-5 shrink-0" />
            <span className="truncate">{profile.username}</span>
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  )
}
