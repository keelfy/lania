import DeerIcon from '@/components/icons/DeerIcon'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import {
  SiDiscord,
  SiDiscordHex,
  SiTelegram,
  SiTelegramHex,
} from '@icons-pack/react-simple-icons'
import {
  BanknoteXIcon,
  CreditCardIcon,
  FileIcon,
  HeadsetIcon,
  HeartIcon,
  ShieldIcon,
  UserIcon,
} from 'lucide-react'
import { getTranslations } from 'next-intl/server'
import Link from 'next/link'

export const footerSocialLinks = [
  {
    label: 'telegram',
    icon: SiTelegram,
    color: SiTelegramHex,
    href: process.env.NEXT_PUBLIC_TELEGRAM_CHANNEL_URL ?? '#',
  },
  {
    label: 'discord',
    icon: SiDiscord,
    color: SiDiscordHex,
    href: process.env.NEXT_PUBLIC_DISCORD_SERVER_URL ?? '#',
  },
  {
    label: 'support',
    icon: HeadsetIcon,
    href: process.env.NEXT_PUBLIC_SUPPORT_URL ?? '#',
  },
]

export const footerLegalLinkd = [
  {
    label: 'contacts',
    icon: HeadsetIcon,
    href: '/contact',
  },
  {
    label: 'oferta',
    icon: FileIcon,
    href: '/wiki/legal/oferta',
  },
  {
    label: 'tos',
    icon: UserIcon,
    href: '/wiki/legal/tos',
  },
  {
    label: 'privacy',
    icon: ShieldIcon,
    href: '/wiki/legal/privacy',
  },
  {
    label: 'payments',
    icon: CreditCardIcon,
    href: '/wiki/legal/payments',
  },
  {
    label: 'refund',
    icon: BanknoteXIcon,
    href: '/wiki/legal/refund',
  },
]

type Props = {
  locale: string
  className?: string
}

export default async function Footer({ locale, className }: Props) {
  const t = await getTranslations({ locale, namespace: 'footer' })
  return (
    <footer
      className={cn(
        'border-primary-foreground flex flex-0 items-center border-t backdrop-blur',
        className,
      )}
    >
      <div className="container mx-auto px-4 py-12">
        <div className="flex flex-col items-start gap-4">
          <div className="flex w-full items-center justify-between gap-2">
            <div className="flex items-center gap-2 px-3">
              <DeerIcon className="size-6" />
              <p className="text-md font-mono font-semibold">
                LANIA.GG © {new Date().getFullYear()}
              </p>
            </div>
            <div className="flex items-center gap-2">
              {footerSocialLinks.map((link) => (
                <Button variant="ghost" size="icon" asChild key={link.href}>
                  <Link href={link.href}>
                    <link.icon className="size-5" color={link.color} />
                  </Link>
                </Button>
              ))}
            </div>
          </div>
          <div className="flex flex-wrap items-center gap-2">
            {footerLegalLinkd.map((link) => (
              <Button
                asChild
                variant="ghost"
                key={link.href}
                className="justify-self-start"
              >
                <Link href={link.href} className="flex items-center gap-2">
                  <link.icon className="size-4" />
                  <p>{t(`legalLinks.${link.label}`)}</p>
                </Link>
              </Button>
            ))}
          </div>
          <div className="flex w-full flex-col items-center justify-between gap-4 sm:flex-row">
            <p className="text-muted-foreground px-3 text-sm">
              Not an official Minecraft product. We are in no way affiliated
              with or endorsed by Mojang Synergies AB, Microsoft Corporation or
              other rightsholders.
            </p>
            <p className="text-muted-foreground self-end text-sm">
              {t('madeBy.text')}&nbsp;
              <a
                href="https://twitch.tv/keelfy"
                target="_blank"
                rel="noopener noreferrer"
                className="hover:underline"
              >
                {t('madeBy.name')}
              </a>
              &nbsp;
              <HeartIcon className="inline-block size-3" />
            </p>
          </div>
        </div>
      </div>
    </footer>
  )
}
