import {
  CalendarRangeIcon,
  IdCardIcon,
  LucideIcon,
  MegaphoneIcon,
  PackageIcon,
  PaletteIcon,
  ReceiptIcon,
  UsersIcon,
} from 'lucide-react'

// Shared between the desktop sidebar and the mobile sheet so both stay in sync.
export const ADMIN_NAV_ITEMS: readonly {
  href: string
  labelKey: string
  icon: LucideIcon
  children?: readonly { href: string; labelKey: string }[]
}[] = [
  { href: '/admin/users', labelKey: 'users', icon: UsersIcon },
  { href: '/admin/profiles', labelKey: 'profiles', icon: IdCardIcon },
  { href: '/admin/seasons', labelKey: 'seasons', icon: CalendarRangeIcon },
  {
    href: '/admin/cosmetics',
    labelKey: 'cosmetics',
    icon: PaletteIcon,
    children: [
      { href: '/admin/cosmetics/colors', labelKey: 'nameColors' },
      { href: '/admin/cosmetics/prefixes', labelKey: 'namePrefixes' },
    ],
  },
  { href: '/admin/products', labelKey: 'products', icon: PackageIcon },
  { href: '/admin/orders', labelKey: 'orders', icon: ReceiptIcon },
  {
    href: '/admin/announcements',
    labelKey: 'announcements',
    icon: MegaphoneIcon,
  },
] as const
