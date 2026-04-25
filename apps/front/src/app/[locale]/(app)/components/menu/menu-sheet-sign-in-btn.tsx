'use client'

import { Button } from '@/components/ui/button'
import { LogInIcon } from 'lucide-react'
import { useTranslations } from 'next-intl'
import Link from 'next/link'
import { usePathname, useSearchParams } from 'next/navigation'
import React from 'react'

export default function MenuSheetSignInButton() {
  const pathname = usePathname()
  const searchParams = useSearchParams()
  const t = useTranslations('navbar')
  const returnTo = React.useMemo(() => {
    return `${pathname}?${searchParams.toString()}`
  }, [pathname, searchParams])

  return (
    <Button asChild>
      <Link
        href={{
          pathname: `${process.env.NEXT_PUBLIC_ORY_SDK_URL}/self-service/login/browser`,
          query: {
            return_to: returnTo,
          },
        }}
      >
        <LogInIcon />
        {t('signIn')}
      </Link>
    </Button>
  )
}
