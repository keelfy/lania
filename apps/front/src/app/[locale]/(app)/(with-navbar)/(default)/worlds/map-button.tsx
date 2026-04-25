'use client'

import { useRouter } from 'next/navigation'

type Props = React.ComponentProps<'button'> & {
  mapId: string
}

export default function MapButton({ mapId, ...props }: Props) {
  const router = useRouter()
  return (
    <button
      onClick={() => {
        router.push(`/worlds/${mapId}`)
      }}
      {...props}
    />
  )
}
