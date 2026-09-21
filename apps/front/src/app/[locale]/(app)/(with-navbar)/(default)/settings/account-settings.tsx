'use client'

import ProviderIcon from '@/app/[locale]/(app)/auth/components/provider-icon'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import LoadingSpinner from '@/components/ui/loading-spinner'
import ory from '@/lib/ory'
import { cn } from '@/lib/utils'
import {
  isResponseError,
  isUiNodeInputAttributes,
  SettingsFlow,
  UiNode,
  UiNodeGroupEnum,
  UiNodeInputAttributes,
  UiText,
} from '@ory/client-fetch'
import { useLocale, useTranslations } from 'next-intl'
import { usePathname, useRouter, useSearchParams } from 'next/navigation'
import React from 'react'
import DeleteAccountCard from './delete-account-card'

// Must match selfservice.methods.oidc.config.providers in kratos.yml.
// Kratos sends no node for the only linked provider, so this list is the only way to show it.
const SSO_PROVIDERS = [
  { id: 'google', label: 'Google' },
  { id: 'discord', label: 'Discord' },
  { id: 'yandex', label: 'Yandex' },
  { id: 'twitch-basic', label: 'Twitch' },
]

// Kratos message ids with a translation. Other messages are shown in the Kratos text.
const TRANSLATED_MESSAGES: Record<number, string> = {
  1050001: 'saved',
}

const inputAttributes = (node: UiNode) =>
  isUiNodeInputAttributes(node.attributes)
    ? (node.attributes as UiNodeInputAttributes)
    : undefined

const findInput = (nodes: UiNode[], group: string, name: string) =>
  nodes
    .map((node) => (node.group === group ? inputAttributes(node) : undefined))
    .find((attrs) => attrs?.name === name)

// Loads the settings flow from ?flow=, or starts a new one.
function useSettingsFlow() {
  const locale = useLocale()
  const router = useRouter()
  const pathname = usePathname()
  const searchParams = useSearchParams()
  const flowId = searchParams.get('flow') ?? ''
  const [flow, setFlow] = React.useState<SettingsFlow>()
  const [failed, setFailed] = React.useState(false)
  const loadedId = React.useRef('')

  React.useEffect(() => {
    // The flow id is put into the URL after a load, it needs no second load.
    if (flowId && flowId === loadedId.current) return
    let cancelled = false

    const handleError = async (error: unknown) => {
      if (!isResponseError(error)) throw error
      const res = await error.response.json().catch(() => ({}))
      if (res.redirect_browser_to) {
        window.location.href = res.redirect_browser_to
        return true
      }
      if (
        res.error?.id === 'session_inactive' ||
        error.response.status === 401
      ) {
        router.replace(
          `/${locale}/auth/login?goto=${encodeURIComponent(pathname)}`,
        )
        return true
      }
      return false
    }

    const load = async () => {
      if (flowId) {
        try {
          return await ory.getSettingsFlow({ id: flowId })
        } catch (error) {
          // An expired or foreign flow is replaced by a new one.
          if (await handleError(error)) return undefined
        }
      }
      try {
        return await ory.createBrowserSettingsFlow()
      } catch (error) {
        if (!(await handleError(error))) {
          console.error('Failed to create settings flow', error)
          setFailed(true)
        }
        return undefined
      }
    }

    load().then((loaded) => {
      if (cancelled || !loaded) return
      loadedId.current = loaded.id
      setFlow(loaded)
      if (loaded.id !== flowId) {
        router.replace(`${pathname}?flow=${loaded.id}`)
      }
    })

    return () => {
      cancelled = true
    }
  }, [flowId, locale, pathname, router])

  return { flow, failed }
}

