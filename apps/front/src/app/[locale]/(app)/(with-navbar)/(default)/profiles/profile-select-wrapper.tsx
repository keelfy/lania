'use client'

import { Button } from '@/components/ui/button'
import PlayerFace from '@/components/ui/player-face'
import { Profile } from '@/models/profile'
import { CheckIcon, LoaderCircleIcon } from 'lucide-react'
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
    <div
      role="group"
      aria-label={t('chooseProfile')}
      aria-busy={pending}
      className="flex w-full flex-wrap gap-2 lg:w-auto"
    >
      {profiles.map((profile) => {
        const selected = selectedProfile?.id === profile.id
        return (
          <Button
            key={profile.id}
            variant={selected ? 'secondary' : 'outline'}
            aria-pressed={selected}
            disabled={pending}
            className="min-w-0 flex-1 gap-2 lg:flex-none"
            onClick={() => {
              if (selected) return
              const query = new URLSearchParams(searchParams)
              query.set('id', profile.id)
              startTransition(() => router.push(`${pathname}?${query}`))
            }}
          >
            <PlayerFace player={profile} className="size-5 shrink-0" />
            <span className="truncate">{profile.username}</span>
            <span className="size-4 shrink-0">
              {pending ? (
                <LoaderCircleIcon className="size-4 motion-safe:animate-spin" />
              ) : selected ? (
                <CheckIcon className="size-4" />
              ) : null}
            </span>
          </Button>
        )
      })}
    </div>
  )
}
