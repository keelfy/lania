'use client'

import React from 'react'
import { useStore } from 'zustand'

import createAuthStore, { AuthStore } from '@/stores/auth-store'

export type AuthStoreApi = ReturnType<typeof createAuthStore>

export const AuthStoreContext = React.createContext<AuthStoreApi | undefined>(
  undefined,
)

// The store starts without a session, ViewerSync fills it in once the server has read it.
export default function AuthStoreProvider({
  children,
}: React.PropsWithChildren) {
  const [store] = React.useState<AuthStoreApi>(() => createAuthStore())

  return (
    <AuthStoreContext.Provider value={store}>
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
