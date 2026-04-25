'use client'

import { useModalStore } from '@/providers/modal'
import { parseAsString, parseAsStringEnum, useQueryState } from 'nuqs'
import qs from 'querystring'
import React from 'react'
import { ModalParamValue, ModalType } from '../stores/modal'

// Serialize params
const serializeParams = (params?: Record<string, ModalParamValue>) =>
  qs.stringify(params)

// Deserialize params
const deserializeParams = (params: string) => qs.parse(params)

export const ModalQuerySync = () => {
  const { currentModal, modalParams, openModal, closeModal } = useModalStore(
    (state) => state,
  )

  const [modalQuery, setModalQuery] = useQueryState(
    'm',
    parseAsStringEnum<ModalType>(Object.values(ModalType)).withDefault(
      ModalType.None,
    ),
  )

  const [modalParamsQuery, setModalParamsQuery] = useQueryState(
    'mps',
    parseAsString.withDefault(''),
  )

  // Guard to prevent useEffect loops
  const isInternalChange = React.useRef(false)

  // Sync Zustand state to the query string
  React.useEffect(() => {
    if (isInternalChange.current) {
      isInternalChange.current = false // Reset the flag
      return
    }

    if (currentModal) {
      setModalQuery(currentModal)
      setModalParamsQuery(serializeParams(modalParams)) // Serialize params as JSON
    } else {
      setModalQuery(null)
      setModalParamsQuery('')
    }
  }, [currentModal, modalParams, setModalQuery, setModalParamsQuery])

  // Sync query string back to Zustand state
  React.useEffect(() => {
    if (modalQuery) {
      const params = modalParamsQuery ? deserializeParams(modalParamsQuery) : {} // Deserialize params
      if (
        currentModal !== modalQuery ||
        JSON.stringify(modalParams) !== JSON.stringify(params)
      ) {
        isInternalChange.current = true // Mark as internal change
        openModal(modalQuery, params)
      }
    } else if (currentModal) {
      isInternalChange.current = true // Mark as internal change
      closeModal()
    }
  }, [
    modalQuery,
    modalParamsQuery,
    currentModal,
    modalParams,
    openModal,
    closeModal,
  ])

  return null
}
