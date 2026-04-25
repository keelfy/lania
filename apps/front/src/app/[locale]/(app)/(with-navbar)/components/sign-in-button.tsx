'use client'

import React from 'react'

import { Button } from '@/components/ui/button'
import { LogInIcon } from 'lucide-react'
import { useTranslations } from 'next-intl'
import Link from 'next/link'
import { usePathname, useSearchParams } from 'next/navigation'

export default function SignInButton() {
  const pathname = usePathname()
  const searchParams = useSearchParams()
  const t = useTranslations()

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
        <LogInIcon className="size-4" />
        {t('navbar.signIn')}
      </Link>
    </Button>
  )
}
