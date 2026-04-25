import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { cn } from '@/lib/utils'
import {
  isUiNodeInputAttributes,
  LoginFlow,
  RegistrationFlow,
  UiNode,
  UiNodeGroupEnum,
  UiNodeInputAttributes,
} from '@ory/client-fetch'
import ProviderIcon from './provider-icon'

export const groupUiNodes = (
  nodes: UiNode[],
): Record<UiNodeGroupEnum, UiNode[]> => {
  return nodes.reduce(
    (acc, node) => {
      const group = node.group ?? UiNodeGroupEnum.Default
      const assigned =
        group === UiNodeGroupEnum.Oidc
          ? UiNodeGroupEnum.Oidc
          : UiNodeGroupEnum.Default
      if (!acc[assigned]) {
        acc[assigned] = []
      }
      acc[assigned].push(node)
      return acc
    },
    {} as Record<UiNodeGroupEnum, UiNode[]>,
  )
}

export const mapUiNodeGroups = (
  flow: LoginFlow | RegistrationFlow,
  groups: Record<UiNodeGroupEnum, UiNode[]>,
) => {
  return Object.entries(groups).map(([group, nodes]) => {
    return (
      <div
        key={group}
        className={cn(
          'grid grid-cols-1 gap-4 py-2',
          group === 'oidc' && 'grid-cols-2 gap-2',
        )}
      >
        {nodes.map((node) => mapUiNode(flow, node))}
      </div>
    )
  })
}

export const mapUiNode = (
  flow: LoginFlow | RegistrationFlow | undefined,
  node: UiNode,
) => {
  // const active = flow?.active !== undefined
  if (!isUiNodeInputAttributes(node.attributes)) return null

  const attrs = node.attributes as UiNodeInputAttributes
  const nodeType = attrs.type
  const provider = (node.meta.label?.context as { provider: string })?.provider

  // skip extended providers (e.g. twitch-extended)
  if (provider && attrs.value?.toLowerCase().includes('extended')) {
    return null
  }

  if (['traits.username', 'traits.avatar_url'].includes(attrs.name)) {
    return null
  }

  switch (nodeType) {
    case 'email':
    case 'text':
    case 'password':
    case 'hidden':
      return (
        <div className="grid gap-2">
          <Label hidden={attrs.type === 'hidden'} htmlFor={attrs.name}>
            {node.meta.label?.text}
          </Label>
          <Input
            type={attrs.type}
            required={attrs.required}
            name={attrs.name}
            disabled={attrs.disabled}
            maxLength={attrs.maxlength}
            autoComplete={attrs.autocomplete}
            defaultValue={attrs.value}
          />
          {node.messages?.map((message) => (
            <p
              key={message.id}
              className={cn(
                'text-muted-foreground text-xs',
                message.type === 'error' && 'text-destructive',
              )}
            >
              {message.text}
            </p>
          ))}
        </div>
      )
    case 'button':
    case 'submit':
      return (
        <Button
          variant={attrs.type === 'submit' ? 'default' : 'secondary'}
          className="flex w-full items-center gap-2 px-2"
          value={attrs.value}
          disabled={attrs.disabled}
          name={attrs.name}
          type={attrs.type as 'submit' | 'button' | 'reset' | undefined}
        >
          {attrs.name === 'provider' && <ProviderIcon providerId={provider} />}
          {node.meta.label?.text}
        </Button>
      )
  }
}

// export function handleGetFlowError<S>(
//   router: NextRouter,
//   flowType: 'login' | 'registration' | 'settings' | 'recovery' | 'verification',
//   resetFlow: Dispatch<SetStateAction<S | undefined>>,
// ) {
//   return async (err: AxiosError) => {
//     switch (err.response?.data.error?.id) {
//       case 'session_aal2_required':
//         // 2FA is enabled and enforced, but user did not perform 2fa yet!
//         window.location.href = err.response?.data.redirect_browser_to
//         return
//       case 'session_already_available':
//         // User is already signed in, let's redirect them home!
//         await router.push('/')
//         return
//       case 'session_refresh_required':
//         // We need to re-authenticate to perform this action
//         window.location.href = err.response?.data.redirect_browser_to
//         return
//       case 'self_service_flow_return_to_forbidden':
//         // The flow expired, let's request a new one.
//         toast({
//           title: 'The return_to address is not allowed.',
//           variant: 'destructive',
//         })
//         resetFlow(undefined)
//         await router.push('/' + flowType)
//         return
//       case 'self_service_flow_expired':
//         // The flow expired, let's request a new one.
//         toast({
//           title: 'Your interaction expired, please fill out the form again.',
//           variant: 'destructive',
//         })
//         resetFlow(undefined)
//         await router.push('/' + flowType)
//         return
//       case 'security_csrf_violation':
//         // A CSRF violation occurred. Best to just refresh the flow!
//         toast({
//           title:
//             'A security violation was detected, please fill out the form again.',
//           variant: 'destructive',
//         })
//         resetFlow(undefined)
//         await router.push('/' + flowType)
//         return
//       case 'security_identity_mismatch':
//         // The requested item was intended for someone else. Let's request a new flow...
//         resetFlow(undefined)
//         await router.push('/' + flowType)
//         return
//       case 'browser_location_change_required':
//         // Ory Kratos asked us to point the user to this URL.
//         window.location.href = err.response.data.redirect_browser_to
//         return
//     }

//     switch (err.response?.status) {
//       case 410:
//         // The flow expired, let's request a new one.
//         resetFlow(undefined)
//         await router.push('/' + flowType)
//         return
//     }

//     // We are not able to handle the error? Return it.
//     return Promise.reject(err)
//   }
// }
