import { NameCosmetics } from '@/models/profile'
import Image from 'next/image'

type Props = {
  cosmetics?: NameCosmetics
  size?: number
  // The gap the parent puts between its children, taken back for a prefix without a space.
  gap?: string
}

// Renders the glyth and special prefixes of a profile the way they look in the game chat.
export default function NamePrefixes({
  cosmetics,
  size = 20,
  gap = '0.375rem',
}: Props) {
  const prefixes = [cosmetics?.glythPrefix, cosmetics?.specialPrefix].filter(
    (prefix) => prefix?.image,
  )
  if (prefixes.length === 0) return null

  return (
    <>
      {prefixes.map((prefix) => (
        <Image
          key={prefix!.id}
          src={prefix!.image}
          alt={prefix!.name}
          title={prefix!.name}
          width={size}
          height={size}
          style={
            prefix!.noSpace ? { marginRight: `calc(${gap} * -1)` } : undefined
          }
        />
      ))}
    </>
  )
}
