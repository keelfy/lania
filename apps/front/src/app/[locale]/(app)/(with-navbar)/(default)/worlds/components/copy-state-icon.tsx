'use client'

import { cn } from '@/lib/utils'
import { CheckIcon, CopyIcon } from 'lucide-react'
import React from 'react'
import { CopiedAddressContext } from './copyable-server-card'

// The copy/check icon pair shared by the hero and compact server cards.
// Reads `copied` from context so the surrounding tree can stay server
// components — only this leaf needs to be interactive.
export default function CopyStateIcon() {
  const copied = React.useContext(CopiedAddressContext)

  return (
    <span className="relative inline-block size-4">
      <CheckIcon
        className={cn(
          'absolute inset-0 transition-all duration-200',
          copied ? 'scale-100 text-teal-500 opacity-100' : 'scale-75 opacity-0',
        )}
      />
      <CopyIcon
        className={cn(
          'absolute inset-0 size-3.5 transition-all duration-200',
          copied
            ? 'scale-75 opacity-0'
            : 'scale-100 opacity-70 group-hover:opacity-100',
        )}
      />
    </span>
  )
}
