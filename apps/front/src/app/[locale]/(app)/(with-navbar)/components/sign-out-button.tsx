'use client'

import React from 'react'

import { DropdownMenuItem } from '@/components/ui/dropdown-menu'
import LoadingSpinner from '@/components/ui/loading-spinner'
import ory from '@/lib/ory'
import { LogOutIcon } from 'lucide-react'
import { useTranslations } from 'next-intl'
import { toast } from 'sonner'

export default function SignOutDropdownMenuItem() {
  const [isPending, startTransition] = React.useTransition()
  const t = useTranslations()

  const onSignOut = () =>
    startTransition(async () => {
      try {
        const flow = await ory.createBrowserLogoutFlow()
        await ory.updateLogoutFlow({
          token: flow.logout_token,
          returnTo: window.location.href,
        })
        window.location.reload()
      } catch (e) {
        toast.error('Failed to sign out', {
          description:
            e instanceof Error ? e.message : 'Please try again later.',
        })
      }
    })

  return (
    <DropdownMenuItem onClick={onSignOut}>
      {isPending ? (
        <LoadingSpinner />
      ) : (
        <LogOutIcon className="text-destructive size-4" />
      )}
      {t('navbar.userDropdown.signOut')}
    </DropdownMenuItem>
  )
}
