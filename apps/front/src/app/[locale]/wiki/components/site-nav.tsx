import {
  BookIcon,
  CalendarIcon,
  MapIcon,
  ShoppingBagIcon,
  UsersIcon,
} from 'lucide-react'
import DeerIcon from '@/components/icons/DeerIcon'
import Link from 'next/link'
import './site-nav.css'

// Same sections as the site navbar, so the wiki reads as part of the site.
const items = [
  { key: 'worlds', href: '/worlds', icon: MapIcon },
  { key: 'wiki', href: '/wiki', icon: BookIcon },
  { key: 'products', href: '/products', icon: ShoppingBagIcon },
  { key: 'community', href: '/community', icon: UsersIcon },
  { key: 'seasons', href: '/seasons', icon: CalendarIcon },
]

type Props = {
  locale: string
  labels: Record<string, string>
}

// Goes in the navbar's logo slot (with logoLink off), so the links sit next to the logo.
export function SiteNav({ locale, labels }: Props) {
  return (
    <div className="wiki-site-nav">
      <Link href={`/${locale}`} className="wiki-site-logo">
        <DeerIcon style={{ width: '2rem', height: '2rem' }} />
        <span className="wiki-site-logo-text">Lania</span>
      </Link>
      <span className="wiki-site-nav-separator" aria-hidden />
      {items.map((item) => (
        <Link
          key={item.key}
          href={`/${locale}${item.href}`}
          className="wiki-site-nav-item"
          aria-current={item.key === 'wiki' ? 'page' : undefined}
        >
          <item.icon size={16} />
          {labels[item.key]}
        </Link>
      ))}
    </div>
  )
}
