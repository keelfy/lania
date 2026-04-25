'use client'

import React from 'react'
import { useStore } from 'zustand'

import createModalStore, { ModalStore, ModalType } from '@/stores/modal'
import { parseAsString, parseAsStringEnum, useQueryState } from 'nuqs'
import qs from 'querystring'
import { ModalQuerySync } from '@/query-params/modal'

export type ModalStoreApi = ReturnType<typeof createModalStore>

export const ModalStoreContext = React.createContext<ModalStoreApi | undefined>(
  undefined,
)

export type ModalStoreProviderProps = {
  children: React.ReactNode
}

export default function ModalStoreProvider({
  children,
}: ModalStoreProviderProps) {
  const [modalQuery] = useQueryState(
    'm',
    parseAsStringEnum<ModalType>(Object.values(ModalType)).withDefault(
      ModalType.None,
    ),
  )

  const [modalParamsQuery] = useQueryState('mps', parseAsString.withDefault(''))

  const storeRef = React.useRef<ModalStoreApi>(
    createModalStore({
      currentModal: modalQuery,
      modalParams: modalParamsQuery ? qs.parse(modalParamsQuery) : {},
    }),
  )

  return (
    <ModalStoreContext.Provider value={storeRef.current}>
      {children}
      <ModalQuerySync />
    </ModalStoreContext.Provider>
  )
}

export function useModalStore<T>(selector: (store: ModalStore) => T): T {
  const storeContext = React.useContext(ModalStoreContext)

  if (!storeContext) {
    throw new Error(`useModalStore must be used within ModalStoreProvider`)
  }

  return useStore(storeContext, selector)
}
