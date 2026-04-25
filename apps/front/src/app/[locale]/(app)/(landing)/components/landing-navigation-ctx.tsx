'use client'

import { createContext, useContext, useState } from 'react'

export const LandingNavigationContext = createContext<{
  activeSection: string
  setActiveSection: (section: string) => void
}>({
  activeSection: 'section-1',
  setActiveSection: () => {},
})

export function LandingNavigationProvider({
  children,
}: React.PropsWithChildren) {
  const [activeSection, setActiveSection] = useState('section-1')

  return (
    <LandingNavigationContext.Provider
      value={{ activeSection, setActiveSection }}
    >
      {children}
    </LandingNavigationContext.Provider>
  )
}

export function useLandingNavigation() {
  return useContext(LandingNavigationContext)
}