export default function AccountSettings() {
  const t = useTranslations('accountSettings')
  const { flow, failed } = useSettingsFlow()

  if (failed) {
    return <p className="text-destructive text-sm">{t('loadFailed')}</p>
  }
  if (!flow) {
    return (
      <div className="flex min-h-40 items-center justify-center">
        <LoadingSpinner />
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-4">
      <FlowMessages messages={flow.ui.messages} />
      <EmailCard flow={flow} />
      <LinkedAccountsCard flow={flow} />
      <DeleteAccountCard />
    </div>
  )
}

function FlowMessages({
  messages,
  className,
}: {
  messages?: UiText[]
  className?: string
}) {
  const t = useTranslations('accountSettings.messages')
  if (!messages?.length) return null
  return (
    <div className={cn('flex flex-col gap-1', className)}>
      {messages.map((message) => (
        <p
          key={message.id}
          className={cn(
            'text-sm',
            message.type === 'error' ? 'text-destructive' : 'text-primary',
          )}
        >
          {TRANSLATED_MESSAGES[message.id]
            ? t(TRANSLATED_MESSAGES[message.id])
            : message.text}
        </p>
      ))}
    </div>
  )
}

// The CSRF token Kratos checks on every submit of the flow.
function CsrfInput({ flow }: { flow: SettingsFlow }) {
  const csrf = findInput(flow.ui.nodes, UiNodeGroupEnum.Default, 'csrf_token')
  return <input type="hidden" name="csrf_token" value={csrf?.value ?? ''} />
}

function EmailCard({ flow }: { flow: SettingsFlow }) {
  const t = useTranslations('accountSettings.email')
  const profileNodes = flow.ui.nodes.filter(
    (node) => node.group === UiNodeGroupEnum.Profile,
  )
  const emailNode = profileNodes.find(
    (node) => inputAttributes(node)?.name === 'traits.email',
  )
  const email = emailNode && inputAttributes(emailNode)
  if (!email) return null

  // Kratos replaces all traits on submit, so the other traits are sent back unchanged.
  const otherTraits = profileNodes
    .map(inputAttributes)
    .filter(
      (attrs) =>
        attrs?.name.startsWith('traits.') && attrs.name !== 'traits.email',
    )

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-lg">{t('title')}</CardTitle>
        <CardDescription>{t('description')}</CardDescription>
      </CardHeader>
      <CardContent>
        <form
          action={flow.ui.action}
          method={flow.ui.method}
          className="flex flex-col gap-2"
        >
          <CsrfInput flow={flow} />
          {otherTraits.map((attrs) => (
            <input
              key={attrs!.name}
              type="hidden"
              name={attrs!.name}
              value={attrs!.value ?? ''}
            />
          ))}
          <Label htmlFor="traits.email">{t('label')}</Label>
          <div className="flex flex-col gap-2 sm:flex-row">
            <Input
              id="traits.email"
              type="email"
              name="traits.email"
              required
              autoComplete="email"
              defaultValue={email.value ?? ''}
              aria-invalid={emailNode.messages.some(
                (message) => message.type === 'error',
              )}
            />
            <Button type="submit" name="method" value="profile">
              {t('save')}
            </Button>
          </div>
          <FlowMessages messages={emailNode.messages} />
        </form>
      </CardContent>
    </Card>
  )
}

function LinkedAccountsCard({ flow }: { flow: SettingsFlow }) {
  const t = useTranslations('accountSettings.sso')
  const oidcNodes = flow.ui.nodes
    .filter((node) => node.group === UiNodeGroupEnum.Oidc)
    .map(inputAttributes)
    .filter((attrs) => attrs && ['link', 'unlink'].includes(attrs.name))
  const hasUnlink = oidcNodes.some((attrs) => attrs?.name === 'unlink')
  const messages = flow.ui.nodes
    .filter((node) => node.group === UiNodeGroupEnum.Oidc)
    .flatMap((node) => node.messages)

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-lg">{t('title')}</CardTitle>
        <CardDescription>{t('description')}</CardDescription>
      </CardHeader>
      <CardContent>
        <form
          action={flow.ui.action}
          method={flow.ui.method}
          className="flex flex-col gap-3"
        >
          <CsrfInput flow={flow} />
          {SSO_PROVIDERS.map((provider) => {
            const node = oidcNodes.find((attrs) => attrs?.value === provider.id)
            // Kratos sends no unlink node at all when only one provider is linked, the one without a node.
            const linked = node ? node.name === 'unlink' : !hasUnlink
            return (
              <div
                key={provider.id}
                className="flex items-center justify-between gap-2"
              >
                <div className="flex items-center gap-2">
                  <ProviderIcon providerId={provider.id} className="size-5" />
                  <span className="font-semibold">{provider.label}</span>
                  {linked && <Badge variant="secondary">{t('linked')}</Badge>}
                </div>
                {node ? (
                  <Button
                    type="submit"
                    name={node.name}
                    value={node.value}
                    variant={node.name === 'unlink' ? 'outline' : 'default'}
                    size="sm"
                  >
                    {t(node.name === 'unlink' ? 'unlink' : 'link')}
                  </Button>
                ) : (
                  <span className="text-muted-foreground text-right text-xs">
                    {hasUnlink ? t('unavailable') : t('lastMethod')}
                  </span>
                )}
              </div>
            )
          })}
          <FlowMessages messages={messages} />
        </form>
      </CardContent>
    </Card>
  )
}
