'use client'

import McUsername from '@/components/ui/mc-username'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { updateProfileNamePrefix } from '@/lib/api-endpoints'
import { clientApiFetcher } from '@/lib/client'
import { errorToast } from '@/lib/toasts'
import { Profile, ProfileNameCosmeticOptions } from '@/models/profile'
import { CircleDivideIcon } from 'lucide-react'
import { useTranslations } from 'next-intl'
import Image from 'next/image'
import { useRouter } from 'next/navigation'
import React from 'react'
import { toast } from 'sonner'

type Props = {
  selectedProfile: Profile
  cosmeticOptions: ProfileNameCosmeticOptions
}

export default function NameGlythOptionSelect({
  selectedProfile,
  cosmeticOptions,
}: Props) {
  const [selectedNamePrefixId, setSelectedNamePrefixId] = React.useState(
    selectedProfile?.cosmetics.name?.glythPrefix?.id ?? 'none',
  )
  const [isPending, startTransition] = React.useTransition()
  const router = useRouter()
  const t = useTranslations('profiles.cosmetics.glyth')

  const handleSelectProfileNamePrefix = (namePrefixId: string) => {
    if (!selectedProfile || isPending) return

    const option = cosmeticOptions?.glythPrefixes.find(
      (prefix) => prefix.namePrefixId === namePrefixId,
    )
    if (!option && namePrefixId !== 'none') return

    const previousNamePrefixId = selectedNamePrefixId
    setSelectedNamePrefixId(namePrefixId)

    startTransition(async () => {
      try {
        await updateProfileNamePrefix(
          clientApiFetcher,
          selectedProfile.id,
          'glyth',
          {
            optionId: namePrefixId === 'none' ? undefined : option?.id,
          },
        )
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
        setSelectedNamePrefixId(previousNamePrefixId)
      }
    })
  }

  React.useEffect(() => {
    setSelectedNamePrefixId(
      selectedProfile.cosmetics.name?.glythPrefix?.id ?? 'none',
    )
  }, [selectedProfile.cosmetics.name?.glythPrefix?.id])

  return (
    <>
      <Select
        value={selectedNamePrefixId}
        onValueChange={handleSelectProfileNamePrefix}
      >
        <SelectTrigger className="w-full">
          <SelectValue placeholder={t('select')} />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="none">
            <CircleDivideIcon className="size-4" />
            <p className="text-semibold font-medium">{t('none')}</p>
          </SelectItem>
          {cosmeticOptions?.glythPrefixes?.map((prefix) => (
            <SelectItem key={prefix.id} value={prefix.namePrefixId}>
              <Image
                src={prefix.image}
                alt={prefix.name}
                width={20}
                height={20}
                unoptimized
              />
              <McUsername username={prefix.name} />
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    </>
  )
}
