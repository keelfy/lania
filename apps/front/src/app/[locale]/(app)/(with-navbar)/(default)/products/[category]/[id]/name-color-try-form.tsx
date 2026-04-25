'use client'

import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import McUsername from '@/components/ui/mc-username'
import { PencilIcon } from 'lucide-react'
import React from 'react'

type Props = {
  defaultUsername: string
  colors: string[]
}

export default function NameColorTryForm({ defaultUsername, colors }: Props) {
  const [username, setUsername] = React.useState(defaultUsername)

  return (
    <div className="flex flex-col gap-4">
      <div className="bg-card relative flex items-center justify-center rounded-lg px-6 py-10">
        <McUsername
          username={username.length > 0 ? username : defaultUsername}
          colors={colors}
          className="scroll-m-20 text-center text-4xl font-bold"
        />
      </div>
      <div className="flex items-center gap-0">
        <Label className="text-md h-9 rounded-l-md border px-3 text-nowrap">
          <PencilIcon className="size-4" />
          Введите имя:
        </Label>
        <Input
          placeholder={defaultUsername}
          value={username}
          onChange={(e) => setUsername(e.target.value)}
          className="rounded-l-none"
        />
      </div>
    </div>
  )
}
