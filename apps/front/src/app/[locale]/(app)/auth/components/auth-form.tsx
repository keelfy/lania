'use client'

import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import LaniaLogo from '@/components/ui/lania-logo'
import { Label } from '@/components/ui/label'
import LoadingSpinner from '@/components/ui/loading-spinner'
import { Link } from '@/i18n/navigation'
import { cn } from '@/lib/utils'
import {
  isUiNodeInputAttributes,
  UiNode,
  UiNodeGroupEnum,
  UiText,
} from '@ory/client-fetch'
import { useTranslations } from 'next-intl'
import ProviderIcon from './provider-icon'
import { AuthFlow, AuthFlowType, useAuthFlow } from './use-auth-flow'

type Props = {
  className?: string
  type: AuthFlowType
}

// Kratos message ids with a translation. Other messages are shown in the Kratos text.
const TRANSLATED_MESSAGES: Record<number, string> = {
  4010001: 'flowExpired',
  4040001: 'flowExpired',
  4000008: 'duplicateAccount',
  4000028: 'duplicateAccount',
  4000029: 'duplicateAccount',
}

// The page already tells the user to sign in again.
const HIDDEN_MESSAGES = [1010003]

// Traits the provider fills in. They are sent back as they are and never shown.
const HIDDEN_TRAITS = ['traits.username', 'traits.avatarUrl']

// Kratos button label ids.
const LABEL_WITH_PROVIDER = [1010002, 1040002]
const LABEL_SIGN_IN = [1010001]
const LABEL_SIGN_UP = [1040001]
const LABEL_CONTINUE = [1010013, 1040003]

export default function AuthForm({ type, className }: Props) {
  const t = useTranslations('auth')
  const { flow, failed, refresh } = useAuthFlow(type)
  const mode =
    type === 'registration' ? 'registration' : refresh ? 'refresh' : 'login'

  return (
    <Card className={cn('max-w-sm', className)}>
      <CardHeader>
        <CardTitle className="flex items-center justify-between gap-2">
          {t(`${mode}.title`)}
          <Link href="/" className="hover:opacity-80">
            <LaniaLogo />
          </Link>
        </CardTitle>
        <CardDescription>{t(`${mode}.description`)}</CardDescription>
      </CardHeader>
      <CardContent>
        {failed ? (
          <div className="flex flex-col gap-3">
            <p className="text-destructive text-sm">{t('loadFailed')}</p>
            <Button variant="secondary" className="w-full" asChild>
              <Link href="/">{t('backHome')}</Link>
            </Button>
          </div>
        ) : flow ? (
          <FlowForm flow={flow} type={type} />
        ) : (
          <div className="flex min-h-24 items-center justify-center">
            <LoadingSpinner />
          </div>
        )}
      </CardContent>
    </Card>
  )
}

function FlowMessages({
  messages,
  className,
}: {
  messages?: UiText[]
  className?: string
}) {
  const t = useTranslations('auth.messages')
  const shown = messages?.filter(
    (message) => !HIDDEN_MESSAGES.includes(message.id),
  )
  if (!shown?.length) return null
  return (
    <div className={cn('flex flex-col gap-1', className)}>
      {shown.map((message) => (
        <p
          key={message.id}
          className={cn(
            'text-muted-foreground text-sm',
            message.type === 'error' && 'text-destructive',
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

// Sign-in is SSO only, so only these node groups are ever rendered.
// Password, identifier_first, code and link nodes are dropped even if Kratos sends them.
const ALLOWED_GROUPS: Record<AuthFlowType, UiNodeGroupEnum[]> = {
  login: [UiNodeGroupEnum.Default, UiNodeGroupEnum.Oidc],
  registration: [
    UiNodeGroupEnum.Default,
    UiNodeGroupEnum.Oidc,
    UiNodeGroupEnum.Profile,
  ],
}

function FlowForm({ flow, type }: { flow: AuthFlow; type: AuthFlowType }) {
  const allowed = ALLOWED_GROUPS[type]
  const nodes = flow.ui.nodes.filter((node) => allowed.includes(node.group))
  return (
    <form
      noValidate
      action={flow.ui.action}
      method={flow.ui.method}
      className="flex flex-col gap-4"
    >
      {nodes.map((node, index) => (
        <FlowNode key={`${node.group}-${index}`} node={node} />
      ))}
      <FlowMessages messages={flow.ui.messages} />
    </form>
  )
}

function FlowNode({ node }: { node: UiNode }) {
  const t = useTranslations('auth')
  if (!isUiNodeInputAttributes(node.attributes)) return null
  const attrs = node.attributes
  const label = node.meta.label
  const provider = (label?.context as { provider?: string } | undefined)
    ?.provider

  // Providers like twitch-extended only differ by the scopes they ask for.
  if (
    node.group === UiNodeGroupEnum.Oidc &&
    attrs.value?.includes('extended')
  ) {
    return null
  }

  if (attrs.type === 'hidden' || HIDDEN_TRAITS.includes(attrs.name)) {
    return (
      <input type="hidden" name={attrs.name} defaultValue={attrs.value ?? ''} />
    )
  }

  switch (attrs.type) {
    case 'submit':
    case 'button': {
      const text =
        label && LABEL_WITH_PROVIDER.includes(label.id)
          ? t('continueWith', { provider: provider ?? '' })
          : label && LABEL_SIGN_IN.includes(label.id)
            ? t('signIn')
            : label && LABEL_SIGN_UP.includes(label.id)
              ? t('createAccount')
              : label && LABEL_CONTINUE.includes(label.id)
                ? t('continue')
                : label?.text
      return (
        <Button
          type={attrs.type}
          name={attrs.name}
          value={attrs.value}
          disabled={attrs.disabled}
          className="flex w-full items-center gap-2 px-2"
        >
          {attrs.name === 'provider' && (
            <ProviderIcon providerId={attrs.value} className="size-5" />
          )}
          {text}
        </Button>
      )
    }
    case 'email':
    case 'text':
      return (
        <div className="grid gap-2">
          <Label htmlFor={attrs.name}>
            {attrs.name === 'traits.email' ? t('fields.email') : label?.text}
          </Label>
          <Input
            id={attrs.name}
            type={attrs.type}
            name={attrs.name}
            required={attrs.required}
            disabled={attrs.disabled}
            maxLength={attrs.maxlength}
            autoComplete={attrs.autocomplete}
            defaultValue={attrs.value}
          />
          <FlowMessages messages={node.messages} />
        </div>
      )
    default:
      return null
  }
}
