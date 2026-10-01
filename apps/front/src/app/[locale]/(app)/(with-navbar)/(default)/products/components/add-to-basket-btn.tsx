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
import styles from './purchase-feedback.module.css'
import { useCatalogProfile } from './catalog-try-on'

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
  const catalog = useCatalogProfile()
  const recipientId = profileId ?? catalog?.profileId
  const recipientUnavailable =
    profileId === undefined && !!catalog && (catalog.loading || catalog.failed)
  const [isProfilesLoading, startProfilesTransition] = React.useTransition()
  const [isAddingToBasket, startAddingToBasket] = React.useTransition()

  const t = useTranslations('basket')

  const session = useAuthStore((state) => state.session)
  const router = useRouter()
  const pathname = usePathname()
  const { addItem, removeItem, refresh, items } = useBasket()

  const addItemToBasket = React.useCallback(
    (productId: string, profileId: string) => {
      const mockItemId = addItem(productId, profileId)
      startAddingToBasket(async () => {
        try {
          await addToBasket(clientApiFetcher, productId, profileId)
          // The optimistic item does not know its season yet.
          await refresh()
        } catch (error) {
          removeItem(mockItemId)
          errorToast(t('failedToAddToBasket'), error)
        }
      })
    },
    [addItem, removeItem, refresh, startAddingToBasket, t],
  )

  const notInBasket = !items.some(
    (item) =>
      item.productId === productId &&
      (recipientId ? item.profileId === recipientId : true),
  )
  const pending = isProfilesLoading || isAddingToBasket

  const onClick = () => {
    if (!notInBasket || recipientUnavailable) {
      return
    }

    if (session?.active != true) {
      router.push(`/auth/login?goto=${pathname}`)
      return
    }

    startProfilesTransition(async () => {
      let profileIdToUse = recipientId

      if (!profileIdToUse && !catalog && session?.active == true) {
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
    <Button
      {...props}
      onClick={onClick}
      aria-busy={pending}
      title={notInBasket ? t('buy') : t('inTheBasket')}
      disabled={recipientUnavailable || pending || !notInBasket || disabled}
    >
      {pending ? (
        <Loader2Icon
          aria-hidden="true"
          className="size-4 animate-spin motion-reduce:animate-none"
        />
      ) : notInBasket ? (
        <ShoppingCartIcon aria-hidden="true" className="size-4" />
      ) : (
        <span className={styles.confirmed}>
          <CheckIcon aria-hidden="true" className="size-4" />
        </span>
      )}
      <span
        aria-live="polite"
        aria-atomic="true"
        className={cn(!showText && 'sr-only')}
      >
        {notInBasket ? t('buy') : t('inTheBasket')}
      </span>
    </Button>
  )
}
