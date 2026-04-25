'use client'

import { setLanguage } from '@/app/actions'
import { DropdownMenuItem } from '@/components/ui/dropdown-menu'
import { Locale } from '@/lib/locale'
import { errorToast } from '@/lib/toasts'
import { CheckIcon } from 'lucide-react'
import { usePathname, useRouter } from 'next/navigation'
import React from 'react'

type Props = {
  value: Locale
  label: string
  currentLocale: string
}

export default function LanguageDropdownMenuItem({
  value,
  label,
  currentLocale,
}: Props) {
  const [isLanguageChanging, startTransition] = React.useTransition()
  const pathname = usePathname()
  const router = useRouter()

  const handleClick = () => {
    if (isLanguageChanging) return
    startTransition(() =>
      setLanguage(value, pathname)
        .then(({ redirectTo }) => {
          console.log('redirectTo', redirectTo)
          router.push(redirectTo)
        })
        .catch((error) => {
          errorToast('Failed to change language', error)
        }),
    )
  }

  return (
    <DropdownMenuItem
      key={value}
      onClick={handleClick}
      disabled={isLanguageChanging}
    >
      {value === currentLocale ? (
        <CheckIcon className="size-4" />
      ) : (
        <div className="size-4" />
      )}
      {label}
    </DropdownMenuItem>
  )
}
