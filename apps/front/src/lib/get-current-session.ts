import ory from '@/lib/ory'
import { isResponseError, Session } from '@ory/client-fetch'
import { headers } from 'next/headers'
import { cache } from 'react'

// Deduplicated per request: the root layout, the navbar, and every
// requireAdmin() call would otherwise each hit Kratos separately.
export const getCurrentSession = cache(async () => {
  try {
    return (await ory.toSession({
      cookie: (await headers()).get('cookie') ?? '',
    })) as Session
  } catch (error) {
    if (isResponseError(error) && error.response.status === 401) {
      return undefined
    }
    console.error(error)
    return undefined
  }
})

export async function isCurrentSessionActive() {
  const currentSession = await getCurrentSession()
  return currentSession?.active === true
}
