'use client'

import { motion, useReducedMotion } from 'framer-motion'

export default function NavbarHighlighter() {
  const reducedMotion = useReducedMotion()
  return (
    <motion.span
      aria-hidden
      data-navbar-highlight
      layoutId={reducedMotion ? undefined : 'navbar-highlight'}
      className="navbar-highlight"
      transition={{ duration: reducedMotion ? 0 : 0.2, ease: 'easeOut' }}
    />
  )
}
