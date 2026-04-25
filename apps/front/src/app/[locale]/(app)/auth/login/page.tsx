'use client'

import ory from '@/lib/ory'
import { isResponseError, LoginFlow } from '@ory/client-fetch'
import { redirect } from 'next/navigation'
import { useQueryState } from 'nuqs'
import { parseAsBoolean, parseAsString } from 'nuqs/server'
import React from 'react'
import AuthForm from '../components/auth-form'

export default function SignInPage() {
  const [goto] = useQueryState('goto', parseAsString.withDefault('/'))
  const [refresh] = useQueryState('refresh', parseAsBoolean.withDefault(false))
  const [flowId, setFlowId] = useQueryState(
    'flow',
    parseAsString.withDefault(''),
  )

  const [flow, setFlow] = React.useState<LoginFlow>()
  const [, startFlowTransition] = React.useTransition()

  React.useEffect(() => {
    if (flow) return

    startFlowTransition(async () => {
      let flow: LoginFlow | undefined = undefined

      if (flowId.length > 0) {
        try {
          flow = await ory.getLoginFlow({
            id: flowId,
          })
        } catch (error) {
          if (isResponseError(error)) {
            error.response
              .json()
              .then((res) =>
                console.log(
                  'Failed to load existing authorization flow',
                  JSON.stringify(res.error.message),
                ),
              )
          }
          flow = undefined
        }
      }

      if (!flow) {
        const returnTo =
          process.env.NEXT_PUBLIC_DOMAIN + decodeURIComponent(goto)
        try {
          flow = await ory.createBrowserLoginFlow({
            returnTo,
            refresh,
          })
        } catch (error) {
          if (isResponseError(error)) {
            const res = await error.response.json()

            switch (res.error.id) {
              case 'session_already_available':
                return redirect(returnTo)
              default:
                console.log(
                  'Failed to create authorization flow',
                  JSON.stringify(res.error.message),
                )
                break
            }
          }
        }
      }

      setFlow(flow)
    })
  }, [])

  React.useEffect(() => {
    if (flowId !== flow?.id) {
      setFlowId(flow?.id ?? '')
    }
  }, [flow])

  return (
    <AuthForm
      flowType={refresh ? 'refresh' : 'login'}
      flow={flow}
      className="mx-auto flex-1"
    />
  )
}
