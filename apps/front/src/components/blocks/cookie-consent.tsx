'use client'

import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { cn } from '@/lib/utils'
import { Cookie } from 'lucide-react'
import { useTranslations } from 'next-intl'
import * as React from 'react'
import RichText from '../ui/rich-text'

// Define prop types
interface CookieConsentProps extends React.HTMLAttributes<HTMLDivElement> {
  variant?: 'default' | 'small' | 'mini'
  demo?: boolean
  onAcceptCallback?: () => void
  onDeclineCallback?: () => void
  description?: string
  learnMoreHref?: string
}

export const COOKIE_CONSENT_COOKIE = 'cookieConsent'

const CookieConsent = React.forwardRef<HTMLDivElement, CookieConsentProps>(
  (
    {
      variant = 'default',
      demo = false,
      onAcceptCallback = () => {},
      onDeclineCallback = () => {},
      className,
      learnMoreHref = '#',
      ...props
    },
    ref,
  ) => {
    const [isOpen, setIsOpen] = React.useState(false)
    const [hide, setHide] = React.useState(false)
    const t = useTranslations('cookieConsent')

    const handleAccept = React.useCallback(() => {
      setIsOpen(false)
      document.cookie = `${COOKIE_CONSENT_COOKIE}=true; expires=Fri, 31 Dec 9999 23:59:59 GMT`
      setTimeout(() => {
        setHide(true)
      }, 700)
      onAcceptCallback()
    }, [onAcceptCallback])

    const handleDecline = React.useCallback(() => {
      setIsOpen(false)
      setTimeout(() => {
        setHide(true)
      }, 700)
      onDeclineCallback()
    }, [onDeclineCallback])

    React.useEffect(() => {
      try {
        setIsOpen(true)
        if (
          document.cookie.includes(`${COOKIE_CONSENT_COOKIE}=true`) &&
          !demo
        ) {
          setIsOpen(false)
          setTimeout(() => {
            setHide(true)
          }, 700)
        }
      } catch (error) {
        console.warn('Cookie consent error:', error)
      }
    }, [demo])

    if (hide) return null

    const containerClasses = cn(
      'fixed z-50 transition-all duration-700',
      !isOpen ? 'translate-y-full opacity-0' : 'translate-y-0 opacity-100',
      className,
    )

    const commonWrapperProps = {
      ref,
      className: cn(
        containerClasses,
        variant === 'mini'
          ? 'left-0 right-0 sm:left-4 bottom-4 w-full sm:max-w-3xl'
          : 'bottom-0 left-0 right-0 sm:left-4 sm:bottom-4 w-full sm:max-w-md',
      ),
      ...props,
    }

    if (variant === 'default') {
      return (
        <div {...commonWrapperProps}>
          <Card className="m-3 shadow-lg">
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-lg">{t('title')}</CardTitle>
              <Cookie className="h-5 w-5" />
            </CardHeader>
            <CardContent className="space-y-2">
              <CardDescription className="text-sm">
                {t('description')}
              </CardDescription>
              <p className="text-muted-foreground text-xs">
                <RichText>
                  {(tags) => t.rich('byClicking', { ...tags })}
                </RichText>
              </p>
              <a
                href={learnMoreHref}
                className="text-primary text-xs underline underline-offset-4 hover:no-underline"
              >
                {t('learnMore')}
              </a>
            </CardContent>
            <CardFooter className="flex gap-2 pt-2">
              <Button
                onClick={handleDecline}
                variant="secondary"
                className="flex-1"
              >
                {t('decline')}
              </Button>
              <Button onClick={handleAccept} className="flex-1">
                {t('accept')}
              </Button>
            </CardFooter>
          </Card>
        </div>
      )
    }

    if (variant === 'small') {
      return (
        <div {...commonWrapperProps}>
          <Card className="m-3 shadow-lg">
            <CardHeader className="flex h-0 flex-row items-center justify-between space-y-0 px-4 pb-2">
              <CardTitle className="text-base">{t('title')}</CardTitle>
              <Cookie className="h-4 w-4" />
            </CardHeader>
            <CardContent className="px-4 pt-0 pb-2">
              <CardDescription className="text-sm">
                {t('description')}
              </CardDescription>
            </CardContent>
            <CardFooter className="flex h-0 gap-2 px-4 py-2">
              <Button
                onClick={handleDecline}
                variant="secondary"
                size="sm"
                className="flex-1 rounded-full"
              >
                {t('decline')}
              </Button>
              <Button
                onClick={handleAccept}
                size="sm"
                className="flex-1 rounded-full"
              >
                {t('accept')}
              </Button>
            </CardFooter>
          </Card>
        </div>
      )
    }

    if (variant === 'mini') {
      return (
        <div {...commonWrapperProps}>
          <Card className="mx-3 p-0 py-3 shadow-lg">
            <CardContent className="grid gap-4 p-0 px-3.5 sm:flex">
              <CardDescription className="flex-1 text-xs sm:text-sm">
                {t('description')}
              </CardDescription>
              <div className="flex items-center justify-end gap-2 sm:gap-3">
                <Button
                  onClick={handleDecline}
                  size="sm"
                  variant="secondary"
                  className="h-7 text-xs"
                >
                  {t('decline')}
                  <span className="sr-only sm:hidden">{t('decline')}</span>
                </Button>
                <Button
                  onClick={handleAccept}
                  size="sm"
                  className="h-7 text-xs"
                >
                  {t('accept')}
                  <span className="sr-only sm:hidden">{t('accept')}</span>
                </Button>
              </div>
            </CardContent>
          </Card>
        </div>
      )
    }

    return null
  },
)

CookieConsent.displayName = 'CookieConsent'
export { CookieConsent }
export default CookieConsent
