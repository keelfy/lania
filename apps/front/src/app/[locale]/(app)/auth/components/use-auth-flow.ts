'use client'

import { usePathname, useRouter } from '@/i18n/navigation'
import ory from '@/lib/ory'
import { isResponseError, LoginFlow, RegistrationFlow } from '@ory/client-fetch'
import { useSearchParams } from 'next/navigation'
import React from 'react'

export type AuthFlowType = 'login' | 'registration'
export type AuthFlow = LoginFlow | RegistrationFlow

// The return path must stay on the site, so only a plain path is accepted.
const parseGoto = (goto: string | null) =>
  goto && goto.startsWith('/') && !goto.startsWith('//') ? goto : '/'

// Loads the flow from ?flow=, or starts a new login flow.
// Sign in and sign up are one step for social providers, so a registration flow is only opened by Kratos
// when the provider left out data the account needs. A registration page without a flow goes to sign in.
export function useAuthFlow(type: AuthFlowType) {
  const router = useRouter()
  const pathname = usePathname()
  const searchParams = useSearchParams()
  const flowId = searchParams.get('flow') ?? ''
  const goto = parseGoto(searchParams.get('goto'))
  const refresh = searchParams.get('refresh') === 'true'

  const [flow, setFlow] = React.useState<AuthFlow>()
  const [failed, setFailed] = React.useState(false)
  const loadedId = React.useRef('')

  React.useEffect(() => {
    // The flow id is put into the URL after a load, it needs no second load.
    if (flowId && flowId === loadedId.current) return
    let cancelled = false
    const returnTo = process.env.NEXT_PUBLIC_DOMAIN + goto
    setFailed(false)

    const load = async (): Promise<AuthFlow | undefined> => {
      if (flowId) {
        try {
          return type === 'login'
            ? await ory.getLoginFlow({ id: flowId })
            : await ory.getRegistrationFlow({ id: flowId })
        } catch (error) {
          // An expired or foreign flow is replaced by a new one.
          console.log('Failed to load existing flow', error)
        }
      }

      if (type === 'registration') {
        router.replace(`/auth/login?goto=${encodeURIComponent(goto)}`)
        return undefined
      }

      try {
        return await ory.createBrowserLoginFlow({ returnTo, refresh })
      } catch (error) {
        const res = isResponseError(error)
          ? await error.response.json().catch(() => ({}))
          : undefined
        if (res?.error?.id === 'session_already_available') {
          if (!cancelled) window.location.href = returnTo
          return undefined
        }
        console.error('Failed to create login flow', error)
        if (!cancelled) setFailed(true)
        return undefined
      }
    }

    load().then((loaded) => {
      if (cancelled || !loaded) return
      loadedId.current = loaded.id
      setFlow(loaded)
      if (loaded.id !== flowId) {
        const params = new URLSearchParams(searchParams.toString())
        params.set('flow', loaded.id)
        router.replace(`${pathname}?${params}`)
      }
    })

    return () => {
      cancelled = true
    }
  }, [flowId, goto, refresh, type, pathname, router, searchParams])

  return { flow, failed, refresh }
}
