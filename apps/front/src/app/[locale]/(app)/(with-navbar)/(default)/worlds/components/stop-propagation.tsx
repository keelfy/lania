'use client'

import React from 'react'

// The maps grid sits inside the same Card that copies the address on
// click, so a click or Enter/Space on a map Link would otherwise bubble
// up and trigger the copy too. This stops it at the boundary.
export default function StopPropagation({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    // Not interactive itself, just a propagation firewall around
    // children that are, so no role/tabIndex belongs here.
    // eslint-disable-next-line jsx-a11y/no-static-element-interactions
    <div
      onClick={(event) => event.stopPropagation()}
      onKeyDown={(event) => event.stopPropagation()}
    >
      {children}
    </div>
  )
}
