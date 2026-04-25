'use client'

import React from 'react'
import { useLandingNavigation } from './landing-navigation-ctx'

type SidebarSectionProps = React.PropsWithChildren<
  React.ComponentProps<'section'>
>

export default function SidebarSection({
  children,
  ...props
}: SidebarSectionProps) {
  const { activeSection, setActiveSection } = useLandingNavigation()
  const sectionRef = React.useRef<HTMLDivElement>(null)

  React.useEffect(() => {
    const handleIntersection = function (entries: IntersectionObserverEntry[]) {
      entries.forEach((entry) => {
        if (entry.target.id !== activeSection && entry.isIntersecting) {
          setActiveSection(entry.target.id)
        }
      })
    }
    const observer = new IntersectionObserver(handleIntersection, {
      rootMargin: '0px 0px -30% 0px',
      threshold: [0.5],
    })
    if (sectionRef.current) observer.observe(sectionRef.current)
    return () => observer.disconnect()
  }, [activeSection, setActiveSection, sectionRef])

  return (
    <section ref={sectionRef} {...props}>
      {children}
    </section>
  )
}
