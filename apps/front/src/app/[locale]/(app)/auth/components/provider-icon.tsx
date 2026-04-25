import React from 'react'
import {
  SiDiscord,
  SiDiscordHex,
  SiGoogle,
  SiGoogleHex,
  SiTwitch,
  SiTwitchHex,
} from '@icons-pack/react-simple-icons'
import YandexIcon from '@/components/icons/YandexIcon'

type Props = React.ComponentProps<'svg'> & {
  providerId?: string
  className?: string
}

const ProviderIcons = {
  google: SiGoogle,
  twitch: SiTwitch,
  discord: SiDiscord,
  yandex: YandexIcon,
}

const ProviderIconHex = {
  google: SiGoogleHex,
  twitch: SiTwitchHex,
  discord: SiDiscordHex,
  yandex: undefined,
}

export default function ProviderIcon({ providerId, ...props }: Props) {
  if (!providerId) return undefined
  const idParts = providerId.split('-')
  const id = idParts[0].toLowerCase() as keyof typeof ProviderIcons
  const Icon = ProviderIcons[id]
  const IconHex = ProviderIconHex[id]
  if (!Icon) return undefined
  return <Icon color={IconHex} {...props} />
}
