'use client'

import { Profile } from '@/models/profile'
import React from 'react'
import McUsername from './mc-username'
import PlayerFace from './player-face'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from './select'

type Props = React.ComponentProps<typeof SelectTrigger> & {
  profiles: Profile[]
  selectedProfileId: string | undefined
  onSelectProfileId: (id: string) => void
  placeholder?: string
}

export default function ProfileSelect({
  profiles,
  selectedProfileId,
  onSelectProfileId,
  placeholder,
  ...props
}: Props) {
  const items = React.useMemo(
    () =>
      profiles.map((player) => (
        <SelectItem key={player.id} value={player.id}>
          <div className="flex items-center gap-2">
            <PlayerFace player={player} className="size-4" />
            <McUsername
              username={player.username}
              colors={player.cosmetics.name.colors.colors}
              className="text-md"
            />
          </div>
        </SelectItem>
      )),
    [profiles],
  )
  return (
    <Select value={selectedProfileId} onValueChange={onSelectProfileId}>
      <SelectTrigger {...props}>
        <SelectValue placeholder={placeholder} />
      </SelectTrigger>
      <SelectContent>{items}</SelectContent>
    </Select>
  )
}
