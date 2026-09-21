import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
  HandCoinsIcon,
  MessageCircleIcon,
  SettingsIcon,
  ShieldCheckIcon,
  UserIcon,
} from 'lucide-react'
import { useLocale, useTranslations } from 'next-intl'
import Link from 'next/link'
import SignOutDropdownMenuItem from '../(with-navbar)/components/sign-out-button'

const menuItems = [
  {
    label: 'profiles',
    href: '/profiles',
    icon: UserIcon,
  },
  {
    label: 'orders',
    href: '/orders',
    icon: HandCoinsIcon,
  },
  {
    label: 'settings',
    href: '/settings',
    icon: SettingsIcon,
  },
  {
    label: 'support',
    href: process.env.NEXT_PUBLIC_SUPPORT_URL ?? '#',
    icon: MessageCircleIcon,
    disabled: true,
  },
]

type Props = {
  isAdmin?: boolean
}

export default function UserDropdownMenu({ isAdmin = false }: Props) {
  const t = useTranslations('navbar.userDropdown')
  const locale = useLocale()
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button size="icon">
          <UserIcon className="size-6" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        <DropdownMenuGroup>
          {menuItems.map((item) => (
            <DropdownMenuItem asChild key={item.label} disabled={item.disabled}>
              <Link href={item.href}>
                <item.icon className="size-4" />
                {t(item.label)}
              </Link>
            </DropdownMenuItem>
          ))}
          {isAdmin && (
            <DropdownMenuItem asChild>
              <Link href={`/${locale}/admin/users`}>
                <ShieldCheckIcon className="size-4" />
                {t('admin')}
              </Link>
            </DropdownMenuItem>
          )}
        </DropdownMenuGroup>
        <DropdownMenuSeparator />
        <SignOutDropdownMenuItem />
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
