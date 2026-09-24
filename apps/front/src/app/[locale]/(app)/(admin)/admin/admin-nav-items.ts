import {
  CalendarRangeIcon,
  IdCardIcon,
  PackageIcon,
  PaletteIcon,
  UsersIcon,
} from 'lucide-react'

// Shared between the desktop sidebar and the mobile sheet so both stay in sync.
export const ADMIN_NAV_ITEMS = [
  { href: '/admin/users', labelKey: 'users', icon: UsersIcon },
  { href: '/admin/profiles', labelKey: 'profiles', icon: IdCardIcon },
  { href: '/admin/seasons', labelKey: 'seasons', icon: CalendarRangeIcon },
  { href: '/admin/cosmetics', labelKey: 'cosmetics', icon: PaletteIcon },
  { href: '/admin/products', labelKey: 'products', icon: PackageIcon },
] as const
