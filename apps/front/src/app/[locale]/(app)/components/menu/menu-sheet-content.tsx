import CurrencySelect from '@/components/ui/currency-select'
import { Separator } from '@/components/ui/separator'
import {
  SheetClose,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet'
import { Currency } from '@/lib/currency'
import { Locale } from '@/lib/locale'
import { cn } from '@/lib/utils'
import {
  Book,
  CalendarIcon,
  GamepadIcon,
  HandCoinsIcon,
  MapIcon,
  MessageCircleIcon,
  SettingsIcon,
  ShoppingBagIcon,
  UsersIcon,
} from 'lucide-react'
import { useTranslations } from 'next-intl'
import Link from 'next/link'
import MenuSheetLanguageSelect from './menu-sheet-language-select'
import MenuSheetSignInButton from './menu-sheet-sign-in-btn'
import MenuSheetSignOutButton from './menu-sheet-sign-out-btn'

const menuItems = [
  [
    {
      label: 'items.worlds',
      href: '/worlds',
      icon: MapIcon,
    },
    {
      label: 'items.wiki',
      href: '/wiki',
      icon: Book,
    },
    {
      label: 'items.products',
      href: '/products',
      icon: ShoppingBagIcon,
    },
    {
      label: 'items.community',
      href: '/community',
      icon: UsersIcon,
    },
    {
      label: 'items.seasons',
      href: '/seasons',
      icon: CalendarIcon,
    },
  ],
  [
    {
      label: 'userDropdown.profiles',
      href: '/profiles',
      icon: GamepadIcon,
    },
    {
      label: 'userDropdown.orders',
      href: '/orders',
      icon: HandCoinsIcon,
    },
    {
      label: 'userDropdown.settings',
      href: '/settings',
      icon: SettingsIcon,
      disabled: true,
    },
    {
      label: 'userDropdown.support',
      href: process.env.NEXT_PUBLIC_SUPPORT_URL ?? '#',
      icon: MessageCircleIcon,
      disabled: true,
    },
  ],
]

type Props = {
  sessionActive: boolean
  locale: Locale
  currency: Currency
}

export default function MenuSheetContent({
  sessionActive,
  locale,
  currency,
}: Props) {
  const t = useTranslations('navbar')
  return (
    <SheetContent className="flex flex-col gap-6">
      <SheetHeader>
        <SheetTitle>{t('title')}</SheetTitle>
        <SheetDescription className="sr-only">{t('title')}</SheetDescription>
      </SheetHeader>
      <div className="flex h-full flex-col justify-between gap-6 px-4">
        <div className="flex flex-1 flex-col gap-6">
          {menuItems.map((items, index) => (
            <div className="flex flex-col gap-4" key={index}>
              {items.map((item) => (
                <SheetClose key={item.label}>
                  <Link
                    href={item.href}
                    className={cn(
                      'flex items-center gap-2 text-lg font-medium',
                      item.disabled &&
                        'text-muted-foreground pointer-events-none opacity-70',
                    )}
                  >
                    <item.icon className="size-5" />
                    {t(item.label)}
                  </Link>
                </SheetClose>
              ))}
              {index < menuItems.length - 1 && <Separator className="mt-4" />}
            </div>
          ))}
        </div>
      </div>
      <SheetFooter>
        <div className="my-4 flex flex-col gap-2">
          <MenuSheetLanguageSelect currentLocale={locale} className="w-full" />
          <CurrencySelect currency={currency} className="w-full" />
        </div>
        {!sessionActive ? (
          <MenuSheetSignInButton />
        ) : (
          <MenuSheetSignOutButton />
        )}
      </SheetFooter>
    </SheetContent>
  )
}
