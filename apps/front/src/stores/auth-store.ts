import { Session } from '@ory/client-fetch'
import { createStore } from 'zustand'
import { devtools } from 'zustand/middleware'

type Actions = {
  updateSession: (session: Session) => void
  clearSession: () => void
}

type State = {
  session: Session | undefined
}

export type AuthStore = Actions & State

const defaultInitialState: State = {
  session: undefined,
}

const createAuthStore = (initialState: State = defaultInitialState) => {
  return createStore<AuthStore>()(
    devtools((set) => ({
      ...initialState,
      updateSession: (session) => set(() => ({ session })),
      clearSession: () => set(() => ({ ...initialState })),
    })),
  )
}

export default createAuthStore
