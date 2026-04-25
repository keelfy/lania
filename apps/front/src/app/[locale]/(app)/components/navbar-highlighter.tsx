'use client'

import { pathnameStartsWith, usePathname } from '@/i18n/navigation'
import { motion } from 'framer-motion'
import React from 'react'

type Props = {
  href: string
}

export default function NavbarHighlighter({ href }: Props) {
  const pathname = usePathname()
  const isActive = React.useMemo(() => {
    return pathnameStartsWith(pathname, href)
  }, [href, pathname])

  return (
    isActive && (
      <motion.div
        layoutId="navbar-highlight"
        className="bg-accent absolute inset-0 z-5 rounded-md"
        transition={{ type: 'spring', stiffness: 500, damping: 50, mass: 1 }}
      />
    )
  )
}
