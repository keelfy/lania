'use client'

import { NavigationMenuLink } from '@/components/ui/navigation-menu'
import { usePathname } from '@/i18n/navigation'
import { LayoutGroup } from 'framer-motion'
import Link from 'next/link'
import React from 'react'
import NavbarHighlighter from './navbar-highlighter'

const NavigationContext = React.createContext<{
  pathname: string
  highlightedHref: string | null
  setHoveredHref: (href: string | null) => void
  setFocusedHref: (href: string | null) => void
} | null>(null)

export function NavbarNavigation({ children }: React.PropsWithChildren) {
  const pathname = usePathname()
  const [hoveredHref, setHoveredHref] = React.useState<string | null>(null)
  const [focusedHref, setFocusedHref] = React.useState<string | null>(null)
  const id = React.useId()

  return (
    <NavigationContext.Provider
      value={{
        pathname,
        highlightedHref: hoveredHref ?? focusedHref,
        setHoveredHref,
        setFocusedHref,
      }}
    >
      <LayoutGroup id={id}>
        <div onPointerLeave={() => setHoveredHref(null)}>{children}</div>
      </LayoutGroup>
    </NavigationContext.Provider>
  )
}

type Props = {
  href: string
  routeHref: string
  className?: string
  disabled?: boolean
}

export default function NavItemLink({
  href,
  routeHref,
  className,
  disabled,
  children,
}: React.PropsWithChildren<Props>) {
  const navigation = React.useContext(NavigationContext)
  if (!navigation) throw new Error('NavItemLink requires NavbarNavigation')
  const isActive =
    navigation.pathname === routeHref ||
    navigation.pathname.startsWith(`${routeHref}/`)
  const isHighlighted = navigation.highlightedHref
    ? navigation.highlightedHref === routeHref
    : isActive

  return (
    <NavigationMenuLink asChild active={isActive} className={className}>
      <Link
        href={href}
        aria-disabled={disabled || undefined}
        tabIndex={disabled ? -1 : undefined}
        data-highlighted={isHighlighted || undefined}
        onPointerEnter={(event) => {
          if (!disabled && event.pointerType !== 'touch')
            navigation.setHoveredHref(routeHref)
        }}
        onFocus={() => {
          navigation.setHoveredHref(null)
          navigation.setFocusedHref(routeHref)
        }}
        onBlur={() => navigation.setFocusedHref(null)}
        onClick={(event) => {
          if (disabled) event.preventDefault()
        }}
      >
        {isHighlighted ? <NavbarHighlighter /> : null}
        <span className="navbar-link-content">{children}</span>
        {isActive ? <span className="navbar-active-line" aria-hidden /> : null}
      </Link>
    </NavigationMenuLink>
  )
}
