'use client'

import { getBasket } from '@/lib/api-endpoints'
import { clientApiFetcher } from '@/lib/client'
import { BasketItem } from '@/models/basket'
import React from 'react'
import { v4 as uuidv4 } from 'uuid'

type Actions = {
  items: BasketItem[]
  addItem: (productId: string, profileId: string) => string
  removeItem: (id: string) => void
  clearItems: () => void
  setItems: (items: BasketItem[]) => void
  refresh: () => Promise<void>
}

export const BasketContext = React.createContext<Actions>({
  items: [],
  addItem: () => '',
  setItems: () => {},
  removeItem: () => {},
  clearItems: () => {},
  refresh: () => Promise.resolve(),
})

export const useBasket = () => {
  return React.useContext(BasketContext)
}

type Props = {
  initialBasket: BasketItem[]
}

export const BasketProvider = ({
  children,
  initialBasket,
}: React.PropsWithChildren<Props>) => {
  const [basket, setBasket] = React.useState<BasketItem[]>(initialBasket)

  const addItem = React.useCallback((productId: string, profileId: string) => {
    const id = uuidv4()
    setBasket((prev) => [...prev, { productId, profileId, id, quantity: 1 }])
    return id
  }, [])

  const removeItem = React.useCallback((id: string) => {
    setBasket((prev) => prev.filter((item) => item.id !== id))
  }, [])

  const setItems = React.useCallback((items: BasketItem[]) => {
    setBasket(items)
  }, [])

  const clearItems = React.useCallback(() => {
    setBasket([])
  }, [])

  const refresh = React.useCallback(async () => {
    await getBasket(clientApiFetcher)
      .then((basket) => {
        setBasket(basket)
      })
      .catch((err) => {
        console.error(err)
        setBasket([])
      })
  }, [])

  return (
    <BasketContext.Provider
      value={{
        items: basket,
        addItem,
        setItems,
        removeItem,
        clearItems,
        refresh,
      }}
    >
      {children}
    </BasketContext.Provider>
  )
}
