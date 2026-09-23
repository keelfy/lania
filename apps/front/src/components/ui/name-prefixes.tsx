import { NameCosmetics } from '@/models/profile'
import Image from 'next/image'

type Props = {
  cosmetics?: NameCosmetics
  size?: number
}

// Renders the glyth and special prefixes of a profile the way they look in the game chat.
export default function NamePrefixes({ cosmetics, size = 20 }: Props) {
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
        />
      ))}
    </>
  )
}
