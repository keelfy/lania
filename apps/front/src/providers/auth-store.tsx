'use client'

import React from 'react'
import { useStore } from 'zustand'

import createAuthStore, { AuthStore } from '@/stores/auth-store'
import { Session } from '@ory/client-fetch'

export type AuthStoreApi = ReturnType<typeof createAuthStore>

export const AuthStoreContext = React.createContext<AuthStoreApi | undefined>(
  undefined,
)

export type AuthStoreProviderProps = React.PropsWithChildren<{
  session: Session | undefined
}>

export default function AuthStoreProvider({
  children,
  ...props
}: AuthStoreProviderProps) {
  const storeRef = React.useRef<AuthStoreApi>(createAuthStore(props))

  return (
    <AuthStoreContext.Provider value={storeRef.current}>
      {children}
    </AuthStoreContext.Provider>
  )
}

export function useAuthStore<T>(selector: (store: AuthStore) => T): T {
  const storeContext = React.useContext(AuthStoreContext)

  if (!storeContext) {
    throw new Error(`useAuthStore must be used within AuthStoreProvider`)
  }

  return useStore(storeContext, selector)
}
