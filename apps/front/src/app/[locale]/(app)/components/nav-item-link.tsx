'use client'

import { NavigationMenuLink } from '@/components/ui/navigation-menu'
import Link from 'next/link'
import React from 'react'

type Props = {
  href: string
  className?: string
}

export default function NavItemLink({
  href,
  className,
  children,
}: React.PropsWithChildren<Props>) {
  return (
    <NavigationMenuLink asChild className={className}>
      <Link href={href}>{children}</Link>
    </NavigationMenuLink>
  )
}
