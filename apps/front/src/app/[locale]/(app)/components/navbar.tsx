import DeerIcon from '@/components/icons/DeerIcon'
import AttentionDot from '@/components/ui/attention-dot'
import { Button } from '@/components/ui/button'
import {
  NavigationMenu,
  NavigationMenuItem,
  NavigationMenuList,
} from '@/components/ui/navigation-menu'
import NavItemLink, { NavbarNavigation } from './nav-item-link'
import { Separator } from '@/components/ui/separator'
import { Sheet, SheetTrigger } from '@/components/ui/sheet'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import { Currency, CURRENCY_COOKIE, DEFAULT_CURRENCY } from '@/lib/currency'
import { isAdminSession } from '@/lib/admin'
import { getCurrentSession } from '@/lib/get-current-session'
import { Locale } from '@/lib/locale'
import { awaitsVerification, getCurrentUserProfiles } from '@/lib/user-profiles'
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
import React, { Suspense } from 'react'
import SignInButton from '../(with-navbar)/components/sign-in-button'
import LanguageDropdownMenu from './language-dropdown-menu'
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
    disabled: false,
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

  return (
    <header
      className={cn(
        'lania-navbar relative mx-auto mt-3 flex min-h-16 w-[calc(100%-1.5rem)] max-w-6xl items-center justify-between gap-2 px-2 sm:gap-3 sm:px-3',
        className,
      )}
      {...props}
    >
      <div className="flex min-w-0 items-center gap-3">
        <Link href={`/${currentLocale}`} className="navbar-brand">
          <span className="navbar-brand-mark">
            <DeerIcon className="size-8" aria-hidden />
          </span>
          <span className="navbar-brand-name">Lania</span>
        </Link>
        <Separator orientation="vertical" className="hidden !h-6 lg:block" />
        <div className="hidden lg:block">
          {/* Path-dependent state stays inside a streamed client boundary. */}
          <Suspense fallback={<NavbarLinksFallback t={t} />}>
            <NavbarNavigation>
              <NavigationMenu viewport={false} aria-label={t('navbar.title')}>
                <NavigationMenuList className="gap-0.5">
                  {navItems.map((item) => (
                    <NavigationMenuItem key={item.href}>
                      <NavItemLink
                        href={`/${currentLocale}${item.href}`}
                        routeHref={item.href}
                        className="navbar-link"
                        disabled={item.disabled}
                      >
                        <item.icon aria-hidden className="size-4" />
                        <span>{t(`navbar.items.${item.labelKey}`)}</span>
                      </NavItemLink>
                    </NavigationMenuItem>
                  ))}
                </NavigationMenuList>
              </NavigationMenu>
            </NavbarNavigation>
          </Suspense>
        </div>
      </div>
      <Suspense fallback={<NavbarActionsFallback />}>
        <NavbarActions currentLocale={currentLocale} />
      </Suspense>
    </header>
  )
}

// The right side of the bar depends on who is signed in and on the currency cookie, so it streams in on its own and the
// rest of the bar is prerendered.
async function NavbarActions({ currentLocale }: Props) {
  const t = await getTranslations({ locale: currentLocale })
  const session = await getCurrentSession()
  const isSessionActive = session?.active === true
  const isAdmin = isAdminSession(session)
  // A profile the owner can verify now; the user menu marks the way to it.
  const profilesNeedAction = (await getCurrentUserProfiles()).some(
    awaitsVerification,
  )
  const currency =
    ((await cookies()).get(CURRENCY_COOKIE)?.value as Currency) ??
    DEFAULT_CURRENCY

  return (
    <div className="navbar-actions flex shrink-0 items-center gap-1.5">
      <div className="navbar-utilities flex items-center gap-0.5">
        <Tooltip delayDuration={500}>
          <NotificationsMenu sessionActive={isSessionActive}>
            <TooltipTrigger asChild>
              <Button variant="ghost" size="icon">
                <span className="sr-only">
                  {t('navbar.notifications.title')}
                </span>
                <BellIcon aria-hidden />
              </Button>
            </TooltipTrigger>
          </NotificationsMenu>
          <TooltipContent side="bottom" sideOffset={8}>
            {t('navbar.notifications.title')}
          </TooltipContent>
        </Tooltip>
        <Tooltip delayDuration={500}>
          <TooltipTrigger asChild>
            <ShoppingBasketButton variant="ghost" className="min-w-11" />
          </TooltipTrigger>
          <TooltipContent side="bottom" sideOffset={8}>
            {t('basket.title')}
          </TooltipContent>
        </Tooltip>
      </div>
      <Sheet>
        <SheetTrigger asChild>
          <Button variant="ghost" size="icon" className="relative lg:hidden">
            <span className="sr-only">{t('navbar.title')}</span>
            <MenuIcon aria-hidden />
            {profilesNeedAction && <AttentionDot />}
          </Button>
        </SheetTrigger>
        <DynamicMenuSheetContent
          sessionActive={isSessionActive}
          isAdmin={isAdmin}
          profilesNeedAction={profilesNeedAction}
          locale={currentLocale as Locale}
          currency={currency}
        />
      </Sheet>
      <Separator orientation="vertical" className="hidden !h-6 lg:block" />
      <div className="hidden items-center gap-1.5 lg:flex">
        {isSessionActive ? (
          <UserDropdownMenu
            isAdmin={isAdmin}
            profilesNeedAction={profilesNeedAction}
          />
        ) : (
          <SignInButton />
        )}
        <LanguageDropdownMenu currentLocale={currentLocale} />
      </div>
    </div>
  )
}

// Keeps the room of the buttons, so the bar does not jump when they arrive.
function NavbarActionsFallback() {
  return (
    <div aria-hidden className="flex shrink-0 items-center gap-1.5">
      <Skeleton className="size-11 rounded-xl" />
      <Skeleton className="size-11 rounded-xl" />
      <Skeleton className="size-11 rounded-xl lg:hidden" />
      <Skeleton className="hidden h-11 w-20 rounded-xl lg:block" />
      <Skeleton className="hidden h-11 w-24 rounded-xl lg:block" />
    </div>
  )
}

function NavbarLinksFallback({
  t,
}: {
  t: Awaited<ReturnType<typeof getTranslations>>
}) {
  return (
    <div aria-hidden className="flex items-center gap-0.5">
      {navItems.map((item) => (
        <span key={item.href} className="navbar-link navbar-link-content">
          <item.icon aria-hidden className="size-4" />
          <span>{t(`navbar.items.${item.labelKey}`)}</span>
        </span>
      ))}
    </div>
  )
}
