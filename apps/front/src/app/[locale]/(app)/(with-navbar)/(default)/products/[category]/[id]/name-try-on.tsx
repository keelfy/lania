'use client'

import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { usernameColorStyle } from '@/components/ui/mc-username'
import { cn } from '@/lib/utils'
import {
  NameColorProductMetadata,
  NamePrefixProductMetadata,
  Product,
  ProductCategory,
} from '@/models/product'
import { useTranslations } from 'next-intl'
import Image from 'next/image'
import React from 'react'
import InGamePreview, { NameLook } from './in-game-preview'
import { useProductProfile } from './product-profile-context'

// What else can be worn with the product: a name color with a glyph, a glyph with a name color.
type Pairing = { id: string; name: string; colors?: string[]; image?: string }

type Props =
  | {
      category: ProductCategory.NameColor
      item: Product<NameColorProductMetadata>
      pairings: Product<NamePrefixProductMetadata>[]
    }
  | {
      category: ProductCategory.NamePrefix
      item: Product<NamePrefixProductMetadata>
      pairings: Product<NameColorProductMetadata>[]
    }

const NONE = 'none'
const CURRENT = 'current'

export default function NameTryOn(props: Props) {
  const t = useTranslations('products.page.tryOn')
  const { selectedProfile } = useProductProfile()
  const wearsColor = props.category === ProductCategory.NameColor

  // What the player typed, kept for the profile it was typed with, so picking another profile shows its name.
  const [typed, setTyped] = React.useState<{
    profileId?: string
    value: string
  }>()
  const typedName =
    typed?.profileId === selectedProfile?.id ? typed?.value : undefined
  const username = typedName || selectedProfile?.username || 'Steve'

  // The profile's own glyph or color, offered first so the player sees the product with what they wear now.
  const current: Pairing | undefined = wearsColor
    ? selectedProfile?.cosmetics.name.glythPrefix && {
        id: CURRENT,
        name: selectedProfile.cosmetics.name.glythPrefix.name,
        image: selectedProfile.cosmetics.name.glythPrefix.image,
      }
    : selectedProfile?.cosmetics.name.colors && {
        id: CURRENT,
        name: selectedProfile.cosmetics.name.colors.name,
        colors: selectedProfile.cosmetics.name.colors.colors,
      }
  const pairings: Pairing[] = [
    ...(current ? [current] : []),
    ...(props.category === ProductCategory.NameColor
      ? props.pairings.map((product) => ({
          id: product.id,
          name: product.name,
          image: product.metadata.prefix,
        }))
      : props.pairings.map((product) => ({
          id: product.id,
          name: product.name,
          colors: product.metadata.colors,
        }))),
  ]

  const [pairingId, setPairingId] = React.useState<string>()
  // Until the player picks, the product is shown with what the profile wears now.
  const pairing =
    pairingId === NONE
      ? undefined
      : (pairings.find((option) => option.id === (pairingId ?? CURRENT)) ??
        undefined)

  const look: NameLook =
    props.category === ProductCategory.NameColor
      ? {
          username,
          colors: props.item.metadata.colors,
          prefixImage: pairing?.image,
        }
      : {
          username,
          colors: pairing?.colors,
          prefixImage: props.item.metadata.prefix,
        }

  return (
    <section className="flex flex-col gap-4">
      <h2 className="text-2xl font-semibold tracking-tight">{t('title')}</h2>
      <InGamePreview look={look} profile={selectedProfile} />
      <div className="grid gap-4 sm:grid-cols-[minmax(0,14rem)_1fr]">
        <div className="flex flex-col gap-2">
          <Label htmlFor="try-on-username">{t('username')}</Label>
          <Input
            id="try-on-username"
            placeholder={selectedProfile?.username ?? 'Steve'}
            value={typedName ?? ''}
            maxLength={16}
            onChange={(event) =>
              setTyped({
                profileId: selectedProfile?.id,
                value: event.target.value,
              })
            }
          />
        </div>
        {pairings.length > 0 && (
          <fieldset className="flex min-w-0 flex-col gap-2">
            <legend className="mb-2 text-sm leading-none font-medium">
              {wearsColor ? t('withGlyph') : t('withColor')}
            </legend>
            <div className="flex flex-wrap gap-1.5">
              <PairingButton
                selected={pairing === undefined}
                onClick={() => setPairingId(NONE)}
              >
                {t('none')}
              </PairingButton>
              {pairings.map((option) => (
                <PairingButton
                  key={option.id}
                  selected={pairing?.id === option.id}
                  onClick={() => setPairingId(option.id)}
                  title={option.name}
                >
                  <PairingSwatch pairing={option} />
                  {option.id === CURRENT ? (
                    t('current')
                  ) : (
                    <span className="sr-only">{option.name}</span>
                  )}
                </PairingButton>
              ))}
            </div>
          </fieldset>
        )}
      </div>
    </section>
  )
}

function PairingButton({
  selected,
  className,
  ...props
}: React.ComponentProps<'button'> & { selected: boolean }) {
  return (
    <button
      type="button"
      aria-pressed={selected}
      className={cn(
        'bg-card hover:bg-accent focus-visible:ring-ring flex h-9 min-w-9 items-center justify-center gap-1.5 rounded-md border px-2 text-sm transition-colors outline-none focus-visible:ring-2',
        selected && 'border-primary ring-primary ring-1',
        className,
      )}
      {...props}
    />
  )
}

function PairingSwatch({ pairing }: { pairing: Pairing }) {
  if (pairing.image) {
    return (
      <Image
        src={pairing.image}
        alt=""
        width={32}
        height={32}
        className="size-5 [image-rendering:pixelated]"
      />
    )
  }
  const { style } = usernameColorStyle(pairing.colors)
  return (
    <span
      className="size-5 rounded-sm border border-white/20"
      style={{
        background: style.backgroundImage ?? style.color,
      }}
    />
  )
}
