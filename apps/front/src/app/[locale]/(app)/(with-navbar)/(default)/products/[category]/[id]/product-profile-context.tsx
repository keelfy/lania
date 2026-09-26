'use client'

import { getPurchases, getUserProfiles } from '@/lib/api-endpoints'
import { clientApiFetcher } from '@/lib/client'
import { Profile } from '@/models/profile'
import { useAuthStore } from '@/providers/auth-store'
import React from 'react'

type ProductProfileContextValue = {
  profiles: Profile[]
  selectedProfile: Profile | undefined
  selectProfile: (id: string) => void
  // The selected profile already has the product in the primary season.
  owned: boolean
}

const ProductProfileContext = React.createContext<
  ProductProfileContextValue | undefined
>(undefined)

type Props = React.PropsWithChildren<{
  productId: string
}>

// Shares the profile the product is bought for between the buy form and the try-on preview,
// so the preview wears the player's own nickname.
export default function ProductProfileProvider({ productId, children }: Props) {
  const isLoggedIn = useAuthStore((state) => state.session?.active === true)
  const [profiles, setProfiles] = React.useState<Profile[]>([])
  const [ownerIds, setOwnerIds] = React.useState<string[]>([])
  const [selectedId, setSelectedId] = React.useState<string>()

  React.useEffect(() => {
    if (!isLoggedIn) return
    getUserProfiles(clientApiFetcher).then(setProfiles).catch(console.error)
    getPurchases(clientApiFetcher, [productId])
      .then((purchases) =>
        setOwnerIds(purchases.map((purchase) => purchase.profileId)),
      )
      .catch(console.error)
  }, [isLoggedIn, productId])

  // The first profile that can still buy the product is picked until the player picks one.
  const selectedProfile =
    profiles.find((profile) => profile.id === selectedId) ??
    profiles.find((profile) => !ownerIds.includes(profile.id)) ??
    profiles[0]

  const value = React.useMemo(
    () => ({
      profiles,
      selectedProfile,
      selectProfile: setSelectedId,
      owned: !!selectedProfile && ownerIds.includes(selectedProfile.id),
    }),
    [profiles, selectedProfile, ownerIds],
  )

  return (
    <ProductProfileContext.Provider value={value}>
      {children}
    </ProductProfileContext.Provider>
  )
}

export function useProductProfile() {
  const value = React.useContext(ProductProfileContext)
  if (!value) {
    throw new Error(
      'useProductProfile must be used within ProductProfileProvider',
    )
  }
  return value
}
