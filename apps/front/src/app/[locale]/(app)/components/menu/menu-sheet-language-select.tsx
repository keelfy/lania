'use client'

import { setLanguage } from '@/app/actions'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Locale, LOCALE_NAMES, LOCALES } from '@/lib/locale'
import { errorToast } from '@/lib/toasts'
import { LanguagesIcon } from 'lucide-react'
import { usePathname, useRouter } from 'next/navigation'
import React from 'react'

type Props = React.ComponentProps<typeof SelectTrigger> & {
  currentLocale: Locale
}

export default function MenuSheetLanguageSelect({
  currentLocale,
  ...props
}: Props) {
  const [value, setValue] = React.useState(currentLocale)
  const [isLanguageChanging, startTransition] = React.useTransition()
  const pathname = usePathname()
  const router = useRouter()

  React.useEffect(() => {
    if (value === currentLocale || isLanguageChanging) return
    startTransition(() =>
      setLanguage(value, pathname)
        .then(({ redirectTo }) => {
          router.push(redirectTo)
        })
        .catch((error) => {
          errorToast('Failed to change language', error)
        }),
    )
  }, [value, pathname, router, currentLocale, isLanguageChanging])

  return (
    <Select
      value={value}
      onValueChange={(value) => setValue(value as Locale)}
      disabled={isLanguageChanging}
    >
      <SelectTrigger disabled={isLanguageChanging} {...props}>
        <SelectValue placeholder={LOCALE_NAMES[currentLocale]} />
      </SelectTrigger>
      <SelectContent>
        {LOCALES.map((lang) => (
          <SelectItem key={lang} value={lang}>
            <LanguagesIcon className="size-4" />
            {LOCALE_NAMES[lang]}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  )
}
