'use client'

import { Button } from '@/components/ui/button'
import { useBasket } from '@/context/basket'
import { addToBasket, getUserProfiles } from '@/lib/api-endpoints'
import { clientApiFetcher } from '@/lib/client'
import { errorToast } from '@/lib/toasts'
import { cn } from '@/lib/utils'
import { useAuthStore } from '@/providers/auth-store'
import { CheckIcon, Loader2Icon, ShoppingCartIcon } from 'lucide-react'
import { useTranslations } from 'next-intl'
import { usePathname, useRouter } from 'next/navigation'
import React from 'react'
import { toast } from 'sonner'

type Props = React.ComponentProps<typeof Button> & {
  productId: string
  profileId: string | undefined
  showText?: boolean
}

export default function AddToBasketButton({
  productId,
  profileId,
  showText = true,
  disabled,
  ...props
}: Props) {
  const [isProfilesLoading, startProfilesTransition] = React.useTransition()
  const [isAddingToBasket, startAddingToBasket] = React.useTransition()

  const t = useTranslations('basket')

  const session = useAuthStore((state) => state.session)
  const router = useRouter()
  const pathname = usePathname()
  const { addItem, removeItem, items } = useBasket()

  const addItemToBasket = React.useCallback(
    (productId: string, profileId: string) => {
      const mockItemId = addItem(productId, profileId)
      startAddingToBasket(async () => {
        try {
          await addToBasket(clientApiFetcher, productId, profileId)
        } catch (error) {
          removeItem(mockItemId)
          errorToast(t('failedToAddToBasket'), error)
        }
      })
    },
    [addItem, removeItem, startAddingToBasket],
  )

  const notInBasket = React.useMemo(
    () =>
      !items.some(
        (item) =>
          item.productId === productId &&
          (profileId ? item.profileId === profileId : true),
      ),
    [items, productId, profileId],
  )

  const Icon = React.useMemo(
    () => (notInBasket ? ShoppingCartIcon : CheckIcon),
    [notInBasket],
  )

  const onClick = () => {
    if (!notInBasket) {
      return
    }

    if (session?.active != true) {
      router.push(`/auth/login?goto=${pathname}`)
      return
    }

    startProfilesTransition(async () => {
      let profileIdToUse = profileId

      if (!profileIdToUse && session?.active == true) {
        await getUserProfiles(clientApiFetcher)
          .then((profiles) => {
            if (profiles.length > 0) {
              profileIdToUse = profiles[0].id
            }
          })
          .catch((error) => {
            console.error(error)
            profileIdToUse = undefined
          })
      }

      if (profileIdToUse) {
        addItemToBasket(productId, profileIdToUse)
      } else {
        toast.error(t('failedToAddToBasket'), {
          description: t('needToCreateProfile'),
        })
      }
    })
  }

  return (
    <Button onClick={onClick} disabled={!notInBasket || disabled} {...props}>
      {isProfilesLoading || isAddingToBasket ? (
        <Loader2Icon className="size-4 animate-spin" />
      ) : (
        <Icon className="size-4" />
      )}
      <span className={cn(!showText && 'sr-only')}>
        {notInBasket ? t('buy') : t('inTheBasket')}
      </span>
    </Button>
  )
}
