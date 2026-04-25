'use client'

import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import LaniaLogo from '@/components/ui/lania-logo'
import { Separator } from '@/components/ui/separator'
import { cn } from '@/lib/utils'
import { zodResolver } from '@hookform/resolvers/zod'
import {} from '@icons-pack/react-simple-icons'
import {
  isUiNodeInputAttributes,
  LoginFlow,
  RegistrationFlow,
  UiNode,
  UiNodeInputAttributes,
} from '@ory/client-fetch'
import Link from 'next/link'
import { parseAsBoolean, parseAsString, useQueryState } from 'nuqs'
import React from 'react'
import { useForm } from 'react-hook-form'
import { z } from 'zod'
import { groupUiNodes, mapUiNode } from './ory-nodes'

type Props = {
  className?: string
  flow: LoginFlow | RegistrationFlow | undefined
  flowType: 'login' | 'registration' | 'refresh'
}

const formSchema = z
  .object({
    flowType: z.enum(['login', 'registration', 'refresh']),
    csrfToken: z.string().optional(),
    email: z.email({ message: 'Invalid email address' }),
    password: z
      .string()
      .min(6, { message: 'Password must be at least 6 characters' }),
    repeatPassword: z.string(),
  })
  .refine(
    (data) =>
      data.flowType !== 'registration' || data.password === data.repeatPassword,
    { message: "Passwords don't match", path: ['repeatPassword'] },
  )

const getOryUiNodeByGroupAndName = (
  nodes: UiNode[],
  group: string,
  name: string,
) => {
  return nodes.find(
    (node) =>
      node.group === group &&
      (node.attributes as UiNodeInputAttributes).name === name,
  )
}

const AuthForm = ({ flowType, className, flow }: Props) => {
  const [goto] = useQueryState('goto', parseAsString.withDefault(''))

  const [refresh] = useQueryState('refresh', parseAsBoolean.withDefault(false))
  const [flowResult] = React.useState<LoginFlow>()
  const groupedNodes = React.useMemo(
    () => groupUiNodes(flow?.ui.nodes ?? []),
    [flow?.ui.nodes],
  )

  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      csrfToken: '',
      flowType,
      email: '',
      password: '',
      repeatPassword: '',
    },
  })

  React.useEffect(() => {
    if (!flow) return
    form.reset({
      csrfToken:
        (
          getOryUiNodeByGroupAndName(flow.ui.nodes, 'default', 'csrf_token')
            ?.attributes as UiNodeInputAttributes
        )?.value ?? '',
      flowType,
      email:
        (
          getOryUiNodeByGroupAndName(flow.ui.nodes, 'default', 'identifier')
            ?.attributes as UiNodeInputAttributes
        )?.value ?? '',
      password: '',
      repeatPassword: '',
    })
  }, [flow?.id])

  React.useEffect(() => {
    const msg = flowResult?.ui?.messages?.find(
      (message) => message.type === 'error',
    )?.text
    if (msg) {
      form.setError('root', {
        message: msg,
      })
    } else {
      form.clearErrors('root')
    }
  }, [flowResult?.ui?.messages])

  return (
    <Card className={cn('max-w-sm', className)}>
      <CardHeader>
        <CardTitle className="flex items-center justify-between gap-2">
          {flowType === 'registration'
            ? 'Sign Up'
            : flowType === 'refresh'
              ? 'Confirm your identity'
              : 'Sign In'}
          <Link href="/" className="hover:opacity-80">
            <LaniaLogo />
          </Link>
        </CardTitle>
        <CardDescription>
          {flowType === 'registration'
            ? 'Sign up using a social provider'
            : flowType === 'refresh'
              ? 'Prove your identity using a social provider'
              : 'Sign in using a social provider'}
        </CardDescription>
      </CardHeader>
      <CardContent className="grid gap-4">
        {Object.entries(groupedNodes).map(([group, nodes], index) => (
          <div
            key={group}
            className={cn(
              'flex flex-col gap-4',
              index > 0 && Object.keys(groupedNodes).length > 2 && 'pt-4',
            )}
          >
            {index > 0 && Object.keys(groupedNodes).length > 2 && <Separator />}
            <form
              key={group}
              noValidate
              action={flow?.ui.action}
              method={flow?.ui.method}
              className="flex flex-col"
            >
              {nodes.map((node, index) => {
                if (
                  'value' in node.attributes &&
                  node.attributes.value === 'twitch-extended'
                ) {
                  return null
                }
                return (
                  <div
                    key={node.type + index}
                    className={cn(
                      'pt-4',
                      index === 0 && 'pt-0',
                      isUiNodeInputAttributes(node.attributes) &&
                        (['traits.username', 'traits.avatar_url'].includes(
                          node.attributes.name,
                        ) ||
                          node.attributes.type === 'hidden') &&
                        'pt-0',
                    )}
                  >
                    {mapUiNode(flow, node)}
                  </div>
                )
              })}
            </form>
          </div>
        ))}
        {flow?.ui.messages?.map((message) => (
          <p
            key={message.id}
            className={cn(
              'text-muted-foreground text-sm',
              message.type === 'error' && 'text-destructive',
            )}
          >
            {message.text}
          </p>
        ))}
      </CardContent>
      <CardFooter className="flex flex-col gap-2">
        {!refresh && (
          <div className="text-center text-sm">
            {flowType === 'registration'
              ? 'Already have an account?'
              : "Don't have an account?"}
            &nbsp;
            <Link
              href={{
                pathname:
                  flowType === 'registration'
                    ? `${process.env.NEXT_PUBLIC_ORY_SDK_URL}/self-service/login/browser`
                    : `${process.env.NEXT_PUBLIC_ORY_SDK_URL}/self-service/registration/browser`,
                query: {
                  return_to: flow?.return_to ?? goto,
                },
              }}
              className="underline"
            >
              {flowType === 'registration' ? 'Sign in' : 'Sign up'}
            </Link>
          </div>
        )}
      </CardFooter>
    </Card>
  )
}

export default AuthForm
