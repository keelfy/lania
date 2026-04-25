import { getTranslations } from 'next-intl/server'
import { redirect } from 'next/navigation'

type Props = {
  params: Promise<{
    world: string
    locale: string
  }>
}

const allowedWorlds = ['survival', 'farms']

export default async function MapsPage({ params }: Props) {
  const { world, locale } = await params
  const t = await getTranslations({ locale, namespace: 'worlds.maps' })

  if (!allowedWorlds.includes(world)) {
    return redirect(`/worlds`)
  }

  return (
    <iframe
      title={t('title').replace('<world/>', t('mapNames.' + world))}
      src={`https://maps.lania.gg/${world}/1/`}
      style={{ border: 'none', width: '100%', height: '100%' }}
      allowFullScreen
    />
  )
}
