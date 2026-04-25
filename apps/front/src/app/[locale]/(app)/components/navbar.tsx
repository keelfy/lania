import DeerIcon from '@/components/icons/DeerIcon'
import { Button } from '@/components/ui/button'
import {
  NavigationMenu,
  NavigationMenuItem,
  NavigationMenuLink,
  NavigationMenuList,
  navigationMenuTriggerStyle,
} from '@/components/ui/navigation-menu'
import { Separator } from '@/components/ui/separator'
import { Sheet, SheetTrigger } from '@/components/ui/sheet'
import { Currency, CURRENCY_COOKIE, DEFAULT_CURRENCY } from '@/lib/currency'
import { isCurrentSessionActive } from '@/lib/get-current-session'
import { Locale } from '@/lib/locale'
import { cn } from '@/lib/utils'
import {
  BellIcon,
  BookIcon,
  CalendarIcon,
  MapIcon,
  MenuIcon,
  ShoppingBagIcon,
  UsersIcon,
} from 'lucide-react'
import { getTranslations } from 'next-intl/server'
import dynamic from 'next/dynamic'
import { cookies } from 'next/headers'
import Link from 'next/link'
import React from 'react'
import SignInButton from '../(with-navbar)/components/sign-in-button'
import LanguageDropdownMenu from './language-dropdown-menu'
import NavbarHighlighter from './navbar-highlighter'
import NotificationsMenu from './notifications-button'
import ShoppingBasketButton from './shopping-basket-button'
import UserDropdownMenu from './user-dropdown'

const DynamicMenuSheetContent = dynamic(
  () => import('./menu/menu-sheet-content'),
)

type NavbarItem = {
  labelKey: string
  href: string
  icon: React.ElementType
  disabled?: boolean
}

const navItems: NavbarItem[] = [
  {
    labelKey: 'worlds',
    href: '/worlds',
    icon: MapIcon,
  },
  {
    labelKey: 'wiki',
    href: '/wiki',
    icon: BookIcon,
  },
  {
    labelKey: 'products',
    href: '/products',
    icon: ShoppingBagIcon,
  },
  {
    labelKey: 'community',
    href: '/community',
    icon: UsersIcon,
  },
  {
    labelKey: 'seasons',
    href: '/seasons',
    icon: CalendarIcon,
  },
]

type Props = {
  currentLocale: Locale
}

export default async function Navbar({
  className,
  currentLocale,
  ...props
}: React.ComponentProps<'header'> & Props) {
  const t = await getTranslations({ locale: currentLocale })
  const isSessionActive = await isCurrentSessionActive()
  const currency =
    ((await cookies()).get(CURRENCY_COOKIE)?.value as Currency) ??
    DEFAULT_CURRENCY

  return (
    <header
      className={cn(
        'bg-background/70 container mx-auto flex h-16 w-full max-w-5xl items-center justify-between gap-3 px-4 py-4 backdrop-blur lg:px-0',
        className,
      )}
      {...props}
    >
      <div className="flex h-full items-center gap-3">
        <Link href="/" className="group flex flex-nowrap items-center gap-2">
          <DeerIcon className="size-8 transition-opacity duration-300 group-hover:opacity-80" />
          <h1 className="bg-gradient-to-r from-white to-teal-400/90 bg-clip-text text-lg font-bold tracking-widest text-transparent uppercase transition-opacity duration-300 group-hover:opacity-80">
            Lania
          </h1>
        </Link>
        <Separator orientation="vertical" className="ml-4 hidden lg:block" />
        <div className="hidden lg:block">
          <NavigationMenu>
            <NavigationMenuList>
              {navItems.map((item) => (
                <NavigationMenuItem key={item.href}>
                  <Link href={`/${currentLocale}${item.href}`} legacyBehavior>
                    <NavigationMenuLink
                      className={cn(
                        navigationMenuTriggerStyle(),
                        'hover:border-accent cursor-pointer border-1 border-transparent bg-transparent transition-all hover:bg-transparent',
                        item.disabled && 'pointer-events-none opacity-70',
                      )}
                    >
                      <div
                        className={cn(
                          'relative z-10 flex items-center gap-2',
                          // 'px-4 py-2 bg-gradient-to-r from-white to-teal-400/90 bg-[length:200%_100%] bg-clip-text bg-[position:0%_50%] text-transparent transition-all duration-300 hover:bg-[position:70%_50%]',
                        )}
                      >
                        <item.icon className="size-4" />
                        <p>{t(`navbar.items.${item.labelKey}`)}</p>
                      </div>
                    </NavigationMenuLink>
                  </Link>
                  <NavbarHighlighter href={item.href} />
                </NavigationMenuItem>
              ))}
            </NavigationMenuList>
          </NavigationMenu>
        </div>
      </div>
      <div className="absolute right-4 flex items-center gap-2 lg:right-0">
        <NotificationsMenu>
          <Button variant="ghost" size="icon">
            <span className="sr-only">{t('notifications.title')}</span>
            <BellIcon className="size-4" />
          </Button>
        </NotificationsMenu>
        <ShoppingBasketButton />
        <Sheet>
          <SheetTrigger asChild>
            <Button variant="ghost" size="icon" className="lg:hidden">
              <MenuIcon className="size-8" />
            </Button>
          </SheetTrigger>
          <DynamicMenuSheetContent
            sessionActive={isSessionActive}
            locale={currentLocale as Locale}
            currency={currency}
          />
        </Sheet>
        <div className="relative hidden items-center gap-6 lg:flex">
          {isSessionActive ? <UserDropdownMenu /> : <SignInButton />}
          <LanguageDropdownMenu
            currentLocale={currentLocale}
            className="absolute right-0 translate-x-[calc(100%+1rem)]"
          />
        </div>
      </div>
    </header>
  )
}
