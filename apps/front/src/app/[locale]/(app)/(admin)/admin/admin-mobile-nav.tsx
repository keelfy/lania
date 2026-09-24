'use client'

import { Button } from '@/components/ui/button'
import {
  Sheet,
  SheetClose,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from '@/components/ui/sheet'
import { ArrowLeftIcon, MenuIcon } from 'lucide-react'
import { useTranslations } from 'next-intl'
import Link from 'next/link'
import React from 'react'
import AdminSidebarNav from './admin-sidebar-nav'

type Props = {
  locale: string
}

export default function AdminMobileNav({ locale }: Props) {
  const t = useTranslations('admin.sidebar')
  const [open, setOpen] = React.useState(false)

  return (
    <Sheet open={open} onOpenChange={setOpen}>
      <SheetTrigger asChild>
        <Button variant="ghost" size="icon">
          <span className="sr-only">{t('badge')}</span>
          <MenuIcon className="size-5" />
        </Button>
      </SheetTrigger>
      <SheetContent side="left" className="w-64 px-0">
        <SheetHeader>
          <SheetTitle>{t('badge')}</SheetTitle>
          <SheetDescription className="sr-only">{t('badge')}</SheetDescription>
        </SheetHeader>
        <AdminSidebarNav
          locale={locale}
          onNavigate={() => setOpen(false)}
          className="flex-1 px-3"
        />
        <SheetFooter>
          <SheetClose asChild>
            <Link
              href={`/${locale}`}
              className="text-muted-foreground hover:text-foreground flex items-center gap-2 text-sm transition-colors"
            >
              <ArrowLeftIcon className="size-4" />
              {t('backToSite')}
            </Link>
          </SheetClose>
        </SheetFooter>
      </SheetContent>
    </Sheet>
  )
}
