import { Link } from '@/i18n/navigation'
import { cn } from '@/lib/utils'
import { Product, UpgradeProductMetadata } from '@/models/product'
import { UserCheckIcon } from 'lucide-react'
import { useTranslations } from 'next-intl'

type Props = React.ComponentProps<'div'> & {
  item: Product<UpgradeProductMetadata>
}

const ICONS = {
  season_access: UserCheckIcon,
}

const LEGAL_LINKS = {
  rules: '/wiki/rules',
  tos: '/wiki/legal/tos',
  privacy: '/wiki/legal/privacy',
  refund: '/wiki/legal/refund',
  payments: '/wiki/legal/payments',
}

function SeasonAccessDescription() {
  const t = useTranslations('products.page.seasonAccess')

  return (
    <>
      <p>{t('description')}</p>
      <p className="mt-2">{t('limit')}</p>
      <p className="text-muted-foreground mt-4">
        {t.rich('agreement', {
          ...Object.fromEntries(
            Object.entries(LEGAL_LINKS).map(([tag, href]) => [
              tag,
              (chunks: React.ReactNode) => (
                <Link
                  href={href}
                  className="font-medium underline underline-offset-2"
                >
                  {chunks}
                </Link>
              ),
            ]),
          ),
        })}
      </p>
    </>
  )
}

const DESCRIPTIONS = {
  season_access: SeasonAccessDescription,
}

export default function UpgradeProductDetails({
  item,
  className,
  ...props
}: Props) {
  const action = item.metadata.action as keyof typeof ICONS
  const Icon = ICONS[action]
  const Description = DESCRIPTIONS[action]

  return (
    <div className={cn('flex flex-col', className)} {...props}>
      <div className="flex items-center gap-3 text-4xl leading-tight sm:text-5xl lg:text-6xl">
        {Icon && <Icon className="size-[1em] shrink-0 stroke-2" />}
        <p className="scroll-m-20 font-extrabold text-balance">{item.name}</p>
      </div>
      <div className="mt-4 max-w-prose">{Description && <Description />}</div>
    </div>
  )
}
