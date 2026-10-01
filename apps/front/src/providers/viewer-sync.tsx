'use client'

import { useBasket } from '@/context/basket'
import { useAuthStore } from '@/providers/auth-store'
import { Session } from '@ory/client-fetch'
import { BasketItem } from '@/models/basket'
import { useEffect } from 'react'

// Puts the session and the basket the server read for this request into their stores. Until it runs, the stores are empty.
export default function ViewerSync({
  session,
  basket,
}: {
  session: Session | undefined
  basket: BasketItem[]
}) {
  const updateSession = useAuthStore((state) => state.updateSession)
  const clearSession = useAuthStore((state) => state.clearSession)
  const { setItems } = useBasket()

  useEffect(() => {
    if (session) updateSession(session)
    else clearSession()
  }, [session, updateSession, clearSession])

  useEffect(() => {
    setItems(basket)
  }, [basket, setItems])

  return null
}
