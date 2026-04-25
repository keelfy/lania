'use client'

import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import ProfileSelect from '@/components/ui/profile-select'
import { getUserProfiles } from '@/lib/api-endpoints'
import { Profile } from '@/models/profile'
import { MessageCircleWarningIcon } from 'lucide-react'
import Link from 'next/link'
import React from 'react'
import AddToBasketButton from '../../components/add-to-basket-btn'
import { clientApiFetcher } from '@/lib/client'

type Props = React.ComponentProps<'div'> & {
  productId: string
}

export default function AddToBasketForm({ productId, ...props }: Props) {
  const [profileId, setProfileId] = React.useState<string | undefined>()
  const [profiles, setProfiles] = React.useState<Profile[]>([])

  const selectedProfile = React.useMemo(
    () => profiles.find((profile) => profile.id === profileId),
    [profiles, profileId],
  )

  React.useEffect(() => {
    getUserProfiles(clientApiFetcher).then(setProfiles)
  }, [])

  return (
    <div {...props}>
      <ProfileSelect
        profiles={profiles}
        placeholder="Выберите игровой профиль"
        selectedProfileId={profileId}
        onSelectProfileId={setProfileId}
        className="w-full"
      />
      {selectedProfile && selectedProfile?.accessStatus !== 'active' && (
        <Alert className="mt-4 w-full">
          <AlertTitle className="flex items-center gap-2">
            <MessageCircleWarningIcon className="size-4" />
            Нет проходки
          </AlertTitle>
          <AlertDescription>
            <p>
              Вы можете купить этот товар, но не сможете использовать его, пока
              не приобритете&nbsp;
              <Link
                href={`/obtain-access?u=${selectedProfile?.username}`}
                className="font-bold text-blue-600 underline"
              >
                проходку для&nbsp;
                <span className="font-bold">{selectedProfile?.username}</span>
              </Link>
              &nbsp;на текущий сезон.
            </p>
          </AlertDescription>
        </Alert>
      )}
      <AddToBasketButton
        productId={productId}
        profileId={profileId}
        size="lg"
        className="mt-4 w-full"
        disabled={!profileId}
      />
    </div>
  )
}
