'use client'

import {
  Drawer,
  DrawerContent,
  DrawerHeader,
  DrawerTitle,
  DrawerTrigger,
} from '@/components/ui/drawer'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'
import { useMediaQuery } from '@/lib/use-media-query'
import { useTranslations } from 'next-intl'
import React from 'react'

export default function NotificationsMenu({
  children,
}: React.PropsWithChildren) {
  const isDesktop = useMediaQuery('(min-width: 1024px)')
  const t = useTranslations('navbar.notifications')

  if (isDesktop) {
    return (
      <Popover>
        <PopoverTrigger asChild>{children}</PopoverTrigger>
        <PopoverContent className="flex min-h-32 items-center justify-center">
          <p className="text-muted-foreground h-full text-center text-sm">
            {t('noNotifications')}
          </p>
        </PopoverContent>
      </Popover>
    )
  }

  return (
    <Drawer>
      <DrawerTrigger asChild>{children}</DrawerTrigger>
      <DrawerContent>
        <DrawerHeader>
          <DrawerTitle>{t('title')}</DrawerTitle>
        </DrawerHeader>
        <div className="min-h-32">
          <p className="text-muted-foreground h-full text-center text-sm">
            {t('noNotifications')}
          </p>
        </div>
      </DrawerContent>
    </Drawer>
  )
}
