import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { LOCALE_NAMES, LOCALES } from '@/lib/locale'
import { cn } from '@/lib/utils'
import { ChevronDownIcon, LanguagesIcon } from 'lucide-react'
import LanguageDropdownMenuItem from './language-dropdown-menu-item'

const LANGUAGES = LOCALES.map((locale) => ({
  value: locale,
  label: LOCALE_NAMES[locale],
}))

type Props = {
  currentLocale: string
  className?: string
}

export default async function LanguageDropdownMenu({
  currentLocale,
  className,
}: Props) {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" className={cn('uppercase', className)}>
          <LanguagesIcon className="size-4" />
          {currentLocale}
          <ChevronDownIcon className="size-4" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent>
        {LANGUAGES.map((language) => (
          <LanguageDropdownMenuItem
            key={language.value}
            currentLocale={currentLocale}
            {...language}
          />
        ))}
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
