'use client'

import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import ProfileSelect from '@/components/ui/profile-select'
import { Link } from '@/i18n/navigation'
import { ProductCategory } from '@/models/product'
import { CheckIcon, MessageCircleWarningIcon } from 'lucide-react'
import { useTranslations } from 'next-intl'
import React from 'react'
import AddToBasketButton from '../../components/add-to-basket-btn'
import { useProductProfile } from './product-profile-context'

type Props = React.ComponentProps<'div'> & {
  productId: string
  category: ProductCategory
}

export default function AddToBasketForm({
  productId,
  category,
  ...props
}: Props) {
  const t = useTranslations('products.page')
  const { profiles, selectedProfile, selectProfile, owned } =
    useProductProfile()

  // Access itself is on sale here, so missing it is no reason to warn.
  const needsAccess =
    !owned &&
    category !== ProductCategory.Upgrade &&
    selectedProfile !== undefined &&
    selectedProfile.accessStatus !== 'active'

  return (
    <div {...props}>
      <ProfileSelect
        profiles={profiles}
        placeholder={t('selectProfile')}
        selectedProfileId={selectedProfile?.id}
        onSelectProfileId={selectProfile}
        className="w-full"
      />
      {owned && selectedProfile && (
        <p className="text-muted-foreground mt-4 text-sm">
          {t.rich('owned.description', {
            username: selectedProfile.username,
            b: (chunks) => (
              <b className="text-foreground font-semibold">{chunks}</b>
            ),
          })}
        </p>
      )}
      {needsAccess && (
        <Alert className="mt-4 w-full">
          <AlertTitle className="flex items-center gap-2">
            <MessageCircleWarningIcon className="size-4" />
            {t('noAccess.title')}
          </AlertTitle>
          <AlertDescription>
            <p>
              {t.rich('noAccess.description', {
                username: selectedProfile.username,
                link: (chunks) => (
                  <Link
                    href={`/obtain-access?u=${selectedProfile.username}`}
                    className="font-bold text-blue-600 underline"
                  >
                    {chunks}
                  </Link>
                ),
                b: (chunks) => <b>{chunks}</b>,
              })}
            </p>
          </AlertDescription>
        </Alert>
      )}
      {owned ? (
        <Button size="lg" className="mt-4 w-full" disabled>
          <CheckIcon className="size-4" />
          {t('owned.button')}
        </Button>
      ) : (
        <AddToBasketButton
          productId={productId}
          profileId={selectedProfile?.id}
          size="lg"
          className="mt-4 w-full"
          // A guest has no profile to pick: the button sends them to log in.
          disabled={profiles.length > 0 && !selectedProfile}
        />
      )}
    </div>
  )
}
