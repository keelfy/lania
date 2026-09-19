'use client'

import { Drawer, DrawerContent, DrawerTrigger } from '@/components/ui/drawer'
import PlayerCard from '@/components/ui/player-card'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'
import { useMediaQuery } from '@/lib/use-media-query'
import { PublicProfile } from '@/models/profile'
import { ArrowUpRightIcon } from 'lucide-react'
import { useLocale, useTranslations } from 'next-intl'
import Link from 'next/link'
import React from 'react'

type Props = React.ComponentProps<'button'> & {
  profile: PublicProfile
}

export default function CommunityPlayerItem({
  children,
  profile,
  ...props
}: React.PropsWithChildren<Props>) {
  const isDesktop = useMediaQuery('(min-width: 1024px)')
  const locale = useLocale()
  const t = useTranslations('community.profile')
  const profileLink = (
    <Link
      href={`/${locale}/community/${encodeURIComponent(profile.username)}`}
      className="text-muted-foreground hover:text-foreground flex items-center justify-center gap-1 pb-3 text-sm transition-colors"
    >
      {t('open')}
      <ArrowUpRightIcon className="size-4" />
    </Link>
  )
  if (isDesktop) {
    return (
      <Popover>
        <PopoverTrigger {...props} asChild>
          {children}
        </PopoverTrigger>
        <PopoverContent className="max-w-sm min-w-max p-0">
          {profile?.id && (
            <PlayerCard
              profileId={profile.id}
              nameCosmetics={profile.cosmetics.name}
              className="border-none"
            />
          )}
          {profileLink}
        </PopoverContent>
      </Popover>
    )
  }
  return (
    <Drawer>
      <DrawerTrigger {...props} asChild>
        {children}
      </DrawerTrigger>
      <DrawerContent>
        {profile?.id && (
          <PlayerCard
            profileId={profile.id}
            nameCosmetics={profile.cosmetics.name}
            className="border-none bg-transparent p-0"
          />
        )}
        {profileLink}
      </DrawerContent>
    </Drawer>
  )
}
