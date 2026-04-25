'use client'

import { Button } from '@/components/ui/button'
import LoadingSpinner from '@/components/ui/loading-spinner'
import ory from '@/lib/ory'
import { errorToast } from '@/lib/toasts'
import { LogOutIcon } from 'lucide-react'
import { useTranslations } from 'next-intl'
import React from 'react'

export default function MenuSheetSignOutButton() {
  const [isPending, startTransition] = React.useTransition()
  const t = useTranslations('navbar')

  const onSignOut = () =>
    startTransition(async () => {
      try {
        const flow = await ory.createBrowserLogoutFlow()
        await ory.updateLogoutFlow({
          token: flow.logout_token,
          returnTo: window.location.href,
        })
        window.location.reload()
      } catch (error) {
        errorToast(t('userDropdown.signOutFailed'), error)
      }
    })

  return (
    <Button variant="secondary" onClick={onSignOut}>
      {isPending ? <LoadingSpinner /> : <LogOutIcon />}
      {t('userDropdown.signOut')}
    </Button>
  )
}
