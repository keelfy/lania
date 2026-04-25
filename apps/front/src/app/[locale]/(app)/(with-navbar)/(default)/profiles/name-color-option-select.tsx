'use client'

import McUsername from '@/components/ui/mc-username'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { updateProfileNameColor } from '@/lib/api-endpoints'
import { clientApiFetcher } from '@/lib/client'
import { errorToast } from '@/lib/toasts'
import { Profile, ProfileCosmeticOptions } from '@/models/profile'
import { useTranslations } from 'next-intl'
import { useRouter } from 'next/navigation'
import React from 'react'
import { toast } from 'sonner'

type Props = {
  selectedProfile: Profile
  cosmeticOptions: ProfileCosmeticOptions
}

export default function NameColorOptionSelect({
  selectedProfile,
  cosmeticOptions,
}: Props) {
  const [selectedNameColorId, setSelectedNameColorId] = React.useState(
    selectedProfile?.cosmetics.name.colors.id,
  )
  const [isPending, startTransition] = React.useTransition()
  const router = useRouter()
  const t = useTranslations('profiles.cosmetics.nameColor')

  const handleSelectProfileNameColor = (nameColorId: string) => {
    const option = cosmeticOptions?.name.colors.find(
      (color) => color.nameColorId === nameColorId,
    )
    if (!option || !selectedProfile || isPending) return

    const previousNameColorId = selectedNameColorId
    setSelectedNameColorId(nameColorId)

    startTransition(async () => {
      try {
        await updateProfileNameColor(clientApiFetcher, selectedProfile.id, {
          optionId: option.id,
        })
        toast.success(
          t.rich('changed', {
            username: () => <>{selectedProfile?.username}</>,
          }),
          {
            description: t('changedDescription'),
          },
        )
        router.refresh()
      } catch (error) {
        errorToast(
          t.rich('error', {
            username: () => <>{selectedProfile?.username}</>,
          }),
          error,
        )
        setSelectedNameColorId(previousNameColorId)
      }
    })
  }

  React.useEffect(() => {
    setSelectedNameColorId(selectedProfile.cosmetics.name.colors.id)
  }, [selectedProfile.cosmetics.name.colors.id])

  return (
    <>
      <Select
        value={selectedNameColorId}
        onValueChange={handleSelectProfileNameColor}
      >
        <SelectTrigger className="w-full">
          <SelectValue placeholder={t('select')} />
        </SelectTrigger>
        <SelectContent>
          {cosmeticOptions?.name.colors.map((color) => (
            <SelectItem key={color.id} value={color.nameColorId}>
              <McUsername username={color.name} colors={color.colors} />
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    </>
  )
}
