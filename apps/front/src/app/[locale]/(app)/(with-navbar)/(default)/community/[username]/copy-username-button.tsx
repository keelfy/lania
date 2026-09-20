'use client'

import { Button } from '@/components/ui/button'
import { errorToast } from '@/lib/toasts'
import { CheckIcon, CopyIcon } from 'lucide-react'
import { useTranslations } from 'next-intl'
import React from 'react'

export default function CopyUsernameButton({ username }: { username: string }) {
  const t = useTranslations('playerCard')
  const [copied, setCopied] = React.useState(false)

  React.useEffect(() => {
    if (!copied) return
    const timeout = setTimeout(() => setCopied(false), 2000)
    return () => clearTimeout(timeout)
  }, [copied])

  return (
    <Button
      variant="ghost"
      size="icon"
      className="size-8"
      onClick={() =>
        navigator.clipboard
          .writeText(username)
          .then(() => setCopied(true))
          .catch((error) => errorToast(t('copyFailed'), error))
      }
      aria-label={t('copyUsername')}
      title={t('copyUsername')}
    >
      {copied ? (
        <CheckIcon className="size-4" />
      ) : (
        <CopyIcon className="size-4" />
      )}
    </Button>
  )
}
